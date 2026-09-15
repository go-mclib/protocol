package net_structures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-mclib/nbt"
)

// TextComponent represents a Minecraft text component.
// Encoded as NBT over the network (since 1.20.3+).
//
// A text component can be:
//   - A plain string (text content only)
//   - A compound with content, style, and children
//   - A non-empty list, whose first component becomes the parent of the rest
//
// Wire format: NBT (network format, nameless root)
type TextComponent struct {
	// content types (only one should be set)
	Text       string  `nbt:"text,omitempty" json:"text,omitempty"`
	Translate  string  `nbt:"translate,omitempty" json:"translate,omitempty"`
	Fallback   *string `nbt:"fallback,omitempty" json:"fallback,omitempty"`
	Keybind    string  `nbt:"keybind,omitempty" json:"keybind,omitempty"`
	Score      *Score  `nbt:"score,omitempty" json:"score,omitempty"`
	Selector   string  `nbt:"selector,omitempty" json:"selector,omitempty"`
	NBT        string  `nbt:"nbt,omitempty" json:"nbt,omitempty"`
	NBTBlock   string  `nbt:"block,omitempty" json:"block,omitempty"`         // for nbt content type
	NBTEntity  string  `nbt:"entity,omitempty" json:"entity,omitempty"`       // for nbt content type
	NBTStorage string  `nbt:"storage,omitempty" json:"storage,omitempty"`     // for nbt content type
	Interpret  *bool   `nbt:"interpret,omitempty" json:"interpret,omitempty"` // for nbt content type

	// translation arguments (for translate content type)
	With []TextComponent `nbt:"with,omitempty" json:"with,omitempty"`

	// style
	Color         string `nbt:"color,omitempty" json:"color,omitempty"`
	Bold          *bool  `nbt:"bold,omitempty" json:"bold,omitempty"`
	Italic        *bool  `nbt:"italic,omitempty" json:"italic,omitempty"`
	Underlined    *bool  `nbt:"underlined,omitempty" json:"underlined,omitempty"`
	Strikethrough *bool  `nbt:"strikethrough,omitempty" json:"strikethrough,omitempty"`
	Obfuscated    *bool  `nbt:"obfuscated,omitempty" json:"obfuscated,omitempty"`
	Font          string `nbt:"font,omitempty" json:"font,omitempty"`
	Insertion     string `nbt:"insertion,omitempty" json:"insertion,omitempty"`

	// click/hover events
	ClickEvent *ClickEvent `nbt:"click_event,omitempty" json:"clickEvent,omitempty"`
	HoverEvent *HoverEvent `nbt:"hover_event,omitempty" json:"hoverEvent,omitempty"`

	// children
	Extra []TextComponent `nbt:"extra,omitempty" json:"extra,omitempty"`
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
		tc.Fallback == nil &&
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

// hasContentType returns true if any content type field is non-empty.
func (tc *TextComponent) hasContentType() bool {
	return tc.Text != "" || tc.Translate != "" || tc.Keybind != "" ||
		tc.Score != nil || tc.Selector != "" || tc.NBT != ""
}

// Encode writes the text component as NBT to the writer.
// Simple text-only components are encoded as NBT String tags for efficiency.
func (tc *TextComponent) Encode(buf *PacketBuffer) error {
	var data []byte
	var err error

	if tc.isSimpleText() {
		data, err = nbt.Encode(nbt.String(tc.Text), "", true)
	} else {
		// Nested extra/with parents need an explicit content type too.
		var tag nbt.Tag
		tag, err = nbt.MarshalTag(tc)
		if err == nil {
			normalizeComponentTag(tag)
			data, err = nbt.EncodeNetwork(tag)
		}
	}

	if err != nil {
		return err
	}
	_, err = buf.Write(data)
	return err
}

// normalizeComponentTag restores empty literal content omitted by reflection.
// Only component-valued fields are traversed; event payloads remain untouched.
func normalizeComponentTag(tag nbt.Tag) {
	switch value := tag.(type) {
	case nbt.List:
		for _, child := range value.Elements {
			normalizeComponentTag(child)
		}
	case nbt.Compound:
		hasContent := false
		for _, key := range [...]string{"text", "translate", "keybind", "score", "selector", "nbt"} {
			if _, ok := value[key]; ok {
				hasContent = true
				break
			}
		}
		if !hasContent {
			value["text"] = nbt.String("")
		}
		normalizeComponentTag(value["extra"])
		normalizeComponentTag(value["with"])
		if hover, ok := value["hover_event"].(nbt.Compound); ok {
			switch hover["action"] {
			case nbt.String("show_text"):
				normalizeComponentTag(hover["value"])
			case nbt.String("show_entity"):
				normalizeComponentTag(hover["name"])
			}
		}
	}
}

// MarshalJSON retains the empty literal content of styled or parent components.
func (tc TextComponent) MarshalJSON() ([]byte, error) {
	type component TextComponent
	if tc.hasContentType() {
		return json.Marshal(component(tc))
	}
	return json.Marshal(struct {
		Text string `json:"text"`
		component
	}{component: component(tc)})
}

// UnmarshalJSON decodes strings, component objects, and non-empty component lists.
// Lists retain the first component's style and append after its existing extras.
func (tc *TextComponent) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("empty text component")
	}
	var decoded TextComponent
	switch data[0] {
	case '"':
		if err := json.Unmarshal(data, &decoded.Text); err != nil {
			return err
		}
	case '[':
		var list []TextComponent
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		if len(list) == 0 {
			return fmt.Errorf("empty component list")
		}
		decoded = list[0]
		decoded.Extra = append(decoded.Extra, list[1:]...)
	case '{':
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		if err := validateComponentContent(func(key string) bool {
			value, ok := fields[key]
			return ok && (key == "type" || !bytes.Equal(bytes.TrimSpace(value), []byte("null")))
		}, func() (string, error) {
			var kind string
			err := json.Unmarshal(fields["type"], &kind)
			return kind, err
		}); err != nil {
			return err
		}
		type plain TextComponent
		value := struct {
			*plain
			With  []json.RawMessage `json:"with"`
			Extra json.RawMessage   `json:"extra"`
		}{plain: (*plain)(&decoded)}
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		if value.Extra != nil {
			if err := json.Unmarshal(value.Extra, &decoded.Extra); err != nil {
				return err
			}
			if len(decoded.Extra) == 0 {
				return fmt.Errorf("empty component extra list")
			}
		}
		if raw, ok := fields["with"]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("null translation arguments")
		}
		if len(value.With) != 0 {
			decoded.With = make([]TextComponent, len(value.With))
			for i, raw := range value.With {
				raw = bytes.TrimSpace(raw)
				if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) ||
					(len(raw) > 0 && (raw[0] == '-' || raw[0] >= '0' && raw[0] <= '9')) {
					// Vanilla accepts primitive arguments, but not primitive components.
					decoded.With[i].Text = string(raw)
				} else if err := json.Unmarshal(raw, &decoded.With[i]); err != nil {
					return fmt.Errorf("translation argument %d: %w", i, err)
				}
			}
		}
	default:
		return fmt.Errorf("invalid JSON text component")
	}
	*tc = decoded
	return nil
}

