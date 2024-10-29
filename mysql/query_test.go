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
			desc:     "",
			in:       "sElecT * from foo WhErE id = 1 OR id = 2 OR iD in (1,2,3)",
			expected: "select * from foo where id = ? or id = ? or iD in ?",
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
