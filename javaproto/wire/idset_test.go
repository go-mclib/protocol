package wire_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
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
