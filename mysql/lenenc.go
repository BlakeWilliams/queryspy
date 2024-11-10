package mysql

import (
	"fmt"
	"io"
)

func LenEncString(w io.Writer, message string) {
	messageLen := uint64(len([]byte(message)))

	if messageLen >= 1<<24 {
		w.Write([]byte{
			0xFE,
			byte(messageLen),
			byte(messageLen >> 8),
			byte(messageLen >> 16),
			byte(messageLen >> 24),
			byte(messageLen >> 32),
			byte(messageLen >> 40),
			byte(messageLen >> 48),
			byte(messageLen >> 56),
		})
	} else if messageLen >= 1<<16 {
		w.Write([]byte{
			0xFD,
			byte(messageLen),
			byte(messageLen >> 8),
			byte(messageLen >> 16),
		})
	} else if messageLen >= 251 {
		w.Write([]byte{
			0xFC,
			byte(messageLen),
			byte(messageLen >> 8),
		})
	} else {
		w.Write([]byte{
			byte(messageLen),
		})
	}

	w.Write([]byte(message))
}

// lenEnc accepts an io.Reader and returns the number of size of the encoded string.
func lenEnc(r io.Reader) (uint64, error) {
	initialSize := make([]byte, 1)
	_, err := r.Read(initialSize)
	if err != nil {
		return 0, fmt.Errorf("error reading size of lenenc string: %w", err)
	}

	switch initialSize[0] {
	case 0xFC:
		size := make([]byte, 2)
		_, err := r.Read(size)
		if err != nil {
			return 0, fmt.Errorf("error reading remainign size of lenenc string: %w", err)
		}

		return uint64(int(size[1]<<8) | int(size[0])), nil
	case 0xFD:
		size := make([]byte, 3)
		_, err := r.Read(size)
		if err != nil {
			return 0, fmt.Errorf("error reading remainign size of lenenc string: %w", err)
		}

		return uint64(size[0]) | uint64(size[1])<<8 | uint64(size[2])<<16, nil
	case 0xFE:
		size := make([]byte, 8)
		_, err := r.Read(size)
		if err != nil {
			return 0, fmt.Errorf("error reading remainign size of lenenc string: %w", err)
		}

		return uint64(size[0]) | uint64(size[1])<<8 | uint64(size[2])<<16 | uint64(size[3])<<24 | uint64(size[4])<<32 | uint64(size[5])<<40 | uint64(size[6])<<48 | uint64(size[7])<<56, nil
	default:
		return uint64(initialSize[0]), nil
	}
}

// ReadLenEncString reads a lenenc string from an io.Reader and returns the
// number of bytes read and the string.
func ReadLenEncString(r io.Reader) ([]byte, error) {
	payloadSize, err := lenEnc(r)
	if err != nil {
		return nil, fmt.Errorf("error reading lenenc string: %w", err)
	}

	message := make([]byte, payloadSize)
	_, err = r.Read(message)
	if err != nil {
		return nil, fmt.Errorf("error reading lenenc string: %w", err)
	}

	return message, nil
}
