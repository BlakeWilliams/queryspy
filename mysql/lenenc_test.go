package mysql

import (
	"bytes"
	"io"
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
			var buf bytes.Buffer
			LenEncString(&buf, tC.input)
			// payload := buf.Bytes()

			r := bytes.NewReader(buf.Bytes())
			actualLength, err := lenEnc(r)
			require.NoError(t, err)

			require.NoError(t, err)
			require.Equal(t, tC.expectedLength, actualLength)

			r.Seek(0, io.SeekStart)
			res, err := ReadLenEncString(r)
			require.NoError(t, err)
			require.Equal(t, tC.input, string(res))
		})
	}
}
