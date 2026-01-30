package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

// Test vectors from wiki.vg/Protocol and manual calculation
// https://wiki.vg/Protocol#VarInt_and_VarLong

func TestVarIntEncode(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.VarInt
		expected []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"two", 2, []byte{0x02}},
		{"max single byte", 127, []byte{0x7f}},
		{"min two bytes", 128, []byte{0x80, 0x01}},
		{"255", 255, []byte{0xff, 0x01}},
		{"25565 (default MC port)", 25565, []byte{0xdd, 0xc7, 0x01}},
		{"2097151 (max 3 bytes)", 2097151, []byte{0xff, 0xff, 0x7f}},
		{"2147483647 (max int32)", 2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		{"negative one", -1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
		{"negative two", -2, []byte{0xfe, 0xff, 0xff, 0xff, 0x0f}},
		{"-2147483648 (min int32)", -2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("ToBytes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVarIntDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected ns.VarInt
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x01}, 1},
		{"max single byte", []byte{0x7f}, 127},
		{"min two bytes", []byte{0x80, 0x01}, 128},
		{"255", []byte{0xff, 0x01}, 255},
		{"25565", []byte{0xdd, 0xc7, 0x01}, 25565},
		{"2097151", []byte{0xff, 0xff, 0x7f}, 2097151},
		{"max int32", []byte{0xff, 0xff, 0xff, 0xff, 0x07}, 2147483647},
		{"negative one", []byte{0xff, 0xff, 0xff, 0xff, 0x0f}, -1},
		{"min int32", []byte{0x80, 0x80, 0x80, 0x80, 0x08}, -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewReader(tt.input)
			got, err := buf.ReadVarInt()
			if err != nil {
				t.Fatalf("ReadVarInt() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("ReadVarInt() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVarIntRoundTrip(t *testing.T) {
	values := []ns.VarInt{0, 1, 127, 128, 255, 256, 25565, 2097151, 2147483647, -1, -128, -2147483648}

	for _, v := range values {
		t.Run("", func(t *testing.T) {
			// Encode
			buf := ns.NewWriter()
			if err := buf.WriteVarInt(v); err != nil {
				t.Fatalf("WriteVarInt() error = %v", err)
			}

			// Decode
			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadVarInt()
			if err != nil {
				t.Fatalf("ReadVarInt() error = %v", err)
			}

			if got != v {
				t.Errorf("RoundTrip: wrote %v, got %v", v, got)
			}
		})
	}
}

func TestVarIntLen(t *testing.T) {
	tests := []struct {
		value    ns.VarInt
		expected int
	}{
		{0, 1},
		{127, 1},
		{128, 2},
		{16383, 2},
		{16384, 3},
		{2097151, 3},
		{2097152, 4},
		{268435455, 4},
		{268435456, 5},
		{2147483647, 5},
		{-1, 5},
	}

	for _, tt := range tests {
		got := tt.value.Len()
		if got != tt.expected {
			t.Errorf("VarInt(%d).Len() = %d, want %d", tt.value, got, tt.expected)
		}
	}
}

func TestVarIntTooLong(t *testing.T) {
	// 6 continuation bytes - invalid
	input := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	buf := ns.NewReader(input)
	_, err := buf.ReadVarInt()
	if err == nil {
		t.Error("ReadVarInt() should error on too many bytes")
	}
}

func TestVarLongEncode(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.VarLong
		expected []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"max single byte", 127, []byte{0x7f}},
		{"128", 128, []byte{0x80, 0x01}},
		{"max int64", 9223372036854775807, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{"negative one", -1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("ToBytes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVarLongDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected ns.VarLong
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x01}, 1},
		{"max single byte", []byte{0x7f}, 127},
		{"128", []byte{0x80, 0x01}, 128},
		{"max int64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, 9223372036854775807},
		{"negative one", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewReader(tt.input)
			got, err := buf.ReadVarLong()
			if err != nil {
				t.Fatalf("ReadVarLong() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("ReadVarLong() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVarLongRoundTrip(t *testing.T) {
	values := []ns.VarLong{0, 1, 127, 128, 255, 9223372036854775807, -1, -9223372036854775808}

	for _, v := range values {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteVarLong(v); err != nil {
				t.Fatalf("WriteVarLong() error = %v", err)
			}

			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadVarLong()
			if err != nil {
				t.Fatalf("ReadVarLong() error = %v", err)
			}

			if got != v {
				t.Errorf("RoundTrip: wrote %v, got %v", v, got)
			}
		})
	}
}
