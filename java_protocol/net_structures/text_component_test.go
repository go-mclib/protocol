package net_structures_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/go-mclib/nbt"
	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

// TextComponent wire format: NBT tag (network format, no root name)
//   - Simple text: NBT String tag (0x08) + length + UTF-8 bytes
//   - Complex: NBT Compound tag (0x0a) + fields + end tag (0x00)
//
// Reference: https://minecraft.wiki/w/Java_Edition_protocol/Data_types#Text_Component

var textComponentTestCases = []struct {
	name string
	raw  []byte
	text string
}{
	{
		name: "simple text",
		raw:  []byte{0x08, 0x00, 0x05, 'H', 'e', 'l', 'l', 'o'},
		text: "Hello",
	},
	{
		name: "world",
		raw:  []byte{0x08, 0x00, 0x05, 'W', 'o', 'r', 'l', 'd'},
		text: "World",
	},
}

func TestTextComponent(t *testing.T) {
	for _, tc := range textComponentTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.TextComponent
			buf := ns.NewReader(tc.raw)
			if err := got.Decode(buf); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.Text != tc.text {
				t.Errorf("Text mismatch: got %q, want %q", got.Text, tc.text)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			comp := ns.TextComponent{Text: tc.text}
			buf := ns.NewWriter()
			if err := comp.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestTextComponent_Complex(t *testing.T) {
	cases := []ns.TextComponent{
		{Text: "Styled", Color: "red"},
		{Text: "Hello, ", Extra: []ns.TextComponent{{Text: "World", Color: "gold"}}},
		{Translate: "chat.type.text", With: []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}}},
		{Text: "Click me", ClickEvent: &ns.ClickEvent{Action: "open_url", URL: "https://minecraft.net"}},
	}

	for _, tc := range cases {
		name := tc.Text
		if name == "" {
			name = tc.Translate
		}
		t.Run(name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}

			var decoded ns.TextComponent
			if err := decoded.Decode(ns.NewReader(buf.Bytes())); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if decoded.Text != tc.Text {
				t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
			}
			if decoded.Color != tc.Color {
				t.Errorf("Color mismatch: got %q, want %q", decoded.Color, tc.Color)
			}
			if decoded.Translate != tc.Translate {
				t.Errorf("Translate mismatch: got %q, want %q", decoded.Translate, tc.Translate)
			}
			if len(decoded.Extra) != len(tc.Extra) {
				t.Errorf("Extra length mismatch: got %d, want %d", len(decoded.Extra), len(tc.Extra))
			}
			if len(decoded.With) != len(tc.With) {
				t.Errorf("With length mismatch: got %d, want %d", len(decoded.With), len(tc.With))
			}

			// re-encode should match
			buf2 := ns.NewWriter()
			if err := decoded.Encode(buf2); err != nil {
				t.Fatalf("re-encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
				t.Errorf("round-trip mismatch:\n  original:   %x\n  re-encoded: %x", buf.Bytes(), buf2.Bytes())
			}
		})
	}
}

func TestTextComponent_StringOptimization(t *testing.T) {
	simple := ns.NewTextComponent("Hello")
	styled := ns.TextComponent{Text: "Hello", Color: "red"}

	simpleBuf := ns.NewWriter()
	simple.Encode(simpleBuf)

	styledBuf := ns.NewWriter()
	styled.Encode(styledBuf)

	if simpleBuf.Bytes()[0] != 0x08 {
		t.Errorf("simple text should use String tag (0x08), got 0x%02x", simpleBuf.Bytes()[0])
	}
	if styledBuf.Bytes()[0] != 0x0a {
		t.Errorf("styled text should use Compound tag (0x0a), got 0x%02x", styledBuf.Bytes()[0])
	}
	if simpleBuf.Len() >= styledBuf.Len() {
		t.Errorf("simple (%d bytes) should be smaller than styled (%d bytes)", simpleBuf.Len(), styledBuf.Len())
	}
}