// validateComponentContent rejects compounds that would otherwise silently decode
// to empty text. A discriminator, when present, must select a supported content.
func validateComponentContent(has func(string) bool, kind func() (string, error)) error {
	if has("type") {
		contentType, err := kind()
		if err != nil {
			return err
		}
		switch contentType {
		case "translatable":
			if has("translate") {
				return nil
			}
		case "text", "keybind", "score", "selector", "nbt":
			if has(contentType) {
				return nil
			}
		}
		return fmt.Errorf("invalid text component type %q", contentType)
	}
	for _, field := range []string{"text", "translate", "keybind", "score", "selector", "nbt"} {
		if has(field) {
			return nil
		}
	}
	return fmt.Errorf("text component has no content")
}

// UnmarshalNBT decodes strings, compounds, and non-empty component lists.
// Vanilla's heterogeneous-list wrappers are unwrapped only at list boundaries.
func (tc *TextComponent) UnmarshalNBT(tag nbt.Tag) error {
	var decoded TextComponent
	switch value := tag.(type) {
	case nbt.String:
		decoded.Text = string(value)
	case nbt.List:
		list, err := decodeComponentList(value, false)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return fmt.Errorf("empty component list")
		}
		decoded = list[0]
		decoded.Extra = append(decoded.Extra, list[1:]...)
	case nbt.Compound:
		if err := validateComponentContent(func(key string) bool {
			return value[key] != nil
		}, func() (string, error) {
			kind, ok := value["type"].(nbt.String)
			if !ok {
				return "", fmt.Errorf("component type must be a string")
			}
			return string(kind), nil
		}); err != nil {
			return err
		}
		fields := value
		if value["extra"] != nil || value["with"] != nil {
			fields = make(nbt.Compound, len(value))
			for key, child := range value {
				if key != "extra" && key != "with" {
					fields[key] = child
				}
			}
		}
		type plain TextComponent
		if err := nbt.UnmarshalTag(fields, (*plain)(&decoded)); err != nil {
			return err
		}
		if extra, ok := value["extra"]; ok {
			var err error
			decoded.Extra, err = decodeComponentList(extra, false)
			if err != nil {
				return err
			}
			if len(decoded.Extra) == 0 {
				return fmt.Errorf("empty component extra list")
			}
		}
		if with, ok := value["with"]; ok {
			var err error
			decoded.With, err = decodeComponentList(with, true)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid NBT text component %T", tag)
	}
	*tc = decoded
	return nil
}

