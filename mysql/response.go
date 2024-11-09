package mysql

import (
	"bytes"
	"encoding/binary"
)

func NewOKPacket(originalPacket *Packet, message string) *Packet {
	var payload bytes.Buffer

	payload.Write([]byte{
		// assign 0 as command for OK header
		0x00,
		// assign 0 for number of affected rows
		0x00,
		// assign 0 as last insert id
		0x00,
		// assign 2 bytes of 0's for server status flags
		0x00,
		0x00,
		// assign 2 bytes of 0's for warnings
		0x00,
		0x00,
	})

	// dead code needed to handle ok packet when ClientCapabilitySessionTrack is _false_
	if false {
		// TODO write nul terminated string if above is not true
	} else {
		LenEncString(&payload, message)
	}

	headerFragment := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerFragment, uint32(payload.Len()))
	// Hacky, but overwrite the sequence ID with the original packet's sequence ID
	headerFragment[3] = originalPacket.RawSeq() + 1

	return &Packet{
		header:     headerFragment,
		rawPayload: payload.Bytes(),
	}
}
