package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

// IDSet wire format:
//   VarInt type (0 = tag, >0 = inline count + 1)
//   if type=0: Identifier (tag name)
//   if type>0: VarInt × (type-1) IDs

var idSetTestCases = []struct {
	name    string
	raw     []byte
	isTag   bool
	tagName ns.Identifier
	ids     []ns.VarInt
}{
	{
		name: "tag reference",
		// type=0, identifier="minecraft:test" (length 14)
		raw:     []byte{0x00, 0x0e, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 't', 'e', 's', 't'},
		isTag:   true,
		tagName: "minecraft:test",
	},
	{
		name:  "empty inline",
		raw:   []byte{0x01},
		isTag: false,
		ids:   []ns.VarInt{},
	},
	{
		name:  "single inline ID",
		raw:   []byte{0x02, 0x2a},
		isTag: false,
		ids:   []ns.VarInt{42},
	},
	{
		name:  "multiple inline IDs",
		raw:   []byte{0x04, 0x01, 0x02, 0x03},
		isTag: false,
		ids:   []ns.VarInt{1, 2, 3},
	},
}

func TestIDSet(t *testing.T) {
	for _, tc := range idSetTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.IDSet
			if err := got.Decode(ns.NewReader(tc.raw)); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsTag != tc.isTag {
				t.Errorf("IsTag mismatch: got %v, want %v", got.IsTag, tc.isTag)
			}
			if got.IsTag {
				if got.TagName != tc.tagName {
					t.Errorf("TagName mismatch: got %q, want %q", got.TagName, tc.tagName)
				}
			} else {
				if len(got.IDs) != len(tc.ids) {
					t.Fatalf("IDs length mismatch: got %d, want %d", len(got.IDs), len(tc.ids))
				}
				for i, id := range tc.ids {
					if got.IDs[i] != id {
						t.Errorf("ID[%d] mismatch: got %d, want %d", i, got.IDs[i], id)
					}
				}
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			var idset *ns.IDSet
			if tc.isTag {
				idset = &ns.IDSet{IsTag: true, TagName: tc.tagName}
			} else {
				idset = &ns.IDSet{IDs: tc.ids}
			}
			buf := ns.NewWriter()
			if err := idset.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// XOrY wire format:
//   Boolean isX
//   X or Y value

func TestXOrY(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.XOrY[ns.VarInt, ns.String]
		isX      bool
		xVal     ns.VarInt
		yVal     ns.String
	}{
		{
			name:     "X value",
			raw:      []byte{0x01, 0x2a},
			expected: ns.XOrY[ns.VarInt, ns.String]{IsX: true, X: 42},
			isX:      true,
			xVal:     42,
		},
		{
			name:     "Y value",
			raw:      []byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o'},
			expected: ns.XOrY[ns.VarInt, ns.String]{Y: "hello"},
			isX:      false,
			yVal:     "hello",
		},
	}

	decodeX := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	decodeY := func(buf *ns.PacketBuffer) (ns.String, error) { return buf.ReadString(32767) }
	encodeX := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }
	encodeY := func(buf *ns.PacketBuffer, v ns.String) error { return buf.WriteString(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.XOrY[ns.VarInt, ns.String]
			if err := got.DecodeWith(ns.NewReader(tc.raw), decodeX, decodeY); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsX != tc.isX {
				t.Errorf("IsX mismatch: got %v, want %v", got.IsX, tc.isX)
			}
			if tc.isX && got.X != tc.xVal {
				t.Errorf("X mismatch: got %d, want %d", got.X, tc.xVal)
			}
			if !tc.isX && got.Y != tc.yVal {
				t.Errorf("Y mismatch: got %q, want %q", got.Y, tc.yVal)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.expected.EncodeWith(buf, encodeX, encodeY); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

// IDOrX wire format:
//   VarInt id (0 = inline, >0 = registry id + 1)
//   T value (if id = 0)

func TestIDOrX(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected ns.IDOrX[ns.VarInt]
	}{
		{
			name:     "registry reference",
			raw:      []byte{0x2b},
			expected: ns.IDOrX[ns.VarInt]{ID: 42},
		},
		{
			name:     "inline value",
			raw:      []byte{0x00, 0x64},
			expected: ns.IDOrX[ns.VarInt]{IsInline: true, Value: 100},
		},
	}

	decoder := func(buf *ns.PacketBuffer) (ns.VarInt, error) { return buf.ReadVarInt() }
	encoder := func(buf *ns.PacketBuffer, v ns.VarInt) error { return buf.WriteVarInt(v) }

	for _, tc := range testCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.IDOrX[ns.VarInt]
			if err := got.DecodeWith(ns.NewReader(tc.raw), decoder); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.IsInline != tc.expected.IsInline {
				t.Errorf("IsInline mismatch: got %v, want %v", got.IsInline, tc.expected.IsInline)
			}
			if got.IsInline {
				if got.Value != tc.expected.Value {
					t.Errorf("Value mismatch: got %d, want %d", got.Value, tc.expected.Value)
				}
			} else {
				if got.ID != tc.expected.ID {
					t.Errorf("ID mismatch: got %d, want %d", got.ID, tc.expected.ID)
				}
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
