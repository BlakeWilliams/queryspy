package mysql

import (
	"io"
)

type serverAuthPacket struct {
	ProtocolVersion int
	MySQLVersion    string

	lowerCapabilities []byte
	upperCapabilities []byte
	*rawPacket
}

func NewAuthPacket(conn io.Reader) (*serverAuthPacket, error) {
	packet := &rawPacket{}
	err := packet.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	payload := packet.payload
	protocolVersion := payload[0]

	// skip protocol
	payload = payload[1:]

	versionEnd := 0
	for i, b := range payload {
		if b == 0x00 {
			versionEnd = i
			break
		}
	}

	version := payload[:versionEnd]
	payload = payload[versionEnd+1:]

	// 4 for thread, 8 for auth-plugin-data, 1 for filler
	payload = payload[4+8+1:]
	lowerCapabilities := payload[:2]

	// 1 for character set, 2 for character set
	payload = payload[1+2:]
	upperCapabilities := payload[:2]

	return &serverAuthPacket{
		rawPacket:         packet,
		MySQLVersion:      string(version),
		ProtocolVersion:   int(protocolVersion),
		lowerCapabilities: lowerCapabilities,
		upperCapabilities: upperCapabilities,
	}, nil
}

func (p *serverAuthPacket) RemoveSSLSupport() {
	p.lowerCapabilities[1] &^= 0x08
}
