package wire_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-mclib/nbt"
	"github.com/go-mclib/protocol/javaproto/textformat"
	ns "github.com/go-mclib/protocol/javaproto/wire"
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
				if got := textformat.MiniMessage(tc); got != "<blue>[Discord] Eragon » Aaaaaa</blue>" {
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
			if got := textformat.MiniMessage(decoded); got != "<gold>[</gold>Discord] Eragon » Aaaaaa" {
				t.Fatalf("nested parent lost text or inherited style: %q", got)
			}
		})
	}
}

func TestVanillaPlayerHeadChatComponent(t *testing.T) {
	// Encoded by vanilla 26.2 ComponentSerialization.CODEC + NbtIo.writeAnyTag.
	raw, err := base64.StdEncoding.DecodeString("CgkABWV4dHJhCgAAAAQIAAAACVtNZW1iZXJdIAAIAAAAASAACgAGcGxheWVyCAAEbmFtZQAFQWxpY2UAAAgAAAATIEFsaWNlIMK7IGhleSBncmFjZQAIAAR0ZXh0AAAA")
	if err != nil {
		t.Fatal(err)
	}
	var component ns.TextComponent
	if err := component.Decode(ns.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	if got := component.String(); got != "[Member]  [Alice head] Alice » hey grace" {
		t.Fatalf("vanilla player-head chat = %q", got)
	}
}

func TestObjectComponentsRoundTrip(t *testing.T) {
	for _, test := range []struct {
		name, json, plain, mini string
	}{
		{"player_name", `{"player":"Alice"}`, "[Alice head]", "[Alice head]"},
		{"anonymous_profile", `{"type":"object","object":"player","player":{"id":[1,-2,3,-4],"properties":[{"name":"textures","value":"skin","signature":"signed"}]},"hat":false}`, "[unknown player head]", "[unknown player head]"},
		{"block_atlas", `{"sprite":"minecraft:block/stone"}`, "[block/stone]", "[block/stone]"},
		{"custom_atlas", `{"object":"atlas","sprite":"plugin:badge","atlas":"minecraft:gui"}`, "[plugin:badge@gui]", "[plugin:badge@gui]"},
		{"styled_fallback", `{"player":{},"fallback":{"text":"","extra":[{"text":"icon","color":"gold"}]},"color":"aqua","extra":[" tail"]}`, "icon tail", "<aqua><gold>icon</gold> tail</aqua>"},
		{"nested_object_fallback", `{"sprite":"minecraft:block/stone","fallback":{"player":"Alice","fallback":""}}`, "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var component ns.TextComponent
			if err := json.Unmarshal([]byte(test.json), &component); err != nil {
				t.Fatal(err)
			}
			for _, format := range []string{"nbt", "json"} {
				var roundTrip ns.TextComponent
				if format == "nbt" {
					buf := ns.NewWriter()
					if err := component.Encode(buf); err != nil {
						t.Fatal(err)
					}
					if err := roundTrip.Decode(ns.NewReader(buf.Bytes())); err != nil {
						t.Fatal(err)
					}
				} else {
					data, err := json.Marshal(component)
					if err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(data, &roundTrip); err != nil {
						t.Fatalf("%s: %v", data, err)
					}
				}
				if got := roundTrip.String(); got != test.plain {
					t.Fatalf("%s plain = %q, want %q", format, got, test.plain)
				}
				if got := textformat.MiniMessage(roundTrip); got != test.mini {
					t.Fatalf("%s MiniMessage = %q, want %q", format, got, test.mini)
				}
				if test.name == "anonymous_profile" {
					profile := roundTrip.Object.Player.(map[string]any)
					id := profile["id"].([]int32)
					properties := profile["properties"].([]any)
					texture := properties[0].(map[string]any)
					if id[1] != -2 || id[3] != -4 || texture["value"] != "skin" || texture["signature"] != "signed" || roundTrip.Object.Hat == nil || *roundTrip.Object.Hat {
						t.Fatalf("%s lost profile/hat metadata: %+v", format, roundTrip.Object)
					}
				}
			}
		})
	}
}

func TestInvalidObjectComponents(t *testing.T) {
	for _, input := range []string{
		`{"type":"object"}`,
		`{"object":"unknown","player":"Alice"}`,
		`{"object":"atlas","player":"Alice"}`,
		`{"object":"player","sprite":"minecraft:block/stone"}`,
		`{"player":42}`,
		`{"player":{"id":[1,2,3]}}`,
		`{"player":{"id":[1,2,3,4.5]}}`,
		`{"player":"Alice","fallback":{}}`,
	} {
		var component ns.TextComponent
		if err := json.Unmarshal([]byte(input), &component); err == nil {
			t.Errorf("accepted invalid object component %s", input)
		}
	}
}

func TestTextComponent_MarshalNBT(t *testing.T) {
	bold := true
	cases := []struct {
		name string
		tc   ns.TextComponent
	}{
		{"plain", ns.NewTextComponent("hello")},
		{"styled", ns.TextComponent{Text: "hi", Color: "red", Bold: &bold}},
		{"style only", ns.TextComponent{Color: "gold"}},
		{"children", ns.TextComponent{Text: "a", Extra: []ns.TextComponent{{Text: "b"}, {Color: "blue"}}}},
		{"translate", ns.NewTranslateComponent("k", ns.NewTextComponent("x"))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tag, err := nbt.MarshalTag(c.tc)
			if err != nil {
				t.Fatal(err)
			}
			buf := ns.NewWriter()
			if err := c.tc.Encode(buf); err != nil {
				t.Fatal(err)
			}
			want, err := nbt.EncodeNetwork(tag)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf.Bytes(), want) {
				t.Errorf("MarshalTag wire %x != Encode %x", want, buf.Bytes())
			}
			ptrTag, err := nbt.MarshalTag(&c.tc)
			if err != nil || !reflect.DeepEqual(ptrTag, tag) {
				t.Errorf("pointer marshal %#v (%v) != %#v", ptrTag, err, tag)
			}
			nested, err := nbt.MarshalTag(struct {
				C ns.TextComponent `nbt:"c"`
			}{c.tc})
			if err != nil || !reflect.DeepEqual(nested.(nbt.Compound)["c"], tag) {
				t.Errorf("nested marshal %#v (%v) != %#v", nested, err, tag)
			}
			var back ns.TextComponent
			if err := back.UnmarshalNBT(tag); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(back, c.tc) {
				t.Errorf("round-trip got %#v, want %#v", back, c.tc)
			}
		})
	}
}
