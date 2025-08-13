package net_structures

import "errors"

// Variable-length data encoding a two's complement signed 32-bit integer.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:VarInt
type VarInt int32

func (v VarInt) Len() int {
	value := uint32(v)
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

func (v VarInt) ToBytes() (ByteArray, error) {
	var data []byte
	value := uint32(v)
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

func (v *VarInt) FromBytes(data ByteArray) (int, error) {
	var value uint32
	var position int
	currentByte := byte(0)

	for i := 0; i < len(data); i++ {
		currentByte = data[i]
		value |= uint32(currentByte&0x7F) << position

		if (currentByte & 0x80) == 0 {
			*v = VarInt(int32(value))
			return i + 1, nil
		}

		position += 7

		if position >= 32 {
			return 0, errors.New("VarInt too big")
		}
	}

	return 0, errors.New("incomplete VarInt")
}
