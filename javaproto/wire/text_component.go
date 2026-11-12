package wire

import (
	"errors"
	"fmt"
	"maps"
	"slices"
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
	Text       string           `nbt:"text,omitempty" json:"text,omitempty"`
	Translate  string           `nbt:"translate,omitempty" json:"translate,omitempty"`
	Fallback   *string          `nbt:"fallback,omitempty" json:"fallback,omitempty"`
	Keybind    string           `nbt:"keybind,omitempty" json:"keybind,omitempty"`
	Score      *Score           `nbt:"score,omitempty" json:"score,omitempty"`
	Selector   string           `nbt:"selector,omitempty" json:"selector,omitempty"`
	NBT        string           `nbt:"nbt,omitempty" json:"nbt,omitempty"`
	NBTBlock   string           `nbt:"block,omitempty" json:"block,omitempty"`         // for nbt content type
	NBTEntity  string           `nbt:"entity,omitempty" json:"entity,omitempty"`       // for nbt content type
	NBTStorage string           `nbt:"storage,omitempty" json:"storage,omitempty"`     // for nbt content type
	Interpret  *bool            `nbt:"interpret,omitempty" json:"interpret,omitempty"` // for nbt content type
	Object     *ObjectComponent `nbt:"object,omitempty" json:"-"`

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
		tc.Object == nil &&
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

// Literal content keys and the keys that select an object component, as they
// appear in the NBT/JSON component codecs.
var (
	componentContentKeys = []string{"text", "translate", "keybind", "score", "selector", "nbt"}
	objectContentKeys    = []string{"player", "sprite"}
)

// hasContentType returns true if any content type field is non-empty.
func (tc *TextComponent) hasContentType() bool {
	return tc.Text != "" || tc.Translate != "" || tc.Keybind != "" ||
		tc.Score != nil || tc.Selector != "" || tc.NBT != "" || tc.Object != nil
}

// textComponentFields has TextComponent's fields without its MarshalNBT
// method, so reflection can marshal it without recursing into the hook.
type textComponentFields TextComponent

// MarshalNBT implements nbt.TagMarshaler with the component's wire encoding,
// so packets, persistence and nested reflection share one representation.
// Simple text-only components are encoded as NBT String tags for efficiency.
func (tc TextComponent) MarshalNBT() (nbt.Tag, error) {
	if tc.isSimpleText() {
		return nbt.String(tc.Text), nil
	}
	tag, err := tc.marshalCompound()
	if err != nil {
		return nil, err
	}
	// nested extra/with parents need an explicit content type too
	normalizeComponentTag(tag)
	return tag, nil
}

// marshalCompound marshals the component as a compound. Children are always
// compounds so extra/with lists stay homogeneous.
func (tc TextComponent) marshalCompound() (nbt.Compound, error) {
	extra, with := tc.Extra, tc.With
	tc.Extra, tc.With = nil, nil
	tag, err := nbt.MarshalTag(textComponentFields(tc))
	if err != nil {
		return nil, err
	}
	compound := tag.(nbt.Compound)
	for key, children := range map[string][]TextComponent{"extra": extra, "with": with} {
		if len(children) == 0 {
			continue
		}
		elements := make([]nbt.Tag, len(children))
		for i, child := range children {
			if elements[i], err = child.marshalCompound(); err != nil {
				return nil, fmt.Errorf("%s %d: %w", key, i, err)
			}
		}
		compound[key] = nbt.List{ElementType: nbt.TagCompound, Elements: elements}
	}
	return compound, nil
}

// Encode writes the text component as NBT to the writer.
func (tc *TextComponent) Encode(buf *PacketBuffer) error {
	tag, err := tc.MarshalNBT()
	if err != nil {
		return err
	}
	return buf.WriteNBT(tag)
}

// normalizeComponentTag restores omitted literal content and flattens objects.
// Only component-valued fields are traversed; event payloads remain untouched.
func normalizeComponentTag(tag nbt.Tag) {
	switch value := tag.(type) {
	case nbt.List:
		for _, child := range value.Elements {
			normalizeComponentTag(child)
		}
	case nbt.Compound:
		// Reflection nests the API descriptor; the component codec is flat.
		if object, ok := value["object"].(nbt.Compound); ok {
			delete(value, "object")
			maps.Copy(value, object)
			value["type"] = nbt.String("object")
			normalizeComponentTag(value["fallback"])
		}
		hasKey := func(key string) bool {
			_, ok := value[key]
			return ok
		}
		if !slices.ContainsFunc(componentContentKeys, hasKey) && !slices.ContainsFunc(objectContentKeys, hasKey) {
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

// validateComponentContent rejects compounds that would otherwise silently decode
// to empty text. A discriminator, when present, must select a supported content.
func validateComponentContent(has func(string) bool, kind func() (string, error)) (string, error) {
	if has("type") {
		contentType, err := kind()
		if err != nil {
			return "", err
		}
		switch contentType {
		case "translatable":
			if has("translate") {
				return contentType, nil
			}
		case "text", "keybind", "score", "selector", "nbt":
			if has(contentType) {
				return contentType, nil
			}
		case "object":
			if slices.ContainsFunc(objectContentKeys, has) {
				return contentType, nil
			}
		}
		return "", fmt.Errorf("invalid text component type %q", contentType)
	}
	if i := slices.IndexFunc(componentContentKeys, has); i >= 0 {
		return componentContentKeys[i], nil
	}
	if slices.ContainsFunc(objectContentKeys, has) {
		return "object", nil
	}
	return "", errors.New("text component has no content")
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
			return errors.New("empty component list")
		}
		decoded = list[0]
		decoded.Extra = append(decoded.Extra, list[1:]...)
	case nbt.Compound:
		contentType, err := validateComponentContent(func(key string) bool {
			return value[key] != nil
		}, func() (string, error) {
			kind, ok := value["type"].(nbt.String)
			if !ok {
				return "", errors.New("component type must be a string")
			}
			return string(kind), nil
		})
		if err != nil {
			return err
		}
		fields := value
		if value["extra"] != nil || value["with"] != nil || value["object"] != nil || contentType == "object" {
			fields = make(nbt.Compound, len(value))
			for key, child := range value {
				if key != "extra" && key != "with" && key != "object" && !(contentType == "object" && key == "fallback") {
					fields[key] = child
				}
			}
		}
		type plain TextComponent
		if err := nbt.UnmarshalTag(fields, (*plain)(&decoded)); err != nil {
			return err
		}
		if contentType == "object" {
			decoded.Object = new(ObjectComponent)
			if err := nbt.UnmarshalTag(value, decoded.Object); err != nil {
				return err
			}
			if err := decoded.Object.validate(); err != nil {
				return err
			}
		}
		if extra, ok := value["extra"]; ok {
			var err error
			decoded.Extra, err = decodeComponentList(extra, false)
			if err != nil {
				return err
			}
			if len(decoded.Extra) == 0 {
				return errors.New("empty component extra list")
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

var errNumericComponentList = errors.New("numeric array is not a component list")

func decodeComponentList(tag nbt.Tag, arguments bool) ([]TextComponent, error) {
	var elements []nbt.Tag
	switch list := tag.(type) {
	case nbt.List:
		elements = list.Elements
	case nbt.ByteArray:
		return numericArguments(arguments, list, func(v byte) int64 { return int64(int8(v)) })
	case nbt.IntArray:
		return numericArguments(arguments, list, func(v int32) int64 { return int64(v) })
	case nbt.LongArray:
		return numericArguments(arguments, list, func(v int64) int64 { return v })
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
			if text, ok := primitiveArgument(child); ok {
				list[i].Text = text
				continue
			}
		}
		if err := list[i].UnmarshalNBT(child); err != nil {
			return nil, fmt.Errorf("component list element %d: %w", i, err)
		}
	}
	return list, nil
}

// numericArguments converts a numeric array, valid only as translation arguments.
func numericArguments[E any](arguments bool, values []E, toInt func(E) int64) ([]TextComponent, error) {
	if !arguments {
		return nil, errNumericComponentList
	}
	result := make([]TextComponent, len(values))
	for i, v := range values {
		result[i].Text = strconv.FormatInt(toInt(v), 10)
	}
	return result, nil
}

// primitiveArgument formats a numeric translation argument.
func primitiveArgument(tag nbt.Tag) (string, bool) {
	switch v := tag.(type) {
	case nbt.Byte:
		return strconv.FormatInt(int64(v), 10), true
	case nbt.Short:
		return strconv.FormatInt(int64(v), 10), true
	case nbt.Int:
		return strconv.FormatInt(int64(v), 10), true
	case nbt.Long:
		return strconv.FormatInt(int64(v), 10), true
	case nbt.Float:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case nbt.Double:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), true
	}
	return "", false
}

// Decode reads a text component from NBT.
func (tc *TextComponent) Decode(buf *PacketBuffer) error {
	tag, err := buf.ReadNBT()
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
