package mysql

import "bytes"

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
