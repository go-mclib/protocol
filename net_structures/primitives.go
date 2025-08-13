package net_structures

import (
	"encoding/binary"
	"errors"
	"math"
)

// Boolean is either true or false. True is encoded as `0x01`, false as `0x00`.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Boolean
type Boolean bool

func (b Boolean) ToBytes() (ByteArray, error) {
	if b {
		return ByteArray{0x01}, nil
	}
	return ByteArray{0x00}, nil
}

func (b *Boolean) FromBytes(data ByteArray) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for Boolean")
	}
	*b = Boolean(data[0] != 0)
	return 1, nil
}

// An integer between -128 and 127.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Byte
type Byte int8

func (b Byte) ToBytes() (ByteArray, error) {
	return ByteArray{byte(b)}, nil
}

func (b *Byte) FromBytes(data ByteArray) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for Byte")
	}
	*b = Byte(int8(data[0]))
	return 1, nil
}

// An integer between 0 and 255.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Unsigned_Byte
type UnsignedByte uint8

func (ub UnsignedByte) ToBytes() (ByteArray, error) {
	return ByteArray{byte(ub)}, nil
}

func (ub *UnsignedByte) FromBytes(data ByteArray) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for UnsignedByte")
	}
	*ub = UnsignedByte(data[0])
	return 1, nil
}

// An integer between -32768 and 32767.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Short
type Short int16

func (s Short) ToBytes() (ByteArray, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(s))
	return data, nil
}

func (s *Short) FromBytes(data ByteArray) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("insufficient data for Short")
	}
	*s = Short(int16(binary.BigEndian.Uint16(data)))
	return 2, nil
}

// An integer between 0 and 65535.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Unsigned_Short
type UnsignedShort uint16

func (us UnsignedShort) ToBytes() (ByteArray, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(us))
	return data, nil
}

func (us *UnsignedShort) FromBytes(data ByteArray) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("insufficient data for UnsignedShort")
	}
	*us = UnsignedShort(binary.BigEndian.Uint16(data))
	return 2, nil
}

// An integer between -2147483648 and 2147483647.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Int
type Int int32

func (i Int) ToBytes() (ByteArray, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(i))
	return data, nil
}

func (i *Int) FromBytes(data ByteArray) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("insufficient data for Int")
	}
	*i = Int(int32(binary.BigEndian.Uint32(data)))
	return 4, nil
}

// An integer between -9223372036854775808 and 9223372036854775807.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Long
type Long int64

func (l Long) ToBytes() (ByteArray, error) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(l))
	return data, nil
}

func (l *Long) FromBytes(data ByteArray) (int, error) {
	if len(data) < 8 {
		return 0, errors.New("insufficient data for Long")
	}
	*l = Long(int64(binary.BigEndian.Uint64(data)))
	return 8, nil
}

// A single-precision 32-bit IEEE 754 floating point number.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Float
type Float float32

func (f Float) ToBytes() (ByteArray, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(float32(f)))
	return data, nil
}

func (f *Float) FromBytes(data ByteArray) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("insufficient data for Float")
	}
	*f = Float(math.Float32frombits(binary.BigEndian.Uint32(data)))
	return 4, nil
}

// A double-precision 64-bit IEEE 754 floating point number.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Double
type Double float64

func (d Double) ToBytes() (ByteArray, error) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, math.Float64bits(float64(d)))
	return data, nil
}

func (d *Double) FromBytes(data ByteArray) (int, error) {
	if len(data) < 8 {
		return 0, errors.New("insufficient data for Double")
	}
	*d = Double(math.Float64frombits(binary.BigEndian.Uint64(data)))
	return 8, nil
}
