package net_structures

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatTextComponent represents a Minecraft chat text component
type ChatTextComponent struct {
	Text  string              `json:"text,omitempty"`
	Color string              `json:"color,omitempty"`
	Bold  bool                `json:"bold,omitempty"`
	Extra []ChatTextComponent `json:"extra,omitempty"`
	Raw   map[string]any      `json:"-"` // store raw data for unknown fields
}

// ExtractPlainText extracts plain text from a chat component, handling all formatting
func (c ChatTextComponent) ExtractPlainText() string {
	var result strings.Builder

	// Add main text
	if c.Text != "" {
		result.WriteString(c.Text)
	}

	// Add extra components recursively
	for _, extra := range c.Extra {
		result.WriteString(extra.ExtractPlainText())
	}

	return result.String()
}

// String returns a formatted string representation
func (c ChatTextComponent) String() string {
	text := c.ExtractPlainText()
	if text != "" {
		return text
	}

	if c.Raw != nil {
		if translate, ok := c.Raw["translate"].(string); ok {
			if with, ok := c.Raw["with"].([]any); ok {
				var parts []string
				for _, arg := range with {
					switch v := arg.(type) {
					case map[string]any:
						parts = append(parts, extractTextFromMap(v))
					case string:
						parts = append(parts, v)
					default:
						parts = append(parts, fmt.Sprintf("%v", v))
					}
				}
				return fmt.Sprintf("%s [%s]", translate, strings.Join(parts, ", "))
			}
			return translate
		}

		for key, value := range c.Raw {
			if strings.Contains(key, "text") {
				if str, ok := value.(string); ok {
					return str
				}
			}
		}
	}

	return "<empty text component>"
}

// ParseTextComponentFromString attempts to parse a text component from JSON string
func ParseTextComponentFromString(jsonStr string) (ChatTextComponent, error) {
	var component ChatTextComponent

	if !strings.HasPrefix(jsonStr, "{") && !strings.HasPrefix(jsonStr, "[") {
		component.Text = jsonStr
		return component, nil
	}

	if err := json.Unmarshal([]byte(jsonStr), &component); err != nil {
		var raw map[string]any
		if err2 := json.Unmarshal([]byte(jsonStr), &raw); err2 == nil {
			component.Raw = raw
			if text, ok := raw["text"].(string); ok {
				component.Text = text
			}
			if color, ok := raw["color"].(string); ok {
				component.Color = color
			}
			return component, nil
		}
		return component, fmt.Errorf("failed to parse text component: %w", err)
	}

	return component, nil
}
