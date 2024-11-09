package mysql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLenEncString(t *testing.T) {
	testCases := []struct {
		desc                 string
		input                string
		expectedLength       uint64
		expectedHeaderLength uint64
	}{
		{
			desc:                 "1 byte message",
			input:                "hello, world",
			expectedLength:       12,
			expectedHeaderLength: 1,
		},
		{
			desc:                 "2 byte message",
			input:                strings.Repeat("x", 255),
			expectedLength:       255,
			expectedHeaderLength: 3,
		},
		{
			desc:                 "3 byte message",
			input:                strings.Repeat("x", (1<<16)+2),
			expectedLength:       (1 << 16) + 2,
			expectedHeaderLength: 4,
		},
		{
			desc:                 "4 byte message",
			input:                strings.Repeat("x", (1<<24)+3),
			expectedLength:       (1 << 24) + 3,
			expectedHeaderLength: 9,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			payload := make([]byte, 0, tC.expectedLength+1)
			res := LenEncString(payload, tC.input)

			require.Equal(t, tC.expectedLength, LenEnc(res))

			// Double check header length
			sizeHeaderLen := uint64(len(res)) - LenEnc(res)
			require.Equal(t, tC.expectedHeaderLength, sizeHeaderLen)

			decodedStr := ReadLenEncString(res)
			require.Equal(t, tC.input, string(decodedStr))
		})
	}
}
