package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

// BitSet wire format:
//   VarInt length (number of longs)
//   Int64 × length (big-endian)

var bitSetTestCases = []struct {
	name     string
	raw      []byte
	expected []int64
}{
	{
		name:     "empty",
		raw:      []byte{0x00},
		expected: []int64{},
	},
	{
		name: "single long with bit 0",
		// length=1, long=1 (bit 0 set)
		raw:      []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		expected: []int64{1},
	},
	{
		name: "single long with bit 63",
		// length=1, long=0x8000000000000000 (bit 63 set)
		raw:      []byte{0x01, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		expected: []int64{-9223372036854775808}, // signed representation
	},
	{
		name: "two longs",
		// length=2, long1=3 (bits 0,1), long2=5 (bits 64,66)
		raw: []byte{
			0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05,
		},
		expected: []int64{3, 5},
	},
}

func TestBitSet(t *testing.T) {
	for _, tc := range bitSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.BitSet
			if err := got.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			longs := got.Longs()
			if len(longs) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(longs), len(tc.expected))
			}
			for i, v := range tc.expected {
				if longs[i] != v {
					t.Errorf("long[%d] mismatch: got %d, want %d", i, longs[i], v)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			bs := bitSetFromLongs(tc.expected)
			buf := ns.NewWriter()
			if err := bs.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestBitSet_GetSet(t *testing.T) {
	bs := ns.NewBitSet(128)

	// initially all bits should be unset
	for i := range 128 {
		if bs.Get(i) {
			t.Errorf("bit %d should be unset", i)
		}
	}

	// set some bits
	bs.Set(0)
	bs.Set(63)
	bs.Set(64)
	bs.Set(127)

	if !bs.Get(0) || !bs.Get(63) || !bs.Get(64) || !bs.Get(127) {
		t.Error("set bits should be set")
	}
	if bs.Get(1) || bs.Get(62) || bs.Get(65) {
		t.Error("unset bits should remain unset")
	}

	// clear a bit
	bs.Clear(63)
	if bs.Get(63) {
		t.Error("cleared bit should be unset")
	}
}

// FixedBitSet wire format:
//   ceil(n/8) bytes (no length prefix)

var fixedBitSetTestCases = []struct {
	name    string
	size    int
	raw     []byte
	setBits []int
}{
	{
		name:    "8 bits, none set",
		size:    8,
		raw:     []byte{0x00},
		setBits: []int{},
	},
	{
		name:    "8 bits, bit 0 set",
		size:    8,
		raw:     []byte{0x01},
		setBits: []int{0},
	},
	{
		name:    "8 bits, bits 0,7 set",
		size:    8,
		raw:     []byte{0x81},
		setBits: []int{0, 7},
	},
	{
		name:    "16 bits, bits 0,8 set",
		size:    16,
		raw:     []byte{0x01, 0x01},
		setBits: []int{0, 8},
	},
}

func TestFixedBitSet(t *testing.T) {
	for _, tc := range fixedBitSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			fbs := ns.NewFixedBitSet(tc.size)
			if err := fbs.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			for _, bit := range tc.setBits {
				if !fbs.Get(bit) {
					t.Errorf("bit %d should be set", bit)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			fbs := ns.NewFixedBitSet(tc.size)
			for _, bit := range tc.setBits {
				fbs.Set(bit)
			}
			buf := ns.NewWriter()
			if err := fbs.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// bitSetFromLongs creates a BitSet from raw long values (test helper).
func bitSetFromLongs(longs []int64) *ns.BitSet {
	bs := ns.NewBitSet(len(longs) * 64)
	for i, v := range longs {
		for bit := range 64 {
			if (v & (1 << bit)) != 0 {
				bs.Set(i*64 + bit)
			}
		}
	}
	return bs
}
