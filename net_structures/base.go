// > All data sent over the network (except for VarInt and VarLong) is big-endian,
// that is the bytes are sent from most significant byte to least significant byte.
// The majority of everyday computers are little-endian, therefore it may be necessary
// to change the endianness before sending data over the network.
//
// (Ref.: https://minecraft.wiki/w/Java_Edition_protocol/Packets#Data_types)
package net_structures

import "errors"

// This is just a sequence of zero or more bytes. It represents any data sent over the wire.
// The length is known from the context.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Byte_Array
type ByteArray []byte

// ToBytes returns the byte array as-is. The length is defined by the surrounding context
// (e.g., packet or field definition), so we do not add any length prefix here.
func (b ByteArray) ToBytes() (ByteArray, error) {
	return b, nil
}

// FromBytes consumes all remaining bytes as the byte array. The expected length is defined by
// the surrounding context; therefore we treat the remainder of the input as this field's value.
func (b *ByteArray) FromBytes(data ByteArray) (int, error) {
	if len(data) == 0 {
		// empty array is valid
		*b = ByteArray{}
		return 0, nil
	}

	// copy to avoid aliasing the input slice
	dst := make(ByteArray, len(data))
	copy(dst, data)
	*b = dst
	return len(data), nil
}

// PrefixedByteArray is a byte array prefixed with a VarInt length.
//
// Many packet fields use a VarInt length prefix, followed by that many bytes.
// Use this type for those fields.
type PrefixedByteArray []byte

func (p PrefixedByteArray) ToBytes() (ByteArray, error) {
	lengthBytes, err := VarInt(len(p)).ToBytes()
	if err != nil {
		return nil, err
	}
	out := make(ByteArray, 0, len(lengthBytes)+len(p))
	out = append(out, lengthBytes...)
	out = append(out, []byte(p)...)
	return out, nil
}

func (p *PrefixedByteArray) FromBytes(data ByteArray) (int, error) {
	var length VarInt
	off, err := length.FromBytes(data)
	if err != nil {
		return 0, err
	}
	if int(length) < 0 || len(data) < off+int(length) {
		return 0, errors.New("insufficient data for PrefixedByteArray")
	}
	dst := make([]byte, int(length))
	copy(dst, data[off:off+int(length)])
	*p = PrefixedByteArray(dst)
	return off + int(length), nil
}
