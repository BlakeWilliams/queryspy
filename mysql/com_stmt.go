package mysql

type ComStmtPreparePacket struct {
	*rawPacket
}

func newComStmtPreparePacket(p *rawPacket, proxy *Proxy) *ComStmtPreparePacket {
	return &ComStmtPreparePacket{
		rawPacket: p,
	}
}

func (p *ComStmtPreparePacket) Payload() []byte {
	return p.payload[1:]
}
