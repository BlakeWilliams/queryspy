package mysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLenEncString(t *testing.T) {
	testCases := []struct {
		desc           string
		input          string
		expectedLength uint64
	}{
		{
			desc:           "1 byte message",
			input:          "hello, world",
			expectedLength: 12,
		},
		{
			desc:           "2 byte message",
			input:          strings.Repeat("x", 255),
			expectedLength: 255,
		},
		{
			desc:           "3 byte message",
			input:          strings.Repeat("x", (1<<16)+2),
			expectedLength: (1 << 16) + 2,
		},
		{
			desc:           "4 byte message",
			input:          strings.Repeat("x", (1<<24)+3),
			expectedLength: (1 << 24) + 3,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res := LenEncString(tC.input)
			fmt.Println(LenEnc(res))

			require.Equal(t, tC.expectedLength, LenEnc(res))
		})
	}
}
