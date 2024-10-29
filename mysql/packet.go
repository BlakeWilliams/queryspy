package mysql

import (
	"fmt"
	"io"
)

type Packet struct {
	header     []byte
	rawPayload []byte
}

func (p *Packet) Raw() []byte {
	raw := make([]byte, 0, len(p.header)+len(p.rawPayload))

	raw = append(raw, p.header...)
	raw = append(raw, p.rawPayload...)

	return raw
}

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

func (p *Packet) WriteTo(conn io.Writer) error {
	if _, err := conn.Write(p.header); err != nil {
		return err
	}
	if _, err := conn.Write(p.rawPayload); err != nil {
		return err
	}

	return nil
}

func (p *Packet) Payload() []byte {
	return p.rawPayload[3:]
}

func (p *Packet) Command() byte {
	return p.rawPayload[0]
}

func (p *Packet) CommandName() string {
	if res, ok := commandNames[p.rawPayload[0]]; ok {
		return res
	}

	return fmt.Sprintf("%d", p.rawPayload[0])
}
