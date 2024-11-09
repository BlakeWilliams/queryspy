package mysql

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"

	"vitess.io/vitess/go/vt/sqlparser"
)

var parser *sqlparser.Parser

var once sync.Once

// Query exposes data about a SQL query like the table, and the redacted
// query.
type Query struct {
	// Raw is the unmodified query provided to MySQL
	Raw string
	// Redacted replaces all variables with ?
	Redacted string
	// Comments returns the comments stripped from the beginning and end of the query
	Comments sqlparser.MarginComments
	// Table is the name of the table being queried, if applicable
	Table string
}

var normalizeRegexp = regexp.MustCompile(`\:+[a-zA-Z0-9]+( \/\*.*?\*\/)?`)

func NewQuery(rawQuery string) (*Query, error) {
	once.Do(func() {
		var err error
		parser, err = sqlparser.New(sqlparser.Options{})
		if err != nil {
			panic(err)
		}
	})

	stmt, err := parser.Parse(rawQuery)
	if err != nil {
		return nil, err
	}

	orderedQuery, err := parser.NormalizeAlphabetically(rawQuery)
	if err != nil {
		return nil, err
	}

	fullRedacted, err := parser.RedactSQLQuery(orderedQuery)
	if err != nil {
		return nil, err
	}
	redacted, comments := sqlparser.SplitMarginComments(fullRedacted)
	redacted = normalizeRegexp.ReplaceAllString(redacted, "?")

	return &Query{
		Raw:      rawQuery,
		Redacted: redacted,
		Comments: comments,
		Table:    tableName(stmt),
	}, nil
}

func tableName(stmt sqlparser.Statement) string {
	var tableName string

	sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
		switch n := node.(type) {
		case *sqlparser.Select:
			for _, table := range n.From {
				if tableExpr, ok := table.(sqlparser.SimpleTableExpr); ok {
					tableName = sqlparser.String(tableExpr)
					return false, nil
				}

				if tableExpr, ok := table.(*sqlparser.AliasedTableExpr); ok {
					tableName = sqlparser.String(tableExpr)
					return false, nil
				}
			}
		}

		return true, nil
	}, stmt)

	return strings.ToLower(tableName)
}

func (q *Query) Fingerprint() string {
	hash := md5.New()
	hash.Write([]byte(q.Redacted))

	res := hash.Sum(nil)

	return hex.EncodeToString(res)
}
