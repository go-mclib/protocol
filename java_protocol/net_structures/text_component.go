package net_structures

import (
	"github.com/go-mclib/protocol/nbt"
)

// TextComponent represents a Minecraft text component.
// Encoded as NBT over the network (since 1.20.3+).
//
// A text component can be:
//   - A plain string (text content only)
//   - A compound with content, style, and children
//
// Wire format: NBT (network format, nameless root)
type TextComponent struct {
	// content types (only one should be set)
	Text       string `nbt:"text,omitempty"`
	Translate  string `nbt:"translate,omitempty"`
	Keybind    string `nbt:"keybind,omitempty"`
	Score      *Score `nbt:"score,omitempty"`
	Selector   string `nbt:"selector,omitempty"`
	NBT        string `nbt:"nbt,omitempty"`
	NBTBlock   string `nbt:"block,omitempty"`     // for nbt content type
	NBTEntity  string `nbt:"entity,omitempty"`    // for nbt content type
	NBTStorage string `nbt:"storage,omitempty"`   // for nbt content type
	Interpret  *bool  `nbt:"interpret,omitempty"` // for nbt content type

	// translation arguments (for translate content type)
	With []TextComponent `nbt:"with,omitempty"`

	// style
	Color         string `nbt:"color,omitempty"`
	Bold          *bool  `nbt:"bold,omitempty"`
	Italic        *bool  `nbt:"italic,omitempty"`
	Underlined    *bool  `nbt:"underlined,omitempty"`
	Strikethrough *bool  `nbt:"strikethrough,omitempty"`
	Obfuscated    *bool  `nbt:"obfuscated,omitempty"`
	Font          string `nbt:"font,omitempty"`
	Insertion     string `nbt:"insertion,omitempty"`

	// click/hover events
	ClickEvent *ClickEvent `nbt:"clickEvent,omitempty"`
	HoverEvent *HoverEvent `nbt:"hoverEvent,omitempty"`

	// children
	Extra []TextComponent `nbt:"extra,omitempty"`
}

// Score represents score component content.
type Score struct {
	Name      string `nbt:"name"`
	Objective string `nbt:"objective"`
}

// ClickEvent represents a click event for text components.
type ClickEvent struct {
	Action string `nbt:"action"`
	Value  string `nbt:"value"`
}

// HoverEvent represents a hover event for text components.
type HoverEvent struct {
	Action   string `nbt:"action"`
	Contents any    `nbt:"contents,omitempty"`
}

// NewTextComponent creates a simple text component with the given text.
func NewTextComponent(text string) TextComponent {
	return TextComponent{Text: text}
}

// NewTranslateComponent creates a translatable text component.
func NewTranslateComponent(key string, args ...TextComponent) TextComponent {
	return TextComponent{Translate: key, With: args}
}

// isSimpleText returns true if this component contains only plain text
// with no styling, events, or children.
func (tc *TextComponent) isSimpleText() bool {
	return tc.Text != "" &&
		tc.Translate == "" &&
		tc.Keybind == "" &&
		tc.Score == nil &&
		tc.Selector == "" &&
		tc.NBT == "" &&
		tc.NBTBlock == "" &&
		tc.NBTEntity == "" &&
		tc.NBTStorage == "" &&
		tc.Interpret == nil &&
		len(tc.With) == 0 &&
		tc.Color == "" &&
		tc.Bold == nil &&
		tc.Italic == nil &&
		tc.Underlined == nil &&
		tc.Strikethrough == nil &&
		tc.Obfuscated == nil &&
		tc.Font == "" &&
		tc.Insertion == "" &&
		tc.ClickEvent == nil &&
		tc.HoverEvent == nil &&
		len(tc.Extra) == 0
}

// Encode writes the text component as NBT to the writer.
// Simple text-only components are encoded as NBT String tags for efficiency.
func (tc *TextComponent) Encode(buf *PacketBuffer) error {
	var data []byte
	var err error

	if tc.isSimpleText() {
		// encode as NBT String tag (more compact, less data sent over network)
		data, err = nbt.Encode(nbt.String(tc.Text), "", true)
	} else {
		// encode as NBT Compound tag
		data, err = nbt.MarshalNetwork(tc)
	}

	if err != nil {
		return err
	}
	_, err = buf.Write(data)
	return err
}

// Decode reads a text component from NBT.
func (tc *TextComponent) Decode(buf *PacketBuffer) error {
	nbtReader := nbt.NewReaderFrom(buf.Reader())
	tag, _, err := nbtReader.ReadTag(true)
	if err != nil {
		return err
	}

	// handle plain string shorthand
	if s, ok := tag.(nbt.String); ok {
		*tc = TextComponent{Text: string(s)}
		return nil
	}

	return nbt.UnmarshalTag(tag, tc)
}

// ReadTextComponent reads a text component from the buffer.
func (pb *PacketBuffer) ReadTextComponent() (TextComponent, error) {
	var tc TextComponent
	err := tc.Decode(pb)
	return tc, err
}

// WriteTextComponent writes a text component to the buffer.
func (pb *PacketBuffer) WriteTextComponent(tc TextComponent) error {
	return tc.Encode(pb)
}
