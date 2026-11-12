package wire

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

// BitSet is a dynamically-sized bit set, prefixed by its length in bytes.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  Length (VarInt)│  Little-endian bytes                │
//	└─────────────────┴─────────────────────────────────────┘
//
// The ith bit is set when (Data[i/64] & (1 << (i % 64))) != 0.
// Trailing zero bytes are omitted, matching Java BitSet.toByteArray.
type BitSet struct {
	data []int64
}

// NewBitSet creates a BitSet with the given capacity in bits.
func NewBitSet(capacity int) *BitSet {
	numLongs := (capacity + 63) / 64
	return &BitSet{data: make([]int64, numLongs)}
}

// Decode reads a BitSet from the buffer.
func (b *BitSet) Decode(buf *PacketBuffer) error {
	length, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read bitset length: %w", err)
	}
	raw, err := buf.ReadFixedByteArray(int(length))
	if err != nil {
		return fmt.Errorf("failed to read bitset bytes: %w", err)
	}

	b.data = make([]int64, (len(raw)+7)/8)
	for i := range b.data {
		var word [8]byte
		copy(word[:], raw[i*8:])
		b.data[i] = int64(binary.LittleEndian.Uint64(word[:]))
	}
	return nil
}

// Encode writes a BitSet to the buffer.
func (b *BitSet) Encode(buf *PacketBuffer) error {
	words := len(b.data)
	for words > 0 && b.data[words-1] == 0 {
		words--
	}
	length := 0
	if words > 0 {
		length = (words-1)*8 + (bits.Len64(uint64(b.data[words-1]))+7)/8
	}
	if err := buf.WriteVarInt(VarInt(length)); err != nil {
		return fmt.Errorf("failed to write bitset length: %w", err)
	}
	var word [8]byte
	for i, v := range b.data[:words] {
		binary.LittleEndian.PutUint64(word[:], uint64(v))
		if _, err := buf.Write(word[:min(length-i*8, 8)]); err != nil {
			return fmt.Errorf("failed to write bitset bytes: %w", err)
		}
	}
	return nil
}

// Get returns whether the bit at index i is set.
func (b *BitSet) Get(i int) bool {
	if i < 0 || i/64 >= len(b.data) {
		return false
	}
	return (b.data[i/64] & (1 << (i % 64))) != 0
}

// Set sets the bit at index i.
func (b *BitSet) Set(i int) {
	if i < 0 {
		return
	}
	idx := i / 64
	for len(b.data) <= idx {
		b.data = append(b.data, 0)
	}
	b.data[idx] |= 1 << (i % 64)
}

// Clear clears the bit at index i.
func (b *BitSet) Clear(i int) {
	if i < 0 || i/64 >= len(b.data) {
		return
	}
	b.data[i/64] &^= 1 << (i % 64)
}

// Longs returns the underlying long array.
func (b *BitSet) Longs() []int64 {
	return b.data
}

// FixedBitSet is a fixed-size bit set encoded as ceil(n/8) bytes.
//
// Wire format:
//
//	┌─────────────────────────────────────────────────────────┐
//	│  Bytes (ceil(n/8) bytes, no length prefix)              │
//	└─────────────────────────────────────────────────────────┘
//
// The ith bit is set when (Data[i/8] & (1 << (i % 8))) != 0.
type FixedBitSet struct {
	data []byte
	size int // number of bits
}

// NewFixedBitSet creates a FixedBitSet with the given size in bits.
func NewFixedBitSet(size int) *FixedBitSet {
	numBytes := (size + 7) / 8
	return &FixedBitSet{data: make([]byte, numBytes), size: size}
}

// FixedBitSetFromBytes creates a FixedBitSet of size bits from raw bytes,
// truncating or zero-padding data to ceil(size/8) bytes.
func FixedBitSetFromBytes(data []byte, size int) *FixedBitSet {
	b := NewFixedBitSet(size)
	copy(b.data, data)
	return b
}

// ReadFixedBitSet reads a FixedBitSet of size bits (ceil(size/8) bytes).
func (pb *PacketBuffer) ReadFixedBitSet(size int) (*FixedBitSet, error) {
	data, err := pb.ReadFixedByteArray((size + 7) / 8)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixed bitset: %w", err)
	}
	return &FixedBitSet{data: data, size: size}, nil
}

// Encode writes a FixedBitSet to the buffer.
func (b *FixedBitSet) Encode(buf *PacketBuffer) error {
	return buf.WriteFixedByteArray(b.data)
}

// Get returns whether the bit at index i is set.
func (b *FixedBitSet) Get(i int) bool {
	if i < 0 || i >= b.size {
		return false
	}
	return (b.data[i/8] & (1 << (i % 8))) != 0
}

// Set sets the bit at index i.
func (b *FixedBitSet) Set(i int) {
	if i < 0 || i >= b.size {
		return
	}
	b.data[i/8] |= 1 << (i % 8)
}

// Clear clears the bit at index i.
func (b *FixedBitSet) Clear(i int) {
	if i < 0 || i >= b.size {
		return
	}
	b.data[i/8] &^= 1 << (i % 8)
}

// Size returns the number of bits in the set.
func (b *FixedBitSet) Size() int {
	return b.size
}

// Bytes returns the underlying byte array.
func (b *FixedBitSet) Bytes() []byte {
	return b.data
}
