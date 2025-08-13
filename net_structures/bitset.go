package net_structures

import (
	"encoding/binary"
	"errors"
)

// BitSet - length-prefixed bit set
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:BitSet
type BitSet struct {
	Length int
	Data   []uint64
}

func (b BitSet) ToBytes() (ByteArray, error) {
	result, err := VarInt(len(b.Data)).ToBytes()
	if err != nil {
		return nil, err
	}
	for _, v := range b.Data {
		data := make([]byte, 8)
		binary.BigEndian.PutUint64(data, v)
		result = append(result, data...)
	}
	return result, nil
}

func (b *BitSet) FromBytes(data ByteArray) (int, error) {
	var length VarInt
	bytesRead, err := length.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if int(length) < 0 {
		return 0, errors.New("negative BitSet length")
	}

	totalBytes := bytesRead + int(length)*8
	if len(data) < totalBytes {
		return 0, errors.New("insufficient data for BitSet")
	}

	b.Length = int(length) * 64
	b.Data = make([]uint64, length)

	offset := bytesRead
	for i := 0; i < int(length); i++ {
		b.Data[i] = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}

	return totalBytes, nil
}

// FixedBitSet - bit set with fixed length of n bits
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Fixed_BitSet
type FixedBitSet struct {
	Length int
	Data   []byte
}

func (b FixedBitSet) ToBytes() (ByteArray, error) {
	numBytes := (b.Length + 7) / 8
	if len(b.Data) < numBytes {
		data := make([]byte, numBytes)
		copy(data, b.Data)
		return ByteArray(data), nil
	}
	return ByteArray(b.Data[:numBytes]), nil
}

func (b *FixedBitSet) FromBytes(data ByteArray) (int, error) {
	numBytes := (b.Length + 7) / 8
	if len(data) < numBytes {
		return 0, errors.New("insufficient data for FixedBitSet")
	}

	b.Data = make([]byte, numBytes)
	copy(b.Data, data[:numBytes])
	return numBytes, nil
}

// EnumSet - bitset associated to an enum
type EnumSet FixedBitSet
