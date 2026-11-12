package wire_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// PrefixedArray wire format:
//   VarInt length
//   T × length

func TestPrefixedArray(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected []ns.VarInt
	}{
		{
			name:     "empty",
			raw:      []byte{0x00},
			expected: []ns.VarInt{},
		},
		{
			name:     "single element",
			raw:      []byte{0x01, 0x2a},
			expected: []ns.VarInt{42},
		},
		{
			name:     "multiple elements",
			raw:      []byte{0x03, 0x01, 0x02, 0x03},
			expected: []ns.VarInt{1, 2, 3},
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var arr ns.PrefixedArray[ns.VarInt]
			if err := arr.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if len(arr) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(arr), len(tc.expected))
			}
			for i, v := range tc.expected {
				if arr[i] != v {
					t.Errorf("element[%d] mismatch: got %d, want %d", i, arr[i], v)
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			arr := ns.PrefixedArray[ns.VarInt](tc.expected)
			buf := ns.NewWriter()
			if err := arr.EncodeWith(buf, encoder); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// PrefixedOptional wire format:
//   Boolean present
//   T value (if present)

func TestPrefixedOptional(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.PrefixedOptional[ns.VarInt]
	}{
		{
			name:     "absent",
			raw:      []byte{0x00},
			expected: ns.None[ns.VarInt](),
		},
		{
			name:     "present",
			raw:      []byte{0x01, 0x2a},
			expected: ns.Some[ns.VarInt](42),
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var opt ns.PrefixedOptional[ns.VarInt]
			if err := opt.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if opt.Present != tc.expected.Present {
				t.Errorf("Present mismatch: got %v, want %v", opt.Present, tc.expected.Present)
			}
			if opt.Present && opt.Value != tc.expected.Value {
				t.Errorf("Value mismatch: got %d, want %d", opt.Value, tc.expected.Value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.expected.EncodeWith(buf, encoder); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}
