package mysql

import (
	"encoding/binary"
)

type Response struct {
	Seq    int
	Packet *Packet
}

func NewResponse(originalPacket *Packet, message string) Response {
	res := &Packet{}

	payloadFragment := make([]byte, 0, 30)

	payloadFragment = append(
		payloadFragment,
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
	)

	// dead code needed to handle ok packet when ClientCapabilitySessionTrack is _false_
	if false {
		payloadFragment = append(payloadFragment, []byte(message)...)
	} else {
		LenEncString(payloadFragment, message)
	}

	packetLen := make([]byte, 4)
	binary.BigEndian.PutUint32(packetLen, uint32(len(payloadFragment)))

	headerFragment := make([]byte, 4)
	// MySQL uses little endian, this should be abstracted
	headerFragment[0] = packetLen[3]
	headerFragment[1] = packetLen[2]
	headerFragment[2] = packetLen[1]
	headerFragment[3] = originalPacket.RawSeq() + 1

	res.header = headerFragment
	res.rawPayload = payloadFragment

	return Response{
		Packet: res,
	}
}
