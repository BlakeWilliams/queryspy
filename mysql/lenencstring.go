package mysql

func LenEncString(packet []byte, message string) []byte {
	messageLen := uint64(len([]byte(message)))

	if messageLen >= 1<<24 {
		packet = make([]byte, 0, messageLen+9)

		packet = append(
			packet,
			0xFE,
			byte(messageLen),
			byte(messageLen>>8),
			byte(messageLen>>16),
			byte(messageLen>>24),
			byte(messageLen>>32),
			byte(messageLen>>40),
			byte(messageLen>>48),
			byte(messageLen>>56),
		)
	} else if messageLen >= 1<<16 {
		packet = make([]byte, 0, messageLen+4)

		packet = append(
			packet,
			0xFD,
			byte(messageLen),
			byte(messageLen>>8),
			byte(messageLen>>16),
		)
	} else if messageLen >= 251 {
		packet = append(
			packet,
			0xFC,
			byte(messageLen),
			byte(messageLen>>8),
		)
	} else {
		packet = append(
			packet,
			byte(messageLen),
		)
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

func ReadLenEncString(message []byte) []byte {
	payloadSize := LenEnc(message)
	offset := uint64(len(message)) - payloadSize

	return message[offset:]
}
