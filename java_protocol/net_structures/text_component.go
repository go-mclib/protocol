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

	// click/hover events (1.21.5+ uses snake_case field names)
	ClickEvent *ClickEvent `nbt:"click_event,omitempty"`
	HoverEvent *HoverEvent `nbt:"hover_event,omitempty"`

	// children
	Extra []TextComponent `nbt:"extra,omitempty"`
}

// Score represents score component content.
type Score struct {
	Name      string `nbt:"name"`
	Objective string `nbt:"objective"`
}

// ClickEvent represents a click event for text components (1.21.5+ format).
// Each action type uses a different field; the Action field determines which is relevant.
type ClickEvent struct {
	Action  string `nbt:"action"`
	URL     string `nbt:"url,omitempty"`     // open_url
	Path    string `nbt:"path,omitempty"`    // open_file
	Command string `nbt:"command,omitempty"` // run_command, suggest_command
	Page    int32  `nbt:"page,omitempty"`    // change_page
	Value   string `nbt:"value,omitempty"`   // copy_to_clipboard
	Dialog  any    `nbt:"dialog,omitempty"`  // show_dialog
	ID      string `nbt:"id,omitempty"`      // custom
	Payload any    `nbt:"payload,omitempty"` // custom
}

// HoverEvent represents a hover event for text components (1.21.5+ format).
// Each action type uses different fields; the Action field determines which are relevant.
type HoverEvent struct {
	Action string `nbt:"action"`
	// show_text
	Value any `nbt:"value,omitempty"` // TextComponent (string or compound NBT)
	// show_entity and show_item
	ID string `nbt:"id,omitempty"` // entity type or item ID
	// show_entity
	EntityUUID any `nbt:"uuid,omitempty"` // IntArray in NBT
	Name       any `nbt:"name,omitempty"` // optional TextComponent (string or compound NBT)
	// show_item
	Count      int32 `nbt:"count,omitempty"`
	Components any   `nbt:"components,omitempty"` // item components compound
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

// UnmarshalNBT implements nbt.TagUnmarshaler, allowing TextComponent to be
// correctly unmarshaled from both NBT String (plain text shorthand) and Compound tags.
func (tc *TextComponent) UnmarshalNBT(tag nbt.Tag) error {
	if s, ok := tag.(nbt.String); ok {
		*tc = TextComponent{Text: string(s)}
		return nil
	}
	// use type alias to avoid infinite recursion through TagUnmarshaler
	type plain TextComponent
	return nbt.UnmarshalTag(tag, (*plain)(tc))
}

// Decode reads a text component from NBT.
func (tc *TextComponent) Decode(buf *PacketBuffer) error {
	nbtReader := nbt.NewReaderFrom(buf.Reader())
	tag, _, err := nbtReader.ReadTag(true)
	if err != nil {
		return err
	}
	return tc.UnmarshalNBT(tag)
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
