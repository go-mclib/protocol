package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

func TestUUIDFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ns.UUID
		wantErr  bool
	}{
		{
			"hyphenated",
			"550e8400-e29b-41d4-a716-446655440000",
			ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
			false,
		},
		{
			"no hyphens",
			"550e8400e29b41d4a716446655440000",
			ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
			false,
		},
		{
			"nil uuid",
			"00000000-0000-0000-0000-000000000000",
			ns.NilUUID,
			false,
		},
		{
			"invalid length",
			"550e8400",
			ns.UUID{},
			true,
		},
		{
			"invalid hex",
			"550e8400-e29b-41d4-a716-44665544000g",
			ns.UUID{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ns.UUIDFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UUIDFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("UUIDFromString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUUIDString(t *testing.T) {
	uuid := ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}
	expected := "550e8400-e29b-41d4-a716-446655440000"

	if got := uuid.String(); got != expected {
		t.Errorf("UUID.String() = %q, want %q", got, expected)
	}
}

func TestUUIDFromInt64s(t *testing.T) {
	// UUID: 550e8400-e29b-41d4-a716-446655440000
	// MSB: 0x550e8400e29b41d4 = 6124895493223874004
	// LSB: 0xa716446655440000 = -6386371178661429248
	msb := int64(0x550e8400e29b41d4)
	lsb := int64(-0x58e9bb99aabbffff - 1) // 0xa716446655440000 in two's complement

	uuid := ns.UUIDFromInt64s(msb, lsb)
	expected := ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}

	if uuid != expected {
		t.Errorf("UUIDFromInt64s() = %v, want %v", uuid, expected)
	}

	// Verify round-trip
	if uuid.MostSignificantBits() != msb {
		t.Errorf("MostSignificantBits() = %d, want %d", uuid.MostSignificantBits(), msb)
	}
	if uuid.LeastSignificantBits() != lsb {
		t.Errorf("LeastSignificantBits() = %d, want %d", uuid.LeastSignificantBits(), lsb)
	}
}

func TestUUIDReadWrite(t *testing.T) {
	uuid, _ := ns.UUIDFromString("550e8400-e29b-41d4-a716-446655440000")
	expectedBytes := []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}

	// Write
	buf := ns.NewWriter()
	if err := buf.WriteUUID(uuid); err != nil {
		t.Fatalf("WriteUUID() error = %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("WriteUUID() = %v, want %v", buf.Bytes(), expectedBytes)
	}

	// Read
	reader := ns.NewReader(expectedBytes)
	got, err := reader.ReadUUID()
	if err != nil {
		t.Fatalf("ReadUUID() error = %v", err)
	}
	if got != uuid {
		t.Errorf("ReadUUID() = %v, want %v", got, uuid)
	}
}

func TestUUIDIsNil(t *testing.T) {
	nilUUID := ns.UUID{}
	if !nilUUID.IsNil() {
		t.Error("IsNil() should return true for nil UUID")
	}

	nonNilUUID, _ := ns.UUIDFromString("550e8400-e29b-41d4-a716-446655440000")
	if nonNilUUID.IsNil() {
		t.Error("IsNil() should return false for non-nil UUID")
	}
}

func TestUUIDRoundTrip(t *testing.T) {
	uuids := []string{
		"00000000-0000-0000-0000-000000000000",
		"550e8400-e29b-41d4-a716-446655440000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"12345678-1234-5678-1234-567812345678",
	}

	for _, s := range uuids {
		t.Run(s, func(t *testing.T) {
			uuid, err := ns.UUIDFromString(s)
			if err != nil {
				t.Fatalf("UUIDFromString() error = %v", err)
			}

			buf := ns.NewWriter()
			if err := buf.WriteUUID(uuid); err != nil {
				t.Fatalf("WriteUUID() error = %v", err)
			}

			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadUUID()
			if err != nil {
				t.Fatalf("ReadUUID() error = %v", err)
			}

			if got != uuid {
				t.Errorf("RoundTrip: %v -> %v", uuid, got)
			}
		})
	}
}