func decodeComponentList(tag nbt.Tag, arguments bool) ([]TextComponent, error) {
	var elements []nbt.Tag
	switch list := tag.(type) {
	case nbt.List:
		elements = list.Elements
	case nbt.ByteArray:
		if !arguments {
			return nil, fmt.Errorf("numeric array is not a component list")
		}
		result := make([]TextComponent, len(list))
		for i, value := range list {
			result[i].Text = strconv.FormatInt(int64(int8(value)), 10)
		}
		return result, nil
	case nbt.IntArray:
		if !arguments {
			return nil, fmt.Errorf("numeric array is not a component list")
		}
		result := make([]TextComponent, len(list))
		for i, value := range list {
			result[i].Text = strconv.FormatInt(int64(value), 10)
		}
		return result, nil
	case nbt.LongArray:
		if !arguments {
			return nil, fmt.Errorf("numeric array is not a component list")
		}
		result := make([]TextComponent, len(list))
		for i, value := range list {
			result[i].Text = strconv.FormatInt(value, 10)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected component list, got %T", tag)
	}
	list := make([]TextComponent, len(elements))
	for i, child := range elements {
		// ListTag.tryUnwrap unwraps exactly once, not recursively.
		if wrapper, ok := child.(nbt.Compound); ok && len(wrapper) == 1 {
			if wrapped, ok := wrapper[""]; ok {
				child = wrapped
			}
		}
		if arguments {
			switch primitive := child.(type) {
			case nbt.Byte:
				list[i].Text = strconv.FormatInt(int64(primitive), 10)
				continue
			case nbt.Short:
				list[i].Text = strconv.FormatInt(int64(primitive), 10)
				continue
			case nbt.Int:
				list[i].Text = strconv.FormatInt(int64(primitive), 10)
				continue
			case nbt.Long:
				list[i].Text = strconv.FormatInt(int64(primitive), 10)
				continue
			case nbt.Float:
				list[i].Text = strconv.FormatFloat(float64(primitive), 'f', -1, 32)
				continue
			case nbt.Double:
				list[i].Text = strconv.FormatFloat(float64(primitive), 'f', -1, 64)
				continue
			}
		}
		if err := list[i].UnmarshalNBT(child); err != nil {
			return nil, fmt.Errorf("component list element %d: %w", i, err)
		}
	}
	return list, nil
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

// ReadTextComponent reads a text component from the buffer (NBT wire format).
func (pb *PacketBuffer) ReadTextComponent() (TextComponent, error) {
	var tc TextComponent
	err := tc.Decode(pb)
	return tc, err
}

// WriteTextComponent writes a text component to the buffer (NBT wire format).
func (pb *PacketBuffer) WriteTextComponent(tc TextComponent) error {
	return tc.Encode(pb)
}

// ReadJsonTextComponent reads a text component as a VarInt-prefixed JSON string.
// Used by login disconnect (ByteBufCodecs.lenientJson in vanilla).
func (pb *PacketBuffer) ReadJsonTextComponent() (TextComponent, error) {
	s, err := pb.ReadString(262144)
	if err != nil {
		return TextComponent{}, err
	}
	var tc TextComponent
	if err := json.Unmarshal([]byte(s), &tc); err != nil {
		return TextComponent{}, err
	}
	return tc, nil
}

// WriteJsonTextComponent writes a text component as a VarInt-prefixed JSON string.
func (pb *PacketBuffer) WriteJsonTextComponent(tc TextComponent) error {
	data, err := json.Marshal(tc)
	if err != nil {
		return err
	}
	return pb.WriteString(String(data))
}
