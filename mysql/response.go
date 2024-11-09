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

	// assign 0 as command for OK header
	payloadFragment = append(payloadFragment, 0x00)

	// assign 0 as affected rows
	payloadFragment = append(payloadFragment, 0x00)

	// assign 0 as last insert id
	payloadFragment = append(payloadFragment, 0x00)

	// assign 0's for server status flags
	payloadFragment = append(payloadFragment, 0x00)
	payloadFragment = append(payloadFragment, 0x00)

	// assign 0's for warnings
	payloadFragment = append(payloadFragment, 0x00)
	payloadFragment = append(payloadFragment, 0x00)

	// dead code needed to handle ok packet when ClientCapabilitySessionTrack is _false_
	if false {
		payloadFragment = append(payloadFragment, []byte(message)...)
		payloadFragment = append(payloadFragment, 0x00)
	} else {

		payloadFragment = append(payloadFragment, LenEncString(message)...)
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
