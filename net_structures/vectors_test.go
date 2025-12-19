package net_structures

import (
	"testing"
)

func TestChunkPos_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		pos  ChunkPos
	}{
		{"origin", ChunkPos{X: 0, Z: 0}},
		{"positive", ChunkPos{X: 100, Z: 200}},
		{"negative", ChunkPos{X: -100, Z: -200}},
		{"mixed", ChunkPos{X: 100, Z: -200}},
		{"max", ChunkPos{X: 0x7FFFFFFF, Z: 0x7FFFFFFF}},
		{"min", ChunkPos{X: -0x80000000, Z: -0x80000000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.pos.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}

			// Decode
			var decoded ChunkPos
			bytesRead, err := decoded.FromBytes(encoded)
			if err != nil {
				t.Fatalf("FromBytes() error = %v", err)
			}

			// Verify
			if bytesRead != 8 {
				t.Errorf("expected 8 bytes read, got %d", bytesRead)
			}
			if decoded.X != tt.pos.X || decoded.Z != tt.pos.Z {
				t.Errorf("roundtrip failed: got (%d, %d), want (%d, %d)",
					decoded.X, decoded.Z, tt.pos.X, tt.pos.Z)
			}
		})
	}
}

func TestVec3_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		vec  Vec3
	}{
		{"zero", Vec3{X: 0, Y: 0, Z: 0}},
		{"positive", Vec3{X: 1.5, Y: 2.5, Z: 3.5}},
		{"negative", Vec3{X: -1.5, Y: -2.5, Z: -3.5}},
		{"large", Vec3{X: 1e10, Y: 2e10, Z: 3e10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.vec.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}

			// Decode
			var decoded Vec3
			bytesRead, err := decoded.FromBytes(encoded)
			if err != nil {
				t.Fatalf("FromBytes() error = %v", err)
			}

			// Verify
			if bytesRead != 24 {
				t.Errorf("expected 24 bytes read, got %d", bytesRead)
			}
			if decoded.X != tt.vec.X || decoded.Y != tt.vec.Y || decoded.Z != tt.vec.Z {
				t.Errorf("roundtrip failed: got %+v, want %+v", decoded, tt.vec)
			}
		})
	}
}

func TestVector3f_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		vec  Vector3f
	}{
		{"zero", Vector3f{X: 0, Y: 0, Z: 0}},
		{"positive", Vector3f{X: 1.5, Y: 2.5, Z: 3.5}},
		{"negative", Vector3f{X: -1.5, Y: -2.5, Z: -3.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.vec.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}

			// Decode
			var decoded Vector3f
			bytesRead, err := decoded.FromBytes(encoded)
			if err != nil {
				t.Fatalf("FromBytes() error = %v", err)
			}

			// Verify
			if bytesRead != 12 {
				t.Errorf("expected 12 bytes read, got %d", bytesRead)
			}
			if decoded.X != tt.vec.X || decoded.Y != tt.vec.Y || decoded.Z != tt.vec.Z {
				t.Errorf("roundtrip failed: got %+v, want %+v", decoded, tt.vec)
			}
		})
	}
}

func TestQuaternionf_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		quat Quaternionf
	}{
		{"identity", Quaternionf{X: 0, Y: 0, Z: 0, W: 1}},
		{"arbitrary", Quaternionf{X: 0.5, Y: 0.5, Z: 0.5, W: 0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.quat.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}

			// Decode
			var decoded Quaternionf
			bytesRead, err := decoded.FromBytes(encoded)
			if err != nil {
				t.Fatalf("FromBytes() error = %v", err)
			}

			// Verify
			if bytesRead != 16 {
				t.Errorf("expected 16 bytes read, got %d", bytesRead)
			}
			if decoded != tt.quat {
				t.Errorf("roundtrip failed: got %+v, want %+v", decoded, tt.quat)
			}
		})
	}
}

func TestGlobalPos_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		pos  GlobalPos
	}{
		{
			"overworld",
			GlobalPos{
				Dimension: "minecraft:overworld",
				Pos:       Position{X: 100, Y: 64, Z: -200},
			},
		},
		{
			"nether",
			GlobalPos{
				Dimension: "minecraft:the_nether",
				Pos:       Position{X: 0, Y: 128, Z: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.pos.ToBytes()
			if err != nil {
				t.Fatalf("ToBytes() error = %v", err)
			}

			// Decode
			var decoded GlobalPos
			bytesRead, err := decoded.FromBytes(encoded)
			if err != nil {
				t.Fatalf("FromBytes() error = %v", err)
			}

			// Verify
			if bytesRead != len(encoded) {
				t.Errorf("expected %d bytes read, got %d", len(encoded), bytesRead)
			}
			if decoded.Dimension != tt.pos.Dimension {
				t.Errorf("dimension mismatch: got %s, want %s", decoded.Dimension, tt.pos.Dimension)
			}
			if decoded.Pos != tt.pos.Pos {
				t.Errorf("position mismatch: got %+v, want %+v", decoded.Pos, tt.pos.Pos)
			}
		})
	}
}
