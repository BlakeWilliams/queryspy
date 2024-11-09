package mysql

import (
	"encoding/binary"
)

func LenEncString(message string) []byte {
	var packet []byte
	messageLen := uint64(len([]byte(message)))

	if messageLen >= 1<<24 {
		packet = make([]byte, 0, messageLen+9)

		lenenc := make([]byte, 8)
		binary.LittleEndian.PutUint64(lenenc, uint64(messageLen))
		packet = append(packet, 0xFE)
		packet = append(packet, lenenc...)
	} else if messageLen >= 1<<16 {
		packet = make([]byte, 0, messageLen+4)

		lenenc := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenenc, uint32(messageLen))
		packet = append(packet, 0xFD)
		packet = append(packet, lenenc[:3]...)
	} else if messageLen >= 251 {
		packet = make([]byte, 0, messageLen+3)

		lenenc := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenenc, uint16(messageLen))

		packet = append(packet, 0xFC)
		packet = append(packet, lenenc...)
	} else {
		packet = make([]byte, 0, messageLen+1)

		lenenc := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenenc, uint16(messageLen))
		packet = append(packet, lenenc[0])
	}

	packet = append(packet, []byte(message)...)

	return packet
}

func LenEnc(message []byte) uint64 {
	switch message[0] {
	case 0xFC:
		return uint64(int(message[2]<<8) | int(message[1]))
	case 0xFD:
		return uint64(message[1]) | uint64(message[2])<<8 | uint64(message[3])<<16
	case 0xFE:
		return uint64(message[1]) | uint64(message[2])<<8 | uint64(message[3])<<16 | uint64(message[4])<<24 | uint64(message[5])<<32 | uint64(message[6])<<40 | uint64(message[7])<<48 | uint64(message[8])<<56
	default:
		return uint64(message[0])
	}
}
