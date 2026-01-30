package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

func TestStringReadWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.String
		expected []byte
	}{
		{"empty", "", []byte{0x00}},
		{"hello", "hello", []byte{0x05, 'h', 'e', 'l', 'l', 'o'}},
		{"minecraft", "minecraft", []byte{0x09, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't'}},
		// UTF-8 multi-byte: "é" is 2 bytes (0xc3 0xa9)
		{"café", "café", []byte{0x05, 'c', 'a', 'f', 0xc3, 0xa9}},
		// UTF-8: "日本" is 6 bytes
		{"japanese", "日本", []byte{0x06, 0xe6, 0x97, 0xa5, 0xe6, 0x9c, 0xac}},
	}

	for _, tt := range tests {
		t.Run(tt.name+" write", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteString(tt.value); err != nil {
				t.Fatalf("WriteString() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteString(%q) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}
		})

		t.Run(tt.name+" read", func(t *testing.T) {
			buf := ns.NewReader(tt.expected)
			got, err := buf.ReadString(0)
			if err != nil {
				t.Fatalf("ReadString() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadString() = %q, want %q", got, tt.value)
			}
		})
	}
}

func TestStringMaxLength(t *testing.T) {
	// String with 5 characters, but we set max to 3
	data := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	buf := ns.NewReader(data)
	_, err := buf.ReadString(3)
	if err == nil {
		t.Error("ReadString() should error when exceeding max length")
	}
}

func TestStringRoundTrip(t *testing.T) {
	values := []ns.String{
		"",
		"a",
		"hello world",
		"player_123",
		"minecraft:stone",
		"日本語テスト",
		"emoji: 🎮",
	}

	for _, v := range values {
		t.Run(string(v), func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteString(v); err != nil {
				t.Fatalf("WriteString() error = %v", err)
			}

			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadString(0)
			if err != nil {
				t.Fatalf("ReadString() error = %v", err)
			}

			if got != v {
				t.Errorf("RoundTrip: wrote %q, got %q", v, got)
			}
		})
	}
}

func TestIdentifierReadWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    ns.Identifier
		expected []byte
	}{
		{"stone", "minecraft:stone", []byte{0x0f, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 's', 't', 'o', 'n', 'e'}},
		{"custom", "custom:item", []byte{0x0b, 'c', 'u', 's', 't', 'o', 'm', ':', 'i', 't', 'e', 'm'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteIdentifier(tt.value); err != nil {
				t.Fatalf("WriteIdentifier() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("WriteIdentifier(%q) = %v, want %v", tt.value, buf.Bytes(), tt.expected)
			}

			reader := ns.NewReader(tt.expected)
			got, err := reader.ReadIdentifier()
			if err != nil {
				t.Fatalf("ReadIdentifier() error = %v", err)
			}
			if got != tt.value {
				t.Errorf("ReadIdentifier() = %q, want %q", got, tt.value)
			}
		})
	}
}

func TestIdentifierNamespacePath(t *testing.T) {
	tests := []struct {
		id        ns.Identifier
		namespace string
		path      string
	}{
		{"minecraft:stone", "minecraft", "stone"},
		{"custom:my_item", "custom", "my_item"},
		{"stone", "minecraft", "stone"}, // default namespace
		{"minecraft:textures/block/stone.png", "minecraft", "textures/block/stone.png"},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			if got := tt.id.Namespace(); got != tt.namespace {
				t.Errorf("Namespace() = %q, want %q", got, tt.namespace)
			}
			if got := tt.id.Path(); got != tt.path {
				t.Errorf("Path() = %q, want %q", got, tt.path)
			}
		})
	}
}
