package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	testCases := []struct {
		desc     string
		in       string
		expected string
	}{
		{
			desc:     "basic",
			in:       "sElecT * from foo WhErE id = 1 OR id = 2 OR iD in (1,2,3)",
			expected: "select * from foo where id = ? or id = ? or iD in ?",
		},
		{
			desc:     "mixed case is normalized",
			in:       "sElecT * from foo WhErE name = 'foo'",
			expected: "select * from foo where `name` = ?",
		},
		{
			desc:     "comments are stripped",
			in:       "/* hello world */ sElecT * from foo WhErE name = /* great stuff */ 'foo' /* omg what */",
			expected: "select * from foo where `name` = ?",
		},
		{
			desc:     "excess whitespace is stripped",
			in:       "			 SELECT * from foo  WhErE name = 'foo'  ",
			expected: "select * from foo where `name` = ?",
		},
		{
			desc:     "normalizes 'and' order",
			in:       "select * from foo where second = 2 AND first = ?",
			expected: "select * from foo where `first` = ? and `second` = ?",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			q, err := NewQuery(tC.in)
			require.NoError(t, err)

			require.Equal(t, tC.expected, q.Redacted)
		})
	}
}
