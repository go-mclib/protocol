package net_structures

import (
	"bytes"
	"testing"

	"github.com/go-mclib/protocol/nbt"
)

func TestTextComponent_SimpleText(t *testing.T) {
	tc := NewTextComponent("Hello, World!")

	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(buf.Bytes())
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Text != tc.Text {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
	}
}

func TestTextComponent_WithStyle(t *testing.T) {
	bold := true
	italic := false
	tc := TextComponent{
		Text:   "Styled text",
		Color:  "red",
		Bold:   &bold,
		Italic: &italic,
	}

	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(buf.Bytes())
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Text != tc.Text {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
	}
	if decoded.Color != tc.Color {
		t.Errorf("Color mismatch: got %q, want %q", decoded.Color, tc.Color)
	}
	if decoded.Bold == nil || *decoded.Bold != *tc.Bold {
		t.Errorf("Bold mismatch")
	}
	if decoded.Italic == nil || *decoded.Italic != *tc.Italic {
		t.Errorf("Italic mismatch")
	}
}

func TestTextComponent_WithExtra(t *testing.T) {
	tc := TextComponent{
		Text: "Hello, ",
		Extra: []TextComponent{
			{Text: "World", Color: "gold"},
			{Text: "!"},
		},
	}

	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(buf.Bytes())
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Text != tc.Text {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
	}
	if len(decoded.Extra) != len(tc.Extra) {
		t.Fatalf("Extra length mismatch: got %d, want %d", len(decoded.Extra), len(tc.Extra))
	}
	if decoded.Extra[0].Text != "World" {
		t.Errorf("Extra[0].Text mismatch: got %q, want %q", decoded.Extra[0].Text, "World")
	}
	if decoded.Extra[0].Color != "gold" {
		t.Errorf("Extra[0].Color mismatch: got %q, want %q", decoded.Extra[0].Color, "gold")
	}
}

func TestTextComponent_Translate(t *testing.T) {
	tc := NewTranslateComponent("chat.type.text",
		NewTextComponent("Player"),
		NewTextComponent("Hello"),
	)

	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(buf.Bytes())
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Translate != tc.Translate {
		t.Errorf("Translate mismatch: got %q, want %q", decoded.Translate, tc.Translate)
	}
	if len(decoded.With) != len(tc.With) {
		t.Fatalf("With length mismatch: got %d, want %d", len(decoded.With), len(tc.With))
	}
}

func TestTextComponent_PlainStringShorthand(t *testing.T) {
	// encode a plain string directly as NBT String type
	data, err := nbt.MarshalNetwork(nbt.String("Plain text"))
	if err != nil {
		t.Fatalf("MarshalNetwork failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(data)
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Text != "Plain text" {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, "Plain text")
	}
}

func TestTextComponent_PacketBufferHelpers(t *testing.T) {
	tc := NewTextComponent("Test message")

	buf := NewWriter()
	if err := buf.WriteTextComponent(tc); err != nil {
		t.Fatalf("WriteTextComponent failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadTextComponent()
	if err != nil {
		t.Fatalf("ReadTextComponent failed: %v", err)
	}

	if decoded.Text != tc.Text {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
	}
}

func TestTextComponent_ClickEvent(t *testing.T) {
	tc := TextComponent{
		Text: "Click me",
		ClickEvent: &ClickEvent{
			Action: "open_url",
			Value:  "https://minecraft.net",
		},
	}

	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded TextComponent
	readBuf := NewReader(buf.Bytes())
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.ClickEvent == nil {
		t.Fatal("ClickEvent is nil")
	}
	if decoded.ClickEvent.Action != tc.ClickEvent.Action {
		t.Errorf("ClickEvent.Action mismatch: got %q, want %q", decoded.ClickEvent.Action, tc.ClickEvent.Action)
	}
	if decoded.ClickEvent.Value != tc.ClickEvent.Value {
		t.Errorf("ClickEvent.Value mismatch: got %q, want %q", decoded.ClickEvent.Value, tc.ClickEvent.Value)
	}
}

func TestTextComponent_RoundTrip(t *testing.T) {
	bold := true
	tc := TextComponent{
		Text:      "Complex",
		Color:     "#FF5555",
		Bold:      &bold,
		Font:      "minecraft:default",
		Insertion: "inserted text",
		Extra: []TextComponent{
			{Text: " component", Color: "aqua"},
		},
		ClickEvent: &ClickEvent{
			Action: "run_command",
			Value:  "/say hello",
		},
	}

	// encode
	buf := NewWriter()
	if err := tc.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	encoded := buf.Bytes()

	// decode
	var decoded TextComponent
	readBuf := NewReader(encoded)
	if err := decoded.Decode(readBuf); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// re-encode
	buf2 := NewWriter()
	if err := decoded.Encode(buf2); err != nil {
		t.Fatalf("Re-encode failed: %v", err)
	}
	reencoded := buf2.Bytes()

	// compare
	if !bytes.Equal(encoded, reencoded) {
		t.Errorf("Round-trip encoding mismatch:\n  original:  %x\n  reencoded: %x", encoded, reencoded)
	}
}

func TestTextComponent_StringTagOptimization(t *testing.T) {
	// simple text should use NBT String tag (type 0x08)
	simple := NewTextComponent("Hello")
	buf := NewWriter()
	if err := simple.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	simpleBytes := buf.Bytes()

	// first byte should be NBT String tag type (0x08)
	if simpleBytes[0] != 0x08 {
		t.Errorf("Simple text should use String tag (0x08), got 0x%02x", simpleBytes[0])
	}

	// styled text should use NBT Compound tag (type 0x0a)
	styled := TextComponent{Text: "Hello", Color: "red"}
	buf2 := NewWriter()
	if err := styled.Encode(buf2); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	styledBytes := buf2.Bytes()

	// first byte should be NBT Compound tag type (0x0a)
	if styledBytes[0] != 0x0a {
		t.Errorf("Styled text should use Compound tag (0x0a), got 0x%02x", styledBytes[0])
	}

	// simple encoding should be smaller
	if len(simpleBytes) >= len(styledBytes) {
		t.Errorf("Simple text (%d bytes) should be smaller than styled (%d bytes)",
			len(simpleBytes), len(styledBytes))
	}

	// verify both decode correctly
	var decoded1, decoded2 TextComponent
	if err := decoded1.Decode(NewReader(simpleBytes)); err != nil {
		t.Fatalf("Decode simple failed: %v", err)
	}
	if err := decoded2.Decode(NewReader(styledBytes)); err != nil {
		t.Fatalf("Decode styled failed: %v", err)
	}

	if decoded1.Text != "Hello" {
		t.Errorf("Simple decoded text mismatch: %q", decoded1.Text)
	}
	if decoded2.Text != "Hello" || decoded2.Color != "red" {
		t.Errorf("Styled decoded mismatch: text=%q color=%q", decoded2.Text, decoded2.Color)
	}
}
