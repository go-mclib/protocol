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

func TestPrefixedOptional(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		val := ns.PrefixedOptional[ns.VarInt]{
			Present: true,
			Value:   42,
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("PrefixedOptional.ToBytes() error = %v", err)
		}

		var result ns.PrefixedOptional[ns.VarInt]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("PrefixedOptional.FromBytes() error = %v", err)
		}

		if result.Present != val.Present || result.Value != val.Value {
			t.Errorf("PrefixedOptional roundtrip: got %+v, want %+v", result, val)
		}
	})

	t.Run("not present", func(t *testing.T) {
		val := ns.PrefixedOptional[ns.VarInt]{
			Present: false,
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("PrefixedOptional.ToBytes() error = %v", err)
		}

		var result ns.PrefixedOptional[ns.VarInt]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("PrefixedOptional.FromBytes() error = %v", err)
		}

		if result.Present != val.Present {
			t.Errorf("PrefixedOptional roundtrip: got Present=%v, want Present=%v", result.Present, val.Present)
		}
	})
}

func TestPrefixedArray(t *testing.T) {
	val := ns.PrefixedArray[ns.VarInt]{1, 2, 3, 42, 1000}

	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("PrefixedArray.ToBytes() error = %v", err)
	}

	var result ns.PrefixedArray[ns.VarInt]
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("PrefixedArray.FromBytes() error = %v", err)
	}

	if len(result) != len(val) {
		t.Errorf("PrefixedArray length: got %v, want %v", len(result), len(val))
	}

	for i, v := range val {
		if result[i] != v {
			t.Errorf("PrefixedArray data[%d]: got %v, want %v", i, result[i], v)
		}
	}
}

func TestIDor(t *testing.T) {
	t.Run("registry ID", func(t *testing.T) {
		val := ns.IDor[ns.VarInt]{
			IsID: true,
			ID:   123,
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("IDor.ToBytes() error = %v", err)
		}

		var result ns.IDor[ns.VarInt]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("IDor.FromBytes() error = %v", err)
		}

		if result.IsID != val.IsID || result.ID != val.ID {
			t.Errorf("IDor roundtrip: got %+v, want %+v", result, val)
		}
	})

	t.Run("inline data", func(t *testing.T) {
		val := ns.IDor[ns.VarInt]{
			IsID: false,
			Data: 456,
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("IDor.ToBytes() error = %v", err)
		}

		var result ns.IDor[ns.VarInt]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("IDor.FromBytes() error = %v", err)
		}

		if result.IsID != val.IsID || result.Data != val.Data {
			t.Errorf("IDor roundtrip: got %+v, want %+v", result, val)
		}
	})
}

func TestOr(t *testing.T) {
	t.Run("X type", func(t *testing.T) {
		val := ns.Or[ns.VarInt, ns.String]{
			IsX:  true,
			XVal: 42,
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Or.ToBytes() error = %v", err)
		}

		var result ns.Or[ns.VarInt, ns.String]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Or.FromBytes() error = %v", err)
		}

		if result.IsX != val.IsX || result.XVal != val.XVal {
			t.Errorf("Or roundtrip: got %+v, want %+v", result, val)
		}
	})

	t.Run("Y type", func(t *testing.T) {
		val := ns.Or[ns.VarInt, ns.String]{
			IsX:  false,
			YVal: "hello",
		}

		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Or.ToBytes() error = %v", err)
		}

		var result ns.Or[ns.VarInt, ns.String]
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Or.FromBytes() error = %v", err)
		}

		if result.IsX != val.IsX || result.YVal != val.YVal {
			t.Errorf("Or roundtrip: got %+v, want %+v", result, val)
		}
	})
}

func TestEntityMetadata(t *testing.T) {
	val := ns.EntityMetadata{
		Data: ns.ByteArray{0xAA, 0xBB, 0xCC},
	}

	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("EntityMetadata.ToBytes() error = %v", err)
	}

	var result ns.EntityMetadata
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("EntityMetadata.FromBytes() error = %v", err)
	}

	if len(result.Data) != len(val.Data) {
		t.Errorf("EntityMetadata data length: got %v, want %v", len(result.Data), len(val.Data))
	}

	for i, b := range val.Data {
		if result.Data[i] != b {
			t.Errorf("EntityMetadata data[%d]: got %02x, want %02x", i, result.Data[i], b)
		}
	}
}

func TestFixedByteArray(t *testing.T) {
	fba := ns.NewFixedByteArray(256)
	for i := range fba.Data {
		fba.Data[i] = byte(i % 256)
	}
	
	encoded, err := fba.ToBytes()
	if err != nil {
		t.Fatalf("Failed to encode FixedByteArray: %v", err)
	}
	
	if len(encoded) != 256 {
		t.Errorf("Encoded length mismatch: got %d, want 256", len(encoded))
	}
	
	var decoded ns.FixedByteArray
	decoded.Length = 256
	bytesRead, err := decoded.FromBytes(encoded)
	if err != nil {
		t.Fatalf("Failed to decode FixedByteArray: %v", err)
	}
	
	if bytesRead != 256 {
		t.Errorf("Bytes read mismatch: got %d, want 256", bytesRead)
	}
	
	for i, b := range decoded.Data {
		if b != fba.Data[i] {
			t.Errorf("Decoded data mismatch at %d: got %02x, want %02x", i, b, fba.Data[i])
			break
		}
	}
}

