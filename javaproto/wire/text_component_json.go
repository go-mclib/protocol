package wire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// MarshalJSON retains the empty literal content of styled or parent components.
func (tc TextComponent) MarshalJSON() ([]byte, error) {
	type component TextComponent
	if tc.Object != nil {
		return json.Marshal(struct {
			component
			*ObjectComponent
			Type     string         `json:"type"`
			Kind     string         `json:"object"`
			Fallback *TextComponent `json:"fallback,omitempty"`
		}{component(tc), tc.Object, "object", tc.Object.kind(), tc.Object.Fallback})
	}
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
		return errors.New("empty text component")
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
			return errors.New("empty component list")
		}
		decoded = list[0]
		decoded.Extra = append(decoded.Extra, list[1:]...)
	case '{':
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		contentType, err := validateComponentContent(func(key string) bool {
			value, ok := fields[key]
			return ok && (key == "type" || !bytes.Equal(bytes.TrimSpace(value), []byte("null")))
		}, func() (string, error) {
			var kind string
			err := json.Unmarshal(fields["type"], &kind)
			return kind, err
		})
		if err != nil {
			return err
		}
		type plain TextComponent
		value := struct {
			*plain
			With     []json.RawMessage `json:"with"`
			Extra    json.RawMessage   `json:"extra"`
			Fallback json.RawMessage   `json:"fallback"`
		}{plain: (*plain)(&decoded)}
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		if contentType == "object" {
			decoded.Object = new(ObjectComponent)
			if err := json.Unmarshal(data, decoded.Object); err != nil {
				return err
			}
			if err := decoded.Object.validate(); err != nil {
				return err
			}
		} else if value.Fallback != nil {
			if err := json.Unmarshal(value.Fallback, &decoded.Fallback); err != nil {
				return err
			}
		}
		if value.Extra != nil {
			if err := json.Unmarshal(value.Extra, &decoded.Extra); err != nil {
				return err
			}
			if len(decoded.Extra) == 0 {
				return errors.New("empty component extra list")
			}
		}
		if raw, ok := fields["with"]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("null translation arguments")
		}
		if len(value.With) != 0 {
			if decoded.With, err = decodeJSONArguments(value.With); err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid JSON text component")
	}
	*tc = decoded
	return nil
}

// decodeJSONArguments decodes translation arguments. Vanilla accepts primitive
// arguments (numbers, booleans), but not primitive components.
func decodeJSONArguments(raw []json.RawMessage) ([]TextComponent, error) {
	args := make([]TextComponent, len(raw))
	for i, arg := range raw {
		arg = bytes.TrimSpace(arg)
		if bytes.Equal(arg, []byte("true")) || bytes.Equal(arg, []byte("false")) ||
			(len(arg) > 0 && (arg[0] == '-' || arg[0] >= '0' && arg[0] <= '9')) {
			args[i].Text = string(arg)
		} else if err := json.Unmarshal(arg, &args[i]); err != nil {
			return nil, fmt.Errorf("translation argument %d: %w", i, err)
		}
	}
	return args, nil
}

// ReadJSONTextComponent reads a text component as a VarInt-prefixed JSON string.
// Used by login disconnect (ByteBufCodecs.lenientJson in vanilla).
func (pb *PacketBuffer) ReadJSONTextComponent() (TextComponent, error) {
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

// WriteJSONTextComponent writes a text component as a VarInt-prefixed JSON string.
func (pb *PacketBuffer) WriteJSONTextComponent(tc TextComponent) error {
	data, err := json.Marshal(tc)
	if err != nil {
		return err
	}
	return pb.WriteString(String(data))
}
