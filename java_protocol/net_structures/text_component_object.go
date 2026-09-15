package net_structures

import (
	"fmt"
	"math"
	"strings"
)

// ObjectComponent is inline atlas/player imagery in a text component. Player is
// a profile name string or a map of profile fields (name, id, properties and skin
// patches). Profile metadata is decorative, never authenticated sender identity.
// Fallback is a full component, unlike a translation's string fallback.
// Object fields are flattened alongside the enclosing component's style on wire.
type ObjectComponent struct {
	Kind     string         `nbt:"object,omitempty" json:"object,omitempty"`
	Atlas    string         `nbt:"atlas,omitempty" json:"atlas,omitempty"`
	Sprite   string         `nbt:"sprite,omitempty" json:"sprite,omitempty"`
	Player   any            `nbt:"player,omitempty" json:"player,omitempty"`
	Hat      *bool          `nbt:"hat,omitempty" json:"hat,omitempty"`
	Fallback *TextComponent `nbt:"fallback,omitempty" json:"fallback,omitempty"`
}

func (o *ObjectComponent) kind() string {
	if o.Kind != "" {
		return o.Kind
	}
	if o.Sprite != "" {
		return "atlas"
	}
	return "player"
}

func (o *ObjectComponent) validate() error {
	switch o.kind() {
	case "atlas":
		if o.Sprite == "" {
			return fmt.Errorf("atlas component requires a sprite")
		}
	case "player":
		switch profile := o.Player.(type) {
		case string:
			if profile == "" {
				return fmt.Errorf("player component requires a non-empty profile name")
			}
		case map[string]any:
			if name, exists := profile["name"]; exists {
				if _, ok := name.(string); !ok {
					return fmt.Errorf("player profile name must be a string")
				}
			}
			// JSON numeric arrays must retain the UUID's NBT IntArray wire type.
			if id, exists := profile["id"]; exists {
				switch values := id.(type) {
				case []int32:
					if len(values) != 4 {
						return fmt.Errorf("player profile UUID requires four integers")
					}
				case []any:
					if len(values) != 4 {
						return fmt.Errorf("player profile UUID requires four integers")
					}
					ints := make([]int32, 4)
					for i, value := range values {
						n, ok := value.(float64)
						if !ok || math.Trunc(n) != n || n < math.MinInt32 || n > math.MaxInt32 {
							return fmt.Errorf("player profile UUID requires int32 values")
						}
						ints[i] = int32(n)
					}
					profile["id"] = ints
				default:
					return fmt.Errorf("player profile UUID must be an integer array")
				}
			}
		default:
			return fmt.Errorf("player component requires a profile string or compound")
		}
	default:
		return fmt.Errorf("invalid object component type %q", o.Kind)
	}
	return nil
}

// defaultFallback follows PlayerSprite/AtlasSprite.defaultFallback in vanilla.
func (o *ObjectComponent) defaultFallback() string {
	if o.kind() == "atlas" {
		sprite := strings.TrimPrefix(o.Sprite, "minecraft:")
		atlas := strings.TrimPrefix(o.Atlas, "minecraft:")
		if atlas == "" || atlas == "blocks" {
			return "[" + sprite + "]"
		}
		return "[" + sprite + "@" + atlas + "]"
	}
	name, _ := o.Player.(string)
	if profile, ok := o.Player.(map[string]any); ok {
		name, _ = profile["name"].(string)
	}
	if name == "" {
		return "[unknown player head]"
	}
	return "[" + name + " head]"
}
