package mysql

import (
	"bytes"
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

type ComQueryPacket struct {
	payload []byte
	*rawPacket
}

func newComQueryPacket(p *rawPacket, proxy *Proxy) *ComQueryPacket {
	r := bytes.NewReader(p.payload)
	// Read one to skip command
	_, _ = r.ReadByte()

	if proxy.clientCapabilities.QueryAttributes && p.isClient {
		paramCount, _ := lenEnc(r)
		_, _ = lenEnc(r)

		if paramCount > 0 {
			nullBitMap := make([]byte, (paramCount+7)/8)
			r.Read(nullBitMap)

			bindFlags := make([]byte, 1)
			r.Read(bindFlags)

			if int(bindFlags[0]) != 1 {
				proxy.Logger.Warn("unsupported bind flags", "flags", bindFlags[0])
			}

			if int(bindFlags[0]) == 1 {
				for i := uint64(0); i < paramCount; i++ {
					isNull := (nullBitMap[i/8] & (1 << (i % 8))) != 0
					if isNull {
						continue
					}
					skipBinary(r)
				}
			}
		}
	}

	c := make([]byte, 1)
	var payload bytes.Buffer

	_, err := r.Read(c)
	for err == nil && c[0] != 0x00 {
		payload.Write(c)
		_, err = r.Read(c)
		if err != nil {
			break
		}
	}

	return &ComQueryPacket{
		rawPacket: p,
		payload:   payload.Bytes(),
	}
}

func (p *ComQueryPacket) Payload() []byte {
	return p.payload
}
