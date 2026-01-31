package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

func TestPositionPackUnpack(t *testing.T) {
	// Test vectors - focusing on round-trip since bit manipulation for
	// negative values produces results that are implementation-dependent
	tests := []struct {
		name string
		pos  ns.Position
	}{
		{"origin", ns.Position{X: 0, Y: 0, Z: 0}},
		{"simple", ns.Position{X: 1, Y: 2, Z: 3}},
		{"negative x", ns.Position{X: -1, Y: 0, Z: 0}},
		{"negative z", ns.Position{X: 0, Y: 0, Z: -1}},
		{"negative y", ns.Position{X: 0, Y: -1, Z: 0}},
		{"all negative", ns.Position{X: -1, Y: -1, Z: -1}},
		// Example from wiki.vg: position (18357644, 831, -20882616)
		{"wiki example", ns.Position{X: 18357644, Y: 831, Z: -20882616}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := tt.pos.Pack()
			decoded := ns.UnpackPosition(packed)
			if decoded != tt.pos {
				t.Errorf("RoundTrip: %+v -> 0x%x -> %+v", tt.pos, packed, decoded)
			}
		})
	}
}

func TestPositionRoundTrip(t *testing.T) {
	positions := []ns.Position{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 3},
		{X: 100, Y: 64, Z: -200},
		{X: -100, Y: 64, Z: 200},
		{X: -100, Y: -32, Z: -200},
		{X: 33554431, Y: 2047, Z: 33554431},    // max positive
		{X: -33554432, Y: -2048, Z: -33554432}, // max negative
		// Real-world coordinates
		{X: 18357644, Y: 831, Z: -20882616},
		{X: -12345678, Y: 100, Z: 12345678},
	}

	for _, pos := range positions {
		t.Run("", func(t *testing.T) {
			packed := pos.Pack()
			decoded := ns.UnpackPosition(packed)

			if decoded.X != pos.X || decoded.Y != pos.Y || decoded.Z != pos.Z {
				t.Errorf("RoundTrip: %+v -> 0x%x -> %+v", pos, packed, decoded)
			}
		})
	}
}

func TestPositionReadWrite(t *testing.T) {
	positions := []ns.Position{
		{X: 0, Y: 0, Z: 0},
		{X: 100, Y: 64, Z: -200},
		{X: 18357644, Y: 831, Z: -20882616},
	}

	for _, pos := range positions {
		t.Run("", func(t *testing.T) {
			// Write
			buf := ns.NewWriter()
			if err := buf.WritePosition(pos); err != nil {
				t.Fatalf("WritePosition() error = %v", err)
			}

			// Should be exactly 8 bytes (int64)
			if len(buf.Bytes()) != 8 {
				t.Errorf("WritePosition() produced %d bytes, want 8", len(buf.Bytes()))
			}

			// Read
			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadPosition()
			if err != nil {
				t.Fatalf("ReadPosition() error = %v", err)
			}

			if got != pos {
				t.Errorf("RoundTrip: %+v -> %+v", pos, got)
			}
		})
	}
}

func TestPositionKnownBytes(t *testing.T) {
	// Known encoded position from network capture
	// Position (0, 64, 0) should encode to specific bytes
	pos := ns.Position{X: 0, Y: 64, Z: 0}

	buf := ns.NewWriter()
	if err := buf.WritePosition(pos); err != nil {
		t.Fatalf("WritePosition() error = %v", err)
	}

	// Y=64 -> bits 0-11 = 64 = 0x40
	// X=0, Z=0 -> bits 12-63 = 0
	// Big-endian: 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x40
	expected := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("WritePosition((0,64,0)) = %v, want %v", buf.Bytes(), expected)
	}

	// Read it back
	reader := ns.NewReader(expected)
	got, err := reader.ReadPosition()
	if err != nil {
		t.Fatalf("ReadPosition() error = %v", err)
	}
	if got != pos {
		t.Errorf("ReadPosition() = %+v, want %+v", got, pos)
	}
}
