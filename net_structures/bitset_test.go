package net_structures_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestBitSet(t *testing.T) {
	tests := []struct {
		name string
		val  ns.BitSet
	}{
		{"empty", ns.BitSet{Length: 0, Data: []uint64{}}},
		{"single", ns.BitSet{Length: 64, Data: []uint64{0xFFFFFFFFFFFFFFFF}}},
		{"multiple", ns.BitSet{Length: 128, Data: []uint64{0x0123456789ABCDEF, 0xFEDCBA9876543210}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("BitSet.Marshal() error = %v", err)
			}
			var unmarshaled ns.BitSet
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalBitSet() error = %v", err)
			}
			if len(unmarshaled.Data) != len(tt.val.Data) {
				t.Errorf("UnmarshalBitSet() len = %v, want %v", len(unmarshaled.Data), len(tt.val.Data))
			}
			for i := range tt.val.Data {
				if unmarshaled.Data[i] != tt.val.Data[i] {
					t.Errorf("UnmarshalBitSet() Data[%d] = %x, want %x", i, unmarshaled.Data[i], tt.val.Data[i])
				}
			}
		})
	}
}

func TestFixedBitSet(t *testing.T) {
	tests := []struct {
		name string
		val  ns.FixedBitSet
	}{
		{"8 bits", ns.FixedBitSet{Length: 8, Data: []byte{0xFF}}},
		{"16 bits", ns.FixedBitSet{Length: 16, Data: []byte{0xFF, 0x00}}},
		{"12 bits", ns.FixedBitSet{Length: 12, Data: []byte{0xFF, 0x0F}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("FixedBitSet.Marshal() error = %v", err)
			}
			expectedLen := (tt.val.Length + 7) / 8
			if len(marshaled) != expectedLen {
				t.Errorf("FixedBitSet.Marshal() len = %v, want %v", len(marshaled), expectedLen)
			}

			var unmarshaled ns.FixedBitSet
			unmarshaled.Length = tt.val.Length
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalFixedBitSet() error = %v", err)
			}
			if unmarshaled.Length != tt.val.Length {
				t.Errorf("UnmarshalFixedBitSet() Length = %v, want %v", unmarshaled.Length, tt.val.Length)
			}
		})
	}
}

func TestBitSetErrorCases(t *testing.T) {
	t.Run("negative BitSet length", func(t *testing.T) {
		negOne := ns.VarInt(-1)
		data, _ := negOne.ToBytes() // Negative length
		data = append(data, make([]byte, 8)...)
		var temp ns.BitSet
		_, err := temp.FromBytes(data)
		if err == nil {
			t.Error("BitSet.Unmarshal() should error on negative length")
		}
	})

	t.Run("insufficient data for FixedBitSet", func(t *testing.T) {
		var temp ns.FixedBitSet
		temp.Length = 16
		_, err := temp.FromBytes(ns.ByteArray{0xFF}) // Need 2 bytes for 16 bits
		if err == nil {
			t.Error("FixedBitSet.Unmarshal() should error on insufficient data")
		}
	})
}

func TestBitSetInterface(t *testing.T) {
	val := ns.BitSet{
		Length: 128,
		Data:   []uint64{0x1234567890ABCDEF, 0xFEDCBA0987654321},
	}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("BitSet.Marshal() error = %v", err)
	}

	var result ns.BitSet
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("BitSet.Unmarshal() error = %v", err)
	}
	if result.Length != val.Length || len(result.Data) != len(val.Data) {
		t.Errorf("BitSet interface roundtrip: got %+v, want %+v", result, val)
	}
	for i := range result.Data {
		if result.Data[i] != val.Data[i] {
			t.Errorf("BitSet interface roundtrip data mismatch at index %d: got %x, want %x", i, result.Data[i], val.Data[i])
		}
	}
}

func TestFixedBitSetInterface(t *testing.T) {
	val := ns.FixedBitSet{
		Length: 16,
		Data:   []byte{0xFF, 0x00},
	}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("FixedBitSet.Marshal() error = %v", err)
	}

	var result ns.FixedBitSet
	result.Length = 16
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("FixedBitSet.Unmarshal() error = %v", err)
	}
	if result.Length != val.Length {
		t.Errorf("FixedBitSet interface roundtrip: got length %d, want %d", result.Length, val.Length)
	}
}

func TestBitSetGenericMarshal(t *testing.T) {
	val := ns.BitSet{
		Length: 128,
		Data:   []uint64{0x1234567890ABCDEF, 0xFEDCBA0987654321},
	}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Marshal(BitSet) error = %v", err)
	}

	var result ns.BitSet
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Unmarshal(BitSet) error = %v", err)
	}
	if result.Length != val.Length || len(result.Data) != len(val.Data) {
		t.Errorf("Generic BitSet roundtrip: got %+v, want %+v", result, val)
	}
	for i := range result.Data {
		if result.Data[i] != val.Data[i] {
			t.Errorf("Generic BitSet roundtrip data mismatch at index %d: got %x, want %x", i, result.Data[i], val.Data[i])
		}
	}
}
