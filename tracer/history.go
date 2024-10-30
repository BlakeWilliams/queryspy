package tracer

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/blakewilliams/guesswho/mysql"
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
		Name    string
		Queries map[string]*uint32
		mu      sync.RWMutex
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
				l.Logger.Error("could not parse query", "err", err, "query")
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
		Name:    tableName,
		Queries: make(map[string]*uint32),
	}
	l.Tables[tableName] = table
	l.mu.Unlock()

	return table
}

func (t *Table) Store(query *mysql.Query) uint32 {
	t.mu.RLock()
	val, exists := t.Queries[query.Redacted]
	t.mu.RUnlock()

	if exists {
		return atomic.AddUint32(val, 1)
	}

	t.mu.Lock()
	var i uint32 = 1
	t.Queries[query.Redacted] = &i
	t.mu.Unlock()
	return 1
}
