package mysql

import (
	"encoding/binary"
)

var (
	ClientCapabilityClientProtocol41 uint32 = 0x200
	ClientCapabilitySessionTrack     uint32 = 0x00800000
	ClientCapabilityQueryAttributes  uint32 = 1 << 27
)

type clientCapabilities struct {
	Protocol41      bool
	SessionTrack    bool
	QueryAttributes bool
}

func clientCapabilitiesFrom(b []byte) clientCapabilities {
	capabilities := binary.LittleEndian.Uint32(b)

	return clientCapabilities{
		Protocol41:      capabilities&ClientCapabilityClientProtocol41 == ClientCapabilityClientProtocol41,
		SessionTrack:    capabilities&ClientCapabilitySessionTrack == ClientCapabilitySessionTrack,
		QueryAttributes: capabilities&ClientCapabilityQueryAttributes == ClientCapabilityQueryAttributes,
	}
}
