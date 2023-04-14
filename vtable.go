package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/nbd-wtf/go-nostr"
)

type nostrModule struct {
}

func (m *nostrModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
          id text NOT NULL,
          pubkey text NOT NULL,
          created_at integer NOT NULL,
          kind integer NOT NULL,
          tags jsonb NOT NULL,
          content text NOT NULL,
          sig text NOT NULL
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &nostrTimelineTable{}, nil
}

func (m *nostrModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *nostrModule) DestroyModule() {}

type nostrTimelineTable struct {
	events []nostr.Event
}

func (v *nostrTimelineTable) Open() (sqlite3.VTabCursor, error) {
	relay, err := nostr.RelayConnect(context.Background(), "wss://nostr.wine")
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds: []int{nostr.KindTextNote},
	}
	events, err := relay.QuerySync(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	return &nostrTimelineCursor{0, events}, nil
}

func (v *nostrTimelineTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{}, nil
}

func (v *nostrTimelineTable) Disconnect() error { return nil }
func (v *nostrTimelineTable) Destroy() error    { return nil }

type nostrTimelineCursor struct {
	index  int
	events []*nostr.Event
}

func (vc *nostrTimelineCursor) Column(c *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		c.ResultText(vc.events[vc.index].ID)
	case 1:
		c.ResultText(vc.events[vc.index].PubKey)
	case 2:
		c.ResultInt(int(vc.events[vc.index].CreatedAt.Unix()))
	case 3:
		c.ResultInt(vc.events[vc.index].Kind)
	case 4:
		b, _ := json.Marshal(vc.events[vc.index].Tags)
		c.ResultText(string(b))
	case 5:
		c.ResultText(vc.events[vc.index].Content)
	case 6:
		c.ResultText(vc.events[vc.index].Sig)
	}
	return nil
}

func (vc *nostrTimelineCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	vc.index = 0
	return nil
}

func (vc *nostrTimelineCursor) Next() error {
	vc.index++
	return nil
}

func (vc *nostrTimelineCursor) EOF() bool {
	return vc.index >= len(vc.events)
}

func (vc *nostrTimelineCursor) Rowid() (int64, error) {
	return int64(vc.index), nil
}

func (vc *nostrTimelineCursor) Close() error {
	return nil
}
