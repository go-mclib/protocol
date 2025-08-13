package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestVarInt(t *testing.T) {
	tests := []struct {
		name string
		val  ns.VarInt
		want []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"two", 2, []byte{0x02}},
		{"127", 127, []byte{0x7f}},
		{"128", 128, []byte{0x80, 0x01}},
		{"255", 255, []byte{0xff, 0x01}},
		{"256", 256, []byte{0x80, 0x02}},
		{"25565", 25565, []byte{0xdd, 0xc7, 0x01}},
		{"max", 2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		{"min", -2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
		{"minus one", -1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("VarInt.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("VarInt.Marshal() = %x, want %x", got, tt.want)
			}

			var unmarshaled ns.VarInt
			bytesRead, err := unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalVarInt() error = %v", err)
			}
			if bytesRead != len(got) {
				t.Errorf("VarInt.FromBytes() consumed %d bytes, want %d", bytesRead, len(got))
			}
			if unmarshaled != tt.val {
				t.Errorf("VarInt.Unmarshal() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestVarIntLen(t *testing.T) {
	tests := []struct {
		name string
		val  ns.VarInt
		want int
	}{
		{"zero", 0, 1},
		{"one", 1, 1},
		{"two", 2, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"max", 2147483647, 5},
		{"min", -2147483648, 5},
		{"minus one", -1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.Len()
			if got != tt.want {
				t.Errorf("VarInt.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVarIntErrorCases(t *testing.T) {
	t.Run("VarInt too big", func(t *testing.T) {
		// VarInt with too many bytes
		data := ns.ByteArray{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		var temp ns.VarInt
		_, err := temp.FromBytes(data)
		if err == nil {
			t.Error("VarInt.Unmarshal() should error on too many bytes")
		}
	})
}

func TestVarIntInterface(t *testing.T) {
	val := ns.VarInt(25565)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("VarInt.Marshal() error = %v", err)
	}

	var result ns.VarInt
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("VarInt.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("VarInt interface roundtrip: got %v, want %v", result, val)
	}
}

func TestVarIntGenericMarshal(t *testing.T) {
	val := ns.VarInt(25565)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Marshal(VarInt) error = %v", err)
	}

	var result ns.VarInt
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Unmarshal(VarInt) error = %v", err)
	}
	if result != val {
		t.Errorf("Generic VarInt roundtrip: got %v, want %v", result, val)
	}
}
