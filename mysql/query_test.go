package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	testCases := []struct {
		desc              string
		in                string
		expected          string
		expectedTableName string
	}{
		{
			desc:              "basic",
			in:                "sElecT * from foo WhErE id = 1 OR id = 2 OR iD in (1,2,3)",
			expected:          "select * from foo where id = ? or id = ? or iD in ?",
			expectedTableName: "foo",
		},
		{
			desc:              "mixed case is normalized",
			in:                "sElecT * from foo WhErE name = 'foo'",
			expected:          "select * from foo where `name` = ?",
			expectedTableName: "foo",
		},
		{
			desc:              "comments are stripped",
			in:                "/* hello world */ sElecT * from foo WhErE name = /* great stuff */ 'foo' /* omg what */",
			expected:          "select * from foo where `name` = ?",
			expectedTableName: "foo",
		},
		{
			desc:              "excess whitespace is stripped",
			in:                "			 SELECT * from foo  WhErE name = 'foo'  ",
			expected:          "select * from foo where `name` = ?",
			expectedTableName: "foo",
		},
		{
			desc:              "normalizes 'and' order",
			in:                "select * from foo where second = 2 AND first = ?",
			expected:          "select * from foo where `first` = ? and `second` = ?",
			expectedTableName: "foo",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			q, err := NewQuery(tC.in)
			require.NoError(t, err)

			require.Equal(t, tC.expected, q.Redacted)
			require.Equal(t, tC.expectedTableName, q.Table)
		})
	}
}

func TestQuery_TableName(t *testing.T) {
	query, err := NewQuery("select * from FoO where id = 1")
	require.NoError(t, err)

	require.Equal(t, "foo", query.Table)
}
