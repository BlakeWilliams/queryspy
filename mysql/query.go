package mysql

import (
	"regexp"
	"sync"

	"vitess.io/vitess/go/vt/sqlparser"
)

var parser *sqlparser.Parser
var once sync.Once

type Query struct {
	Raw      string
	Redacted string
	Comments sqlparser.MarginComments
	stmt     sqlparser.Statement
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

	return &Query{Raw: rawQuery, Redacted: redacted, Comments: comments}, nil
}
