package mysql

import (
	"io"
)

type AuthPacket struct {
	ProtocolVersion int
	MySQLVersion    string

	capabilities []byte
	*Packet
}

func NewAuthPacket(conn io.Reader) (*AuthPacket, error) {
	packet := &Packet{}
	err := packet.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	payload := packet.rawPayload
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
	capabilities := payload[:2]

	return &AuthPacket{
		Packet:          packet,
		MySQLVersion:    string(version),
		ProtocolVersion: int(protocolVersion),
		capabilities:    capabilities,
	}, nil
}

func (p *AuthPacket) RemoveSSLSupport() {
	p.capabilities[1] &^= 0x08
}
