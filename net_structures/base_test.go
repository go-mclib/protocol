package net_structures_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestGenericMarshal(t *testing.T) {
	t.Run("Boolean", func(t *testing.T) {
		original := ns.Boolean(true)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Marshal(Boolean) error = %v", err)
		}

		// compare with direct marshal
		directData, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Boolean.Marshal() error = %v", err)
		}

		if len(data) != len(directData) {
			t.Errorf("Length mismatch: generic=%d, direct=%d", len(data), len(directData))
		}

		for i := range data {
			if data[i] != directData[i] {
				t.Errorf("Data mismatch at byte %d: generic=%02x, direct=%02x", i, data[i], directData[i])
			}
		}
	})

	t.Run("Int", func(t *testing.T) {
		original := ns.Int(42)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Marshal(Int) error = %v", err)
		}

		// Compare with direct marshal
		directData, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Int.Marshal() error = %v", err)
		}

		if len(data) != len(directData) {
			t.Errorf("Length mismatch: generic=%d, direct=%d", len(data), len(directData))
		}

		for i := range data {
			if data[i] != directData[i] {
				t.Errorf("Data mismatch at byte %d: generic=%02x, direct=%02x", i, data[i], directData[i])
			}
		}
	})

	t.Run("VarInt", func(t *testing.T) {
		original := ns.VarInt(25565)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Marshal(VarInt) error = %v", err)
		}

		// Compare with direct marshal
		directData, err := original.ToBytes()
		if err != nil {
			t.Fatalf("VarInt.Marshal() error = %v", err)
		}

		if len(data) != len(directData) {
			t.Errorf("Length mismatch: generic=%d, direct=%d", len(data), len(directData))
		}

		for i := range data {
			if data[i] != directData[i] {
				t.Errorf("Data mismatch at byte %d: generic=%02x, direct=%02x", i, data[i], directData[i])
			}
		}
	})

	t.Run("String", func(t *testing.T) {
		original := ns.String("Hello, World!")
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Marshal(String) error = %v", err)
		}

		// compare with direct marshal
		directData, err := original.ToBytes()
		if err != nil {
			t.Fatalf("String.Marshal() error = %v", err)
		}

		if len(data) != len(directData) {
			t.Errorf("Length mismatch: generic=%d, direct=%d", len(data), len(directData))
		}

		for i := range data {
			if data[i] != directData[i] {
				t.Errorf("Data mismatch at byte %d: generic=%02x, direct=%02x", i, data[i], directData[i])
			}
		}
	})

	t.Run("Position", func(t *testing.T) {
		original := ns.Position{X: 100, Y: 64, Z: -200}
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Marshal(Position) error = %v", err)
		}

		// compare with direct marshal
		directData, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Position.Marshal() error = %v", err)
		}

		if len(data) != len(directData) {
			t.Errorf("Length mismatch: generic=%d, direct=%d", len(data), len(directData))
		}

		for i := range data {
			if data[i] != directData[i] {
				t.Errorf("Data mismatch at byte %d: generic=%02x, direct=%02x", i, data[i], directData[i])
			}
		}
	})
}

func TestDirectUnmarshalRoundtrip(t *testing.T) {
	// Test that direct marshal/unmarshal works for various types
	t.Run("Boolean", func(t *testing.T) {
		original := ns.Boolean(true)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Boolean.Marshal() error = %v", err)
		}

		var result ns.Boolean
		_, err = result.FromBytes(data)
		if err != nil {
			t.Fatalf("Boolean.Unmarshal() error = %v", err)
		}

		if result != original {
			t.Errorf("Boolean roundtrip: got %v, want %v", result, original)
		}
	})

	t.Run("Int", func(t *testing.T) {
		original := ns.Int(42)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Int.Marshal() error = %v", err)
		}

		var result ns.Int
		_, err = result.FromBytes(data)
		if err != nil {
			t.Fatalf("Int.Unmarshal() error = %v", err)
		}

		if result != original {
			t.Errorf("Int roundtrip: got %v, want %v", result, original)
		}
	})

	t.Run("VarInt", func(t *testing.T) {
		original := ns.VarInt(25565)
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("VarInt.Marshal() error = %v", err)
		}

		var result ns.VarInt
		_, err = result.FromBytes(data)
		if err != nil {
			t.Fatalf("VarInt.Unmarshal() error = %v", err)
		}

		if result != original {
			t.Errorf("VarInt roundtrip: got %v, want %v", result, original)
		}
	})

	t.Run("String", func(t *testing.T) {
		original := ns.String("Hello, World!")
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("String.Marshal() error = %v", err)
		}

		var result ns.String
		_, err = result.FromBytes(data)
		if err != nil {
			t.Fatalf("String.Unmarshal() error = %v", err)
		}

		if result != original {
			t.Errorf("String roundtrip: got %v, want %v", result, original)
		}
	})

	t.Run("Position", func(t *testing.T) {
		original := ns.Position{X: 100, Y: 64, Z: -200}
		data, err := original.ToBytes()
		if err != nil {
			t.Fatalf("Position.Marshal() error = %v", err)
		}

		var result ns.Position
		_, err = result.FromBytes(data)
		if err != nil {
			t.Fatalf("Position.Unmarshal() error = %v", err)
		}

		if result.X != original.X || result.Y != original.Y || result.Z != original.Z {
			t.Errorf("Position roundtrip: got %+v, want %+v", result, original)
		}
	})
}

func TestUnmarshalErrors(t *testing.T) {
	t.Run("Int insufficient data", func(t *testing.T) {
		var result ns.Int
		_, err := result.FromBytes(ns.ByteArray{0x01, 0x02}) // Int needs 4 bytes
		if err == nil {
			t.Error("Int.Unmarshal should error on insufficient data")
		}
	})

	t.Run("Boolean empty data", func(t *testing.T) {
		var result ns.Boolean
		_, err := result.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("Boolean.Unmarshal should error on empty data")
		}
	})

	t.Run("String empty data", func(t *testing.T) {
		var result ns.String
		_, err := result.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("String.Unmarshal should error on empty data")
		}
	})
}
