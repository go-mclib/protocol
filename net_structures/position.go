package net_structures

import (
	"encoding/binary"
	"errors"
)

// Position - block position encoded as a single long value
// x as a 26-bit integer, z as a 26-bit integer, y as a 12-bit integer
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Position
type Position struct {
	X int32 // -33554432 to 33554431
	Y int16 // -2048 to 2047
	Z int32 // -33554432 to 33554431
}

func (p Position) ToBytes() (ByteArray, error) {
	value := uint64(0)
	value |= uint64(p.X&0x3FFFFFF) << 38
	value |= uint64(p.Z&0x3FFFFFF) << 12
	value |= uint64(p.Y & 0xFFF)

	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, value)
	return data, nil
}

func (p *Position) FromBytes(data ByteArray) (int, error) {
	if len(data) < 8 {
		return 0, errors.New("insufficient data for Position")
	}

	value := binary.BigEndian.Uint64(data)

	x := int32(value >> 38)
	if x >= 0x2000000 {
		x -= 0x4000000
	}

	z := int32((value >> 12) & 0x3FFFFFF)
	if z >= 0x2000000 {
		z -= 0x4000000
	}

	y := int16(value & 0xFFF)
	if y >= 0x800 {
		y -= 0x1000
	}

	p.X = x
	p.Y = y
	p.Z = z
	return 8, nil
}
