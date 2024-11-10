package mysql

import "io"

const (
	TypeDecimal    uint8 = 0x00
	TypeTiny       uint8 = 0x01
	TypeShort      uint8 = 0x02
	TypeLong       uint8 = 0x03
	TypeFloat      uint8 = 0x04
	TypeDouble     uint8 = 0x05
	TypeNull       uint8 = 0x06
	TypeTimestamp  uint8 = 0x07
	TypeLongLong   uint8 = 0x08
	TypeInt24      uint8 = 0x09
	TypeDate       uint8 = 0x0A
	TypeTime       uint8 = 0x0B
	TypeDateTime   uint8 = 0x0C
	TypeYear       uint8 = 0x0D
	TypeNewdate    uint8 = 0x0E
	TypeVarChar    uint8 = 0x0F
	TypeBit        uint8 = 0x10
	TypeNewDecimal uint8 = 0x11
	TypeEnum       uint8 = 0x12
	TypeSet        uint8 = 0x13
	TypeTinyBlob   uint8 = 0x14
	TypeMediumBlob uint8 = 0x15
	TypeLongBlob   uint8 = 0x16
	TypeBlob       uint8 = 0x17
	TypeVarString  uint8 = 0x18
	TypeString     uint8 = 0x19
	TypeGeometry   uint8 = 0x1F
	TypeJSON       uint8 = 0xF5
)

func skipBinary(r io.Reader) error {
	t := make([]byte, 1)
	_, _ = r.Read(t)

	switch t[0] {
	case TypeString, TypeVarString, TypeVarChar, TypeEnum, TypeSet,
		TypeLongBlob, TypeMediumBlob, TypeBlob, TypeTinyBlob, TypeGeometry,
		TypeBit, TypeDecimal, TypeNewDecimal, TypeJSON:
		ReadLenEncString(r)
	case TypeLong, TypeInt24, TypeFloat:
		r.Read(make([]byte, 4))
	case TypeLongLong, TypeDouble:
		r.Read(make([]byte, 8))
	case TypeShort, TypeYear:
		r.Read(make([]byte, 2))
	case TypeTiny:
		r.Read(make([]byte, 1))
	case TypeDate, TypeDateTime, TypeTimestamp:
		length := make([]byte, 1)
		_, _ = r.Read(length)
		_, _ = r.Read(make([]byte, int(length[0])))
	default:
		panic("unhandled type")
	}

	return nil
}
