package tracer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/blakewilliams/guesswho/mysql"
	"github.com/goccy/go-yaml"
)

type (
	// History tracks queries and their tables
	History struct {
		Tables  map[string]*Table
		Queries chan string
		Logger  *slog.Logger
		mu      sync.RWMutex
	}

	Table struct {
		Name      string
		QueryData map[string]queryMeta
		mu        sync.RWMutex
	}

	queryMeta struct {
		query *mysql.Query
		count *uint32
	}
)

func (l *History) Process(ctx context.Context) {
	if l.Logger == nil {
		l.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	if l.Queries == nil {
		l.Queries = make(chan string, 50)
	}

	if l.Tables == nil {
		l.Tables = make(map[string]*Table)
	}

	for {
		select {
		case rawQuery := <-l.Queries:
			query, err := mysql.NewQuery(rawQuery)
			if err != nil {
				l.Logger.Error("could not parse query", "err", err, "query", query.Redacted)
				continue
			}

			if query.Table == "" {
				l.Logger.Error("could not get table", "err", err, "query", query.Redacted)
				continue
			}

			table := l.Table(query.Table)
			count := table.Store(query)

			l.Logger.Info("query", "table", query.Table, "count", count, "query", query.Redacted)
		case <-ctx.Done():
			break
		}
	}
}

func (l *History) Table(tableName string) *Table {
	l.mu.RLock()
	table, exists := l.Tables[tableName]
	l.mu.RUnlock()

	if exists {
		return table
	}

	l.mu.Lock()
	table = &Table{
		Name:      tableName,
		QueryData: make(map[string]queryMeta),
	}
	l.Tables[tableName] = table
	l.mu.Unlock()

	return table
}

func (t *Table) Store(query *mysql.Query) uint32 {
	t.mu.RLock()
	val, exists := t.QueryData[query.Fingerprint()]
	t.mu.RUnlock()

	if exists {
		return atomic.AddUint32(val.count, 1)
	}

	t.mu.Lock()
	var i uint32 = 1
	t.QueryData[query.Fingerprint()] = queryMeta{query: query, count: &i}
	t.mu.Unlock()
	return 1
}

// TODO this probably needs another mutex
func (h *History) Dump() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, table := range h.Tables {
		err := table.Dump()
		if err != nil {
			return fmt.Errorf("could not dump table %s: %w", table.Name, err)
		}
	}

	return nil
}

func (t *Table) Dump() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.QueryData) == 0 {
		return nil
	}

	os.MkdirAll("out", 0755)
	f, err := os.OpenFile(path.Join("out", t.Name+".yaml"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	res, err := yaml.Marshal(t)
	if err != nil {
		panic(err)
	}

	f.Write([]byte("# This file was auto-generated. DO NOT MODIFY\n"))
	f.Write(res)

	return nil
}

func (t *Table) MarshalYAML() (any, error) {
	type (
		queryRow struct {
			Digest string
		}
		fileFormat struct {
			Version int
			Queries yaml.MapSlice
		}
	)

	order := make([]string, 0, len(t.QueryData))
	for query, _ := range t.QueryData {
		order = append(order, query)
	}

	sort.SliceStable(order, func(i, j int) bool {
		return order[i] < order[j]
	})

	root := &fileFormat{
		Version: 1.0,
		Queries: make([]yaml.MapItem, 0, len(t.QueryData)),
	}

	for _, fingerprint := range order {
		root.Queries = append(root.Queries, yaml.MapItem{
			Key: fingerprint,
			Value: queryRow{
				Digest: t.QueryData[fingerprint].query.Redacted,
			},
		})
	}

	return root, nil
}
