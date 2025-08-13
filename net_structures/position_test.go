package net_structures_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestPosition(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Position
	}{
		{"origin", ns.Position{X: 0, Y: 0, Z: 0}},
		{"positive", ns.Position{X: 100, Y: 64, Z: 200}},
		{"negative", ns.Position{X: -100, Y: -64, Z: -200}},
		{"max", ns.Position{X: 33554431, Y: 2047, Z: 33554431}},
		{"min", ns.Position{X: -33554432, Y: -2048, Z: -33554432}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Position.Marshal() error = %v", err)
			}
			var unmarshaled ns.Position
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalPosition() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalPosition() = %+v, want %+v", unmarshaled, tt.val)
			}
		})
	}
}

func TestPositionErrorCases(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		var temp ns.Position
		_, err := temp.FromBytes(ns.ByteArray{0x01, 0x02, 0x03})
		if err == nil {
			t.Error("Position.Unmarshal() should error on insufficient data")
		}
	})
}

func TestPositionInterface(t *testing.T) {
	val := ns.Position{X: 100, Y: 64, Z: 200}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Position.Marshal() error = %v", err)
	}

	var result ns.Position
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Position.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Position interface roundtrip: got %+v, want %+v", result, val)
	}
}

func TestPositionGenericMarshal(t *testing.T) {
	val := ns.Position{X: 100, Y: 64, Z: 200}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Marshal(Position) error = %v", err)
	}

	var result ns.Position
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Unmarshal(Position) error = %v", err)
	}
	if result != val {
		t.Errorf("Generic Position roundtrip: got %+v, want %+v", result, val)
	}
}
