package wire_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// BitSet wire format:
//   VarInt length (number of bytes)
//   Little-endian bytes, with trailing zero bytes omitted

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
		name:     "bit 0",
		raw:      []byte{0x01, 0x01},
		expected: []int64{1},
	},
	{
		name:     "bit 63",
		raw:      []byte{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
		expected: []int64{-9223372036854775808}, // signed representation
	},
	{
		name: "bits across long boundary",
		raw: []byte{
			0x09,
			0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05,
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

func TestBitSetClearedCapacity(t *testing.T) {
	bs := ns.NewBitSet(192)
	bs.Set(130)
	bs.Clear(130)
	bs.Set(8)
	buf := ns.NewWriter()
	if err := bs.Encode(buf); err != nil {
		t.Fatal(err)
	}
	if want := []byte{2, 0, 1}; !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("cleared high bits must not extend the wire payload: got %x, want %x", buf.Bytes(), want)
	}
	bs.Clear(8)
	buf.Reset()
	if err := bs.Encode(buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{0}) {
		t.Fatalf("empty bitset must encode as zero bytes: %x", buf.Bytes())
	}
}

func TestBitSetDecodeBoundaries(t *testing.T) {
	raw := bytes.NewReader([]byte{2, 1, 1, 0x7f})
	var bs ns.BitSet
	if err := bs.Decode(ns.NewReaderFrom(raw)); err != nil {
		t.Fatal(err)
	}
	if !bs.Get(0) || !bs.Get(8) || bs.Get(1) || raw.Len() != 1 {
		t.Fatal("bitset decoded incorrect bits or consumed the following field")
	}
	for _, malformed := range [][]byte{{2, 1}, {0xff, 0xff, 0xff, 0xff, 0x0f}} {
		if err := bs.Decode(ns.NewReader(malformed)); err == nil {
			t.Fatalf("accepted malformed bitset %x", malformed)
		}
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
			fbs, err := ns.NewReader(tc.raw).ReadFixedBitSet(tc.size)
			if err != nil {
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

func TestFixedBitSetFromShortBytes(t *testing.T) {
	fbs := ns.FixedBitSetFromBytes([]byte{0x01}, 20)
	fbs.Set(19)
	if !fbs.Get(0) || !fbs.Get(19) || len(fbs.Bytes()) != 3 {
		t.Fatalf("short data must be zero-padded to the bit set size: %x", fbs.Bytes())
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
