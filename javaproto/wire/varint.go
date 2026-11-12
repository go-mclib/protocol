package wire

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	errVarIntTooBig  = errors.New("VarInt is too big")
	errVarLongTooBig = errors.New("VarLong is too big")
)

// VarInt is a variable-length signed 32-bit integer.
//
// Uses 7 bits per byte with bit 7 as continuation flag.
// Values are encoded in little-endian order within the variable-length format.
// Maximum 5 bytes for 32-bit values.
//
// Encoding:
//
//	For each byte:
//	  - Bits 0-6: 7 bits of data
//	  - Bit 7: 1 if more bytes follow, 0 if this is the last byte
//
// Examples:
//
//	0          -> [0x00]
//	1          -> [0x01]
//	127        -> [0x7f]
//	128        -> [0x80, 0x01]
//	255        -> [0xff, 0x01]
//	2147483647 -> [0xff, 0xff, 0xff, 0xff, 0x07]
//	-1         -> [0xff, 0xff, 0xff, 0xff, 0x0f]
type VarInt int32

// Append appends the VarInt encoding to dst.
func (v VarInt) Append(dst []byte) []byte {
	return binary.AppendUvarint(dst, uint64(uint32(v)))
}

// Len returns the number of bytes needed to encode this VarInt.
func (v VarInt) Len() int {
	value := uint32(v)
	switch {
	case value < 1<<7:
		return 1
	case value < 1<<14:
		return 2
	case value < 1<<21:
		return 3
	case value < 1<<28:
		return 4
	default:
		return 5
	}
}

// DecodeVarInt reads a VarInt from r.
func DecodeVarInt(r io.Reader) (VarInt, error) {
	var value int32
	var position uint
	var b [1]byte

	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		value |= int32(b[0]&0x7F) << position

		if (b[0] & 0x80) == 0 {
			break
		}

		position += 7
		if position >= 35 {
			return 0, errVarIntTooBig
		}
	}

	return VarInt(value), nil
}

// VarLong is a variable-length signed 64-bit integer.
//
// Same encoding as VarInt but for 64-bit values.
// Maximum 10 bytes for 64-bit values.
//
// Examples:
//
//	0                    -> [0x00]
//	9223372036854775807  -> [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f]
//	-1                   -> [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01]
type VarLong int64

// Append appends the VarLong encoding to dst.
func (v VarLong) Append(dst []byte) []byte {
	return binary.AppendUvarint(dst, uint64(v))
}

// Len returns the number of bytes needed to encode this VarLong.
func (v VarLong) Len() int {
	value := uint64(v)
	n := 1
	for value >= 0x80 {
		value >>= 7
		n++
	}
	return n
}
