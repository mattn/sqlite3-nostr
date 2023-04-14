package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mattn/go-sqlite3"
)

func main() {
	sql.Register("sqlite3_with_extensions", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.CreateModule("nostr", &nostrModule{})
		},
	})
	db, err := sql.Open("sqlite3_with_extensions", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create virtual table events using nostr(id, pubkey, created_at, kind, tags, content, sig)")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		rows, err := db.Query(text)
		if err != nil {
			fmt.Println(err)
			continue
		}

		ct, err := rows.ColumnTypes()
		if err != nil {
			log.Fatal(err)
		}
		types := make([]reflect.Type, len(ct))
		for i, typ := range ct {
			types[i] = typ.ScanType()
		}

		// Column values
		values := make([]any, len(ct))
		for i := range values {
			values[i] = reflect.New(types[i]).Interface()
		}
		for rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				log.Fatal(err)
			}
			for i, v := range values {
				if i != 0 {
					fmt.Print(" ")
				}
				switch t := v.(type) {
				case *sql.NullInt64:
					fmt.Print(t.Int64)
				case *sql.NullString:
					fmt.Print(t.String)
				case **interface{}:
					if *t == nil {
						fmt.Print("NULL")
					} else {
						fmt.Print(**t)
					}
				default:
					fmt.Printf("%#T", t)
				}
			}
			fmt.Println()
		}
		rows.Close()
	}
}
