package net_structures_test

import (
	"encoding/json"
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestString(t *testing.T) {
	tests := []struct {
		name      string
		val       ns.String
		maxLength int
	}{
		{"empty", "", 100},
		{"hello", "Hello, World!", 100},
		{"unicode", "Hello, 世界! 🌍", 100},
		{"long", "This is a relatively long string for testing purposes", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("String.Marshal() error = %v", err)
			}

			var unmarshaled ns.String
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalString() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("String.Unmarshal() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestIdentifier(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Identifier
	}{
		{"minecraft:stone", "minecraft:stone"},
		{"minecraft:dirt", "minecraft:dirt"},
		{"custom:item", "custom:item"},
		{"namespace:path/to/resource", "namespace:path/to/resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("MarshalIdentifier() error = %v", err)
			}

			var unmarshaled ns.Identifier
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalIdentifier() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalIdentifier() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestJSONTextComponent(t *testing.T) {
	tests := []struct {
		name string
		val  ns.JSONTextComponent
	}{
		{"simple", ns.JSONTextComponent{"text": "Hello World"}},
		{"complex", ns.JSONTextComponent{"text": "Hello", "color": "red", "bold": true}},
		{"with_extra", ns.JSONTextComponent{"text": "Hello", "extra": []any{map[string]any{"text": " World", "color": "blue"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("MarshalJSONTextComponent() error = %v", err)
			}

			var unmarshaled ns.JSONTextComponent
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalJSONTextComponent() error = %v", err)
			}

			// Compare each key-value pair since map comparison with any values
			// can be tricky with different types
			if len(unmarshaled) != len(tt.val) {
				t.Errorf("UnmarshalJSONTextComponent() length = %v, want %v", len(unmarshaled), len(tt.val))
				return
			}

			for key, expectedVal := range tt.val {
				actualVal, exists := unmarshaled[key]
				if !exists {
					t.Errorf("UnmarshalJSONTextComponent() missing key %v", key)
					continue
				}

				// For simple comparison, convert both to JSON strings
				expectedJSON, _ := json.Marshal(expectedVal)
				actualJSON, _ := json.Marshal(actualVal)
				if string(expectedJSON) != string(actualJSON) {
					t.Errorf("UnmarshalJSONTextComponent() key %v = %v, want %v", key, actualVal, expectedVal)
				}
			}
		})
	}
}

func TestStringErrorCases(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		var s ns.String
		_, err := s.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("String.Unmarshal() should error on insufficient data")
		}
	})
}

func TestStringInterface(t *testing.T) {
	val := ns.String("Hello, Interface!")
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("String.Marshal() error = %v", err)
	}

	var result ns.String
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("String.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("String interface roundtrip: got %v, want %v", result, val)
	}
}

func TestIdentifierInterface(t *testing.T) {
	val := ns.Identifier("minecraft:stone")
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Identifier.Marshal() error = %v", err)
	}

	var result ns.Identifier
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Identifier.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Identifier interface roundtrip: got %v, want %v", result, val)
	}
}

func TestStringGenericMarshal(t *testing.T) {
	val := ns.String("Hello, Generic!")
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Marshal(String) error = %v", err)
	}

	var result ns.String
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Unmarshal(String) error = %v", err)
	}
	if result != val {
		t.Errorf("Generic String roundtrip: got %v, want %v", result, val)
	}
}
