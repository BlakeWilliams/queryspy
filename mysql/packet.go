package mysql

import (
	"fmt"
	"io"
)

type Packet interface {
	Payload() []byte
	CommandName() string
	Command() int
	Seq() byte
	WriteTo(io.Writer) error
}

// rawPacket represents a general MySQL packet.
type rawPacket struct {
	header       []byte
	payload      []byte
	capabilities clientCapabilities
	isClient     bool
}

// ReadFrom reads from a given reader to populate the packet data
func (p *rawPacket) ReadFrom(conn io.Reader) error {
	p.header = make([]byte, 4)
	// Read the 4-byte header
	if _, err := io.ReadFull(conn, p.header); err != nil {
		return err
	}

	packetLength := int(p.header[0]) | int(p.header[1])<<8 | int(p.header[2])<<16

	// Read the payload based on the packet length
	p.payload = make([]byte, packetLength)
	if _, err := io.ReadFull(conn, p.payload); err != nil {
		return err
	}

	return nil
}

// WriteTo writes the packet to a given writer, copying the bytes received
// from `ReadFrom` as-is
func (p *rawPacket) WriteTo(conn io.Writer) error {
	if _, err := conn.Write(p.header); err != nil {
		return err
	}
	if _, err := conn.Write(p.payload); err != nil {
		return err
	}

	return nil
}

// SeqID returns the sequence ID of the packet
func (p *rawPacket) Seq() byte {
	return p.header[3]
}

// Payload returns the payload of the packet, omitting the first 3 reserved
// bytes. This is only designed to be called for ComQuery
func (p *rawPacket) Payload() []byte {
	return p.payload
}

func (p *rawPacket) Command() int {
	if len(p.payload) < 1 {
		return 0
	}

	return int(p.payload[0])
}

func (p *rawPacket) CommandName() string {
	if len(p.payload) < 1 {
		return "NIL"
	}

	if res, ok := commandNames[p.payload[0]]; ok {
		return res
	}

	return fmt.Sprintf("%d", p.payload[0])
}
