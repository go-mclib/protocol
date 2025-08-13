package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestVarLong(t *testing.T) {
	tests := []struct {
		name string
		val  ns.VarLong
		want []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"two", 2, []byte{0x02}},
		{"127", 127, []byte{0x7f}},
		{"128", 128, []byte{0x80, 0x01}},
		{"255", 255, []byte{0xff, 0x01}},
		{"max", 9223372036854775807, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{"min", -9223372036854775808, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{"minus one", -1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("VarLong.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("VarLong.Marshal() = %x, want %x", got, tt.want)
			}

			var unmarshaled ns.VarLong
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalVarLong() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("VarLong.Unmarshal() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestVarLongLen(t *testing.T) {
	tests := []struct {
		name string
		val  ns.VarLong
		want int
	}{
		{"zero", 0, 1},
		{"one", 1, 1},
		{"two", 2, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"max", 9223372036854775807, 9},
		{"min", -9223372036854775808, 10},
		{"minus one", -1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.Len()
			if got != tt.want {
				t.Errorf("VarLong.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVarLongErrorCases(t *testing.T) {
	t.Run("VarLong too big", func(t *testing.T) {
		// VarLong with too many bytes
		data := ns.ByteArray{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		var temp ns.VarLong
		_, err := temp.FromBytes(data)
		if err == nil {
			t.Error("VarLong.Unmarshal() should error on too many bytes")
		}
	})
}

func TestVarLongInterface(t *testing.T) {
	val := ns.VarLong(9223372036854775807)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("VarLong.Marshal() error = %v", err)
	}

	var result ns.VarLong
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("VarLong.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("VarLong interface roundtrip: got %v, want %v", result, val)
	}
}

func TestVarLongGenericMarshal(t *testing.T) {
	val := ns.VarLong(9223372036854775807)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Marshal(VarLong) error = %v", err)
	}

	var result ns.VarLong
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Unmarshal(VarLong) error = %v", err)
	}
	if result != val {
		t.Errorf("Generic VarLong roundtrip: got %v, want %v", result, val)
	}
}
