package net_structures

import (
	"bytes"
	"fmt"

	"github.com/Tnze/go-mc/nbt"
)

// NBT - Named Binary Tag
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:NBT
type NBT struct {
	Data any
}

// NewNBT creates a new NBT instance with the given data
func NewNBT(data any) NBT {
	return NBT{Data: data}
}

// NewEmptyNBT creates a new empty NBT instance
func NewEmptyNBT() NBT {
	return NBT{Data: nil}
}

func (n NBT) ToBytes() (ByteArray, error) {
	if n.Data == nil {
		return ByteArray{0x00}, nil
	}

	var buf bytes.Buffer
	encoder := nbt.NewEncoder(&buf)
	encoder.NetworkFormat(true)
	
	err := encoder.Encode(n.Data, "")
	if err != nil {
		return nil, fmt.Errorf("failed to encode NBT data: %w", err)
	}

	return ByteArray(buf.Bytes()), nil
}

func (n *NBT) FromBytes(data ByteArray) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("insufficient data for NBT")
	}

	if len(data) == 1 && data[0] == 0x00 {
		n.Data = nil
		return 1, nil
	}

	reader := bytes.NewReader(data)
	decoder := nbt.NewDecoder(reader)
	decoder.NetworkFormat(true)

	var nbtData any
	_, err := decoder.Decode(&nbtData)
	if err != nil {
		return 0, fmt.Errorf("failed to decode NBT data: %w", err)
	}

	n.Data = nbtData

	bytesRead := len(data) - reader.Len()
	return bytesRead, nil
}

// DecodeTo decodes the NBT data into the provided destination
func (n *NBT) DecodeTo(dest any) error {
	if n.Data == nil {
		return fmt.Errorf("NBT data is nil")
	}

	encoded, err := n.ToBytes()
	if err != nil {
		return fmt.Errorf("failed to encode NBT for type conversion: %w", err)
	}

	reader := bytes.NewReader(encoded)
	decoder := nbt.NewDecoder(reader)
	decoder.NetworkFormat(true)
	
	_, err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("failed to decode NBT to specific type: %w", err)
	}

	return nil
}

// EncodeFrom encodes data from the provided source into this NBT
func (n *NBT) EncodeFrom(src any) error {
	n.Data = src
	return nil
}

// IsEmpty returns true if the NBT contains no data
func (n NBT) IsEmpty() bool {
	return n.Data == nil
}

// String returns a string representation of the NBT data
func (n NBT) String() string {
	if n.Data == nil {
		return "NBT{empty}"
	}
	return fmt.Sprintf("NBT{%+v}", n.Data)
}
