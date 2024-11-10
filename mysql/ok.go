package mysql

import (
	"bytes"
	"encoding/binary"
)

func NewOKPacket(originalPacket Packet, message string, clientCapabilities clientCapabilities) *rawPacket {
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

	if clientCapabilities.SessionTrack {
		LenEncString(&payload, message)
	} else {
		payload.Write([]byte(message))
		payload.WriteByte(0x00)
	}

	headerFragment := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerFragment, uint32(payload.Len()))
	// Hacky, but overwrite the sequence ID with the original packet's sequence ID
	headerFragment[3] = originalPacket.Seq() + 1

	return &rawPacket{
		header:  headerFragment,
		payload: payload.Bytes(),
	}
}
