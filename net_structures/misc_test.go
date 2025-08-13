package net_structures_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestAngle(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Angle
	}{
		{"zero", 0},
		{"quarter", 64},
		{"half", 128},
		{"three quarters", 192},
		{"full", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Angle.Marshal() error = %v", err)
			}
			var unmarshaled ns.Angle
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalAngle() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalAngle() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name string
		val  ns.UUID
	}{
		{"zero", ns.UUID{}},
		{"ones", ns.UUID{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		{"random", ns.UUID{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("UUID.Marshal() error = %v", err)
			}
			var unmarshaled ns.UUID
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalUUID() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalUUID() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestTeleportFlags(t *testing.T) {
	tests := []struct {
		name string
		val  ns.TeleportFlags
	}{
		{"zero", 0},
		{"all set", 0xFFFFFFFF},
		{"some flags", 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("TeleportFlags.Marshal() error = %v", err)
			}
			var unmarshaled ns.TeleportFlags
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalTeleportFlags() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalTeleportFlags() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestMiscErrorCases(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		// Test Angle with insufficient data
		var a ns.Angle
		_, err := a.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("Angle.Unmarshal() should error on empty data")
		}

		// Test UUID with insufficient data
		var u ns.UUID
		_, err = u.FromBytes(ns.ByteArray{0x01, 0x02, 0x03})
		if err == nil {
			t.Error("UUID.Unmarshal() should error on insufficient data")
		}

		// Test TeleportFlags with insufficient data
		var tf ns.TeleportFlags
		_, err = tf.FromBytes(ns.ByteArray{0x01, 0x02, 0x03})
		if err == nil {
			t.Error("TeleportFlags.Unmarshal() should error on insufficient data")
		}
	})
}

func TestAngleInterface(t *testing.T) {
	val := ns.Angle(128)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Angle.Marshal() error = %v", err)
	}

	var result ns.Angle
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Angle.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Angle interface roundtrip: got %v, want %v", result, val)
	}
}

func TestUUIDInterface(t *testing.T) {
	val := ns.UUID{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF,
		0xFE, 0xDC, 0xBA, 0x09, 0x87, 0x65, 0x43, 0x21,
	}
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("UUID.Marshal() error = %v", err)
	}

	var result ns.UUID
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("UUID.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("UUID interface roundtrip: got %v, want %v", result, val)
	}
}

func TestTeleportFlagsInterface(t *testing.T) {
	val := ns.TeleportFlags(0x12345678)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("TeleportFlags.Marshal() error = %v", err)
	}

	var result ns.TeleportFlags
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("TeleportFlags.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("TeleportFlags interface roundtrip: got %v, want %v", result, val)
	}
}

func TestMiscGenericMarshal(t *testing.T) {
	t.Run("Angle", func(t *testing.T) {
		val := ns.Angle(128)
		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Marshal(Angle) error = %v", err)
		}

		var result ns.Angle
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Unmarshal(Angle) error = %v", err)
		}
		if result != val {
			t.Errorf("Generic Angle roundtrip: got %v, want %v", result, val)
		}
	})

	t.Run("UUID", func(t *testing.T) {
		val := ns.UUID{
			0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF,
			0xFE, 0xDC, 0xBA, 0x09, 0x87, 0x65, 0x43, 0x21,
		}
		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Marshal(UUID) error = %v", err)
		}

		var result ns.UUID
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Unmarshal(UUID) error = %v", err)
		}
		if result != val {
			t.Errorf("Generic UUID roundtrip: got %v, want %v", result, val)
		}
	})
}