// Test Optional and PrefixedOptional with FixedByteArray
func TestOptionalFixedByteArray(t *testing.T) {
	po := ns.PrefixedOptional[ns.FixedByteArray]{
		Present: true,
		Value:   ns.NewFixedByteArray(256),
	}
	
	for i := range po.Value.Data {
		po.Value.Data[i] = byte(i % 256)
	}
	
	encoded, err := po.ToBytes()
	if err != nil {
		t.Fatalf("Failed to encode PrefixedOptional[FixedByteArray]: %v", err)
	}
	
	if len(encoded) != 257 {
		t.Errorf("Encoded length mismatch: got %d, want 257", len(encoded))
	}
	
	var decoded ns.PrefixedOptional[ns.FixedByteArray]
	decoded.Value = ns.NewFixedByteArray(256)
	bytesRead, err := decoded.FromBytes(encoded)
	if err != nil {
		t.Fatalf("Failed to decode PrefixedOptional[FixedByteArray]: %v", err)
	}
	
	if bytesRead != 257 {
		t.Errorf("Bytes read mismatch: got %d, want 257", bytesRead)
	}
	
	if !decoded.Present {
		t.Errorf("Expected Present to be true")
	}
	
	for i, b := range decoded.Value.Data {
		if b != po.Value.Data[i] {
			t.Errorf("Decoded data mismatch at %d: got %02x, want %02x", i, b, po.Value.Data[i])
			break
		}
	}
	
	po2 := ns.PrefixedOptional[ns.FixedByteArray]{
		Present: false,
	}
	
	encoded2, err := po2.ToBytes()
	if err != nil {
		t.Fatalf("Failed to encode PrefixedOptional with Present=false: %v", err)
	}
	
	if len(encoded2) != 1 {
		t.Errorf("Encoded length mismatch for absent optional: got %d, want 1", len(encoded2))
	}
}

// Test PrefixedArray[Byte] behaves like PrefixedByteArray
func TestPrefixedArrayByte(t *testing.T) {
	original := []byte("Hello, World!")
	
	pa := make(ns.PrefixedArray[ns.Byte], len(original))
	for i, b := range original {
		pa[i] = ns.Byte(b)
	}
	
	encoded, err := pa.ToBytes()
	if err != nil {
		t.Fatalf("Failed to encode PrefixedArray[Byte]: %v", err)
	}
	
	var decoded ns.PrefixedArray[ns.Byte]
	bytesRead, err := decoded.FromBytes(encoded)
	if err != nil {
		t.Fatalf("Failed to decode PrefixedArray[Byte]: %v", err)
	}
	
	if bytesRead != len(encoded) {
		t.Errorf("Bytes read mismatch: got %d, want %d", bytesRead, len(encoded))
	}
	
	result := make([]byte, len(decoded))
	for i, b := range decoded {
		result[i] = byte(b)
	}
	
	for i, b := range result {
		if b != original[i] {
			t.Errorf("Decoded data mismatch at %d: got %02x, want %02x", i, b, original[i])
			break
		}
	}
}

// Test NBT ToMap/FromMap functionality
func TestNBTMapMethods(t *testing.T) {
	testData := map[string]any{
		"text":  "Hello World",
		"color": "red",
		"extra": []any{
			map[string]any{"text": " with extra", "color": "blue"},
		},
	}
	
	var nbt ns.NBT
	err := nbt.FromMap(testData)
	if err != nil {
		t.Fatalf("Failed to create NBT from map: %v", err)
	}
	
	resultMap, err := nbt.ToMap()
	if err != nil {
		t.Fatalf("Failed to convert NBT to map: %v", err)
	}
	
	if resultMap["text"] != "Hello World" {
		t.Errorf("Expected text 'Hello World', got %v", resultMap["text"])
	}
	if resultMap["color"] != "red" {
		t.Errorf("Expected color 'red', got %v", resultMap["color"])
	}
	
	t.Logf("NBT ToMap/FromMap test passed: %+v", resultMap)
}

// Test TextComponent ToMap/FromMap functionality
func TestTextComponentMapMethods(t *testing.T) {
	testData := map[string]any{
		"text":  "Hello World",
		"color": "green",
		"bold":  true,
	}
	
	tc, err := ns.NewTextComponentFromMap(testData)
	if err != nil {
		t.Fatalf("Failed to create TextComponent from map: %v", err)
	}
	
	extractedText := tc.GetText()
	if extractedText != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", extractedText)
	}
	
	resultMap := tc.ToMapSafe()
	if resultMap["text"] != "Hello World" {
		t.Errorf("Expected text 'Hello World' in map, got %v", resultMap["text"])
	}
	
	encoded, err := tc.ToBytes()
	if err != nil {
		t.Fatalf("Failed to encode TextComponent: %v", err)
	}
	
	var decoded ns.TextComponent
	_, err = decoded.FromBytes(encoded)
	if err != nil {
		t.Fatalf("Failed to decode TextComponent: %v", err)
	}
	
	if decoded.GetText() != "Hello World" {
		t.Errorf("Round-trip failed: expected 'Hello World', got '%s'", decoded.GetText())
	}
	
	t.Logf("TextComponent round-trip test passed")
}

// Test creating TextComponent from simple string
func TestTextComponentFromString(t *testing.T) {
	tc := ns.NewTextComponentFromString("Simple text")
	
	text := tc.GetText()
	if text != "Simple text" {
		t.Errorf("Expected 'Simple text', got '%s'", text)
	}
	
	textMap := tc.ToMapSafe()
	if textMap["text"] != "Simple text" {
		t.Errorf("Expected map to have text='Simple text', got %v", textMap)
	}
}