func TestTextComponent_MixedListWirePrefix(t *testing.T) {
	// Vanilla ListTag wraps the first literal "[" in a singleton compound whose
	// only key is empty. The other list entry is already a styled compound.
	// This raw network NBT is independent of TextComponent.Encode.
	raw := []byte("\x0a" +
		"\x08\x00\x04text\x00\x00" +
		"\x09\x00\x05extra\x0a\x00\x00\x00\x02" +
		"\x08\x00\x00\x00\x01[\x00" +
		"\x08\x00\x04text\x00\x19Discord] Eragon » Aaaaaa" +
		"\x08\x00\x05color\x00\x04blue\x00\x00")
	var tc ns.TextComponent
	if err := tc.Decode(ns.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	if got := tc.String(); got != "[Discord] Eragon » Aaaaaa" {
		t.Errorf("mixed-list prefix = %q", got)
	}
}

func TestTextComponent_NBTListsAndArguments(t *testing.T) {
	tests := []struct {
		name string
		tag  nbt.Tag
		want string
	}{
		{
			name: "root list retaining first extras",
			tag: nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
				nbt.Compound{
					"text": nbt.String("["), "color": nbt.String("blue"),
					"extra": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{nbt.String("Dis")}},
				},
				nbt.Compound{"": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("cord] "), nbt.String("Eragon"),
				}}},
				nbt.Compound{"text": nbt.String(" » Aaaaaa")},
			}},
			want: "[Discord] Eragon » Aaaaaa",
		},
		{
			name: "fallback with mixed primitive and component arguments",
			tag: nbt.Compound{
				"translate": nbt.String("plugin.discord"),
				"fallback":  nbt.String("[Discord] %1$s » %2$s (%3$s)"),
				"with": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{"": nbt.String("Eragon")},
					nbt.Compound{"text": nbt.String("Aaaaaa")},
					nbt.Compound{"": nbt.Int(42)},
				}},
			},
			want: "[Discord] Eragon » Aaaaaa (42)",
		},
		{
			name: "primitive argument array",
			tag: nbt.Compound{
				"translate": nbt.String("plugin.numbers"),
				"fallback":  nbt.String("%2$s %1$s"),
				"with":      nbt.IntArray{7, 42},
			},
			want: "42 7",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := nbt.EncodeNetwork(test.tag)
			if err != nil {
				t.Fatal(err)
			}
			var tc ns.TextComponent
			if err := tc.Decode(ns.NewReader(raw)); err != nil {
				t.Fatal(err)
			}
			if got := tc.String(); got != test.want {
				t.Errorf("String = %q, want %q", got, test.want)
			}
			if test.name == "root list retaining first extras" {
				if got := tc.MiniMessage(); got != "<blue>[Discord] Eragon » Aaaaaa</blue>" {
					t.Errorf("inherited root style = %q", got)
				}
			}
		})
	}
}

func TestTextComponent_RejectMalformedNBT(t *testing.T) {
	for name, tag := range map[string]nbt.Tag{
		"empty list":                      nbt.List{ElementType: nbt.TagEnd},
		"empty compound":                  nbt.Compound{},
		"number":                          nbt.Int(42),
		"numeric component list":          nbt.List{ElementType: nbt.TagInt, Elements: []nbt.Tag{nbt.Int(1)}},
		"root wrapper is not a component": nbt.Compound{"": nbt.String("[")},
		"double wrapper must not flatten": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
			nbt.Compound{"": nbt.Compound{"": nbt.String("[")}},
		}},
		"wrapper with other fields": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
			nbt.Compound{"": nbt.String("["), "color": nbt.String("blue")},
		}},
		"wrong text type":     nbt.Compound{"text": nbt.Int(1)},
		"empty extra":         nbt.Compound{"text": nbt.String(""), "extra": nbt.List{ElementType: nbt.TagEnd}},
		"non-list arguments":  nbt.Compound{"translate": nbt.String("x"), "with": nbt.String("bad")},
		"wrong discriminator": nbt.Compound{"type": nbt.String("translatable"), "text": nbt.String("bad")},
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := nbt.EncodeNetwork(tag)
			if err != nil {
				t.Fatal(err)
			}
			tc := ns.TextComponent{Text: "unchanged"}
			if err := tc.Decode(ns.NewReader(raw)); err == nil {
				t.Fatal("accepted malformed component")
			}
			if tc.String() != "unchanged" {
				t.Errorf("failed decode replaced previous component: %q", tc.String())
			}
		})
	}
}

func TestNestedEmptyComponentParentsRoundTrip(t *testing.T) {
	original := ns.TextComponent{Extra: []ns.TextComponent{
		{Color: "gold", Extra: []ns.TextComponent{{Text: "["}}},
		{Text: "Discord] Eragon » Aaaaaa"},
	}}
	for _, format := range []string{"nbt", "json"} {
		t.Run(format, func(t *testing.T) {
			var decoded ns.TextComponent
			if format == "nbt" {
				writer := ns.NewWriter()
				if err := original.Encode(writer); err != nil {
					t.Fatal(err)
				}
				if err := decoded.Decode(ns.NewReader(writer.Bytes())); err != nil {
					t.Fatal(err)
				}
			} else {
				raw, err := json.Marshal(original)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(raw, &decoded); err != nil {
					t.Fatal(err)
				}
			}
			if got := decoded.MiniMessage(); got != "<gold>[</gold>Discord] Eragon » Aaaaaa" {
				t.Fatalf("nested parent lost text or inherited style: %q", got)
			}
		})
	}
}
