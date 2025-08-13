package net_structures

import "errors"

// Variable-length data encoding a two's complement signed 64-bit integer.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:VarLong
type VarLong int64

func (v VarLong) Len() int {
	value := uint64(v)
	length := 0
	for {
		length++
		value >>= 7
		if value == 0 {
			break
		}
	}
	return length
}

func (v VarLong) ToBytes() (ByteArray, error) {
	var data []byte
	value := uint64(v)
	for {
		temp := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			temp |= 0x80
		}
		data = append(data, temp)
		if value == 0 {
			break
		}
	}
	return ByteArray(data), nil
}

func (v *VarLong) FromBytes(data ByteArray) (int, error) {
	var value uint64
	var position int
	currentByte := byte(0)

	for i := 0; i < len(data); i++ {
		currentByte = data[i]
		value |= uint64(currentByte&0x7F) << position

		if (currentByte & 0x80) == 0 {
			*v = VarLong(int64(value))
			return i + 1, nil
		}

		position += 7

		if position >= 64 {
			return 0, errors.New("VarLong too big")
		}
	}

	return 0, errors.New("incomplete VarLong")
}
