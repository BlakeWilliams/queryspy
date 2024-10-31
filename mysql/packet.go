package mysql

import (
	"fmt"
	"io"
)

// Packet represents a general MySQL packet.
type Packet struct {
	header     []byte
	rawPayload []byte
}

// ReadFrom reads from a given reader to populate the packet data
func (p *Packet) ReadFrom(conn io.Reader) error {
	p.header = make([]byte, 4)
	// Read the 4-byte header
	if _, err := io.ReadFull(conn, p.header); err != nil {
		return err
	}

	packetLength := int(p.header[0]) | int(p.header[1])<<8 | int(p.header[2])<<16

	// Read the payload based on the packet length
	p.rawPayload = make([]byte, packetLength)
	if _, err := io.ReadFull(conn, p.rawPayload); err != nil {
		return err
	}

	return nil
}

// WriteTo writes the packet to a given writer, copying the bytes received
// from `ReadFrom` as-is
func (p *Packet) WriteTo(conn io.Writer) error {
	if _, err := conn.Write(p.header); err != nil {
		return err
	}
	if _, err := conn.Write(p.rawPayload); err != nil {
		return err
	}

	return nil
}

// SeqID returns the sequence ID of the packet
func (p *Packet) SeqID() int {
	return int(p.header[3])
}

// Payload returns the payload of the packet, omitting the first 3 reserved
// bytes. This is only designed to be called for ComQuery
func (p *Packet) Payload() []byte {
	return p.rawPayload[3:]
}

// RawPayload returns the raw payload of the packet
func (p *Packet) RawPayload() []byte {
	return p.rawPayload
}

func (p *Packet) Command() byte {
	if len(p.rawPayload) < 1 {
		return 0
	}

	return p.rawPayload[0]
}

func (p *Packet) CommandName() string {
	if len(p.rawPayload) < 1 {
		return "NIL"
	}

	if res, ok := commandNames[p.rawPayload[0]]; ok {
		return res
	}

	return fmt.Sprintf("%d", p.rawPayload[0])
}
