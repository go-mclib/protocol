package net_structures_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestNBT(t *testing.T) {
	t.Run("empty NBT", func(t *testing.T) {
		nbt := ns.NewEmptyNBT()

		if !nbt.IsEmpty() {
			t.Error("NewEmptyNBT() should create an empty NBT")
		}

		data, err := nbt.ToBytes()
		if err != nil {
			t.Errorf("NBT.ToBytes() error = %v", err)
		}

		if len(data) != 1 || data[0] != 0x00 {
			t.Errorf("Empty NBT should encode to [0x00], got %v", data)
		}

		var result ns.NBT
		bytesRead, err := result.FromBytes(data)
		if err != nil {
			t.Errorf("NBT.FromBytes() error = %v", err)
		}

		if bytesRead != 1 {
			t.Errorf("NBT.FromBytes() should read 1 byte, got %d", bytesRead)
		}

		if !result.IsEmpty() {
			t.Error("Decoded NBT should be empty")
		}
	})

	t.Run("simple NBT with string", func(t *testing.T) {
		data := map[string]any{
			"name":  "test",
			"value": int32(42),
		}

		nbt := ns.NewNBT(data)

		if nbt.IsEmpty() {
			t.Error("NBT with data should not be empty")
		}

		encoded, err := nbt.ToBytes()
		if err != nil {
			t.Errorf("NBT.ToBytes() error = %v", err)
		}

		if len(encoded) == 0 {
			t.Error("Encoded NBT should not be empty")
		}

		var result ns.NBT
		bytesRead, err := result.FromBytes(encoded)
		if err != nil {
			t.Errorf("NBT.FromBytes() error = %v", err)
		}

		if bytesRead != len(encoded) {
			t.Errorf("NBT.FromBytes() should read all bytes, got %d/%d", bytesRead, len(encoded))
		}

		if result.IsEmpty() {
			t.Error("Decoded NBT should not be empty")
		}
	})

	t.Run("NBT encode/decode specific type", func(t *testing.T) {
		type TestStruct struct {
			Name  string `nbt:"name"`
			Value int32  `nbt:"value"`
		}

		original := TestStruct{
			Name:  "hello",
			Value: 123,
		}
		nbt := ns.NewNBT(original)
		encoded, err := nbt.ToBytes()
		if err != nil {
			t.Errorf("NBT.ToBytes() error = %v", err)
		}
		var result ns.NBT
		_, err = result.FromBytes(encoded)
		if err != nil {
			t.Errorf("NBT.FromBytes() error = %v", err)
		}
		var decoded TestStruct
		err = result.DecodeTo(&decoded)
		if err != nil {
			t.Errorf("NBT.DecodeTo() error = %v", err)
		}

		if decoded.Name != original.Name || decoded.Value != original.Value {
			t.Errorf("Decoded struct mismatch: got %+v, want %+v", decoded, original)
		}
	})
}

func TestNBTErrorCases(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		var nbt ns.NBT
		_, err := nbt.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("NBT.FromBytes() should error on empty data")
		}
	})

	t.Run("decode to nil", func(t *testing.T) {
		nbt := ns.NewEmptyNBT()
		err := nbt.DecodeTo(nil)
		if err == nil {
			t.Error("NBT.DecodeTo() should error when NBT is empty")
		}
	})
}
