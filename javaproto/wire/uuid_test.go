package wire_test

import (
	"bytes"
	"encoding/json"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// UUID wire format: 16 bytes (128-bit, big-endian)

var uuidTestCases = []struct {
	name  string
	raw   []byte
	str   string
	value ns.UUID
}{
	{
		name:  "nil",
		raw:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		str:   "00000000-0000-0000-0000-000000000000",
		value: ns.UUID{},
	},
	{
		name:  "standard",
		raw:   []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
		str:   "550e8400-e29b-41d4-a716-446655440000",
		value: ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
	},
	{
		name:  "all ones",
		raw:   []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		str:   "ffffffff-ffff-ffff-ffff-ffffffffffff",
		value: ns.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	},
}

func TestUUID(t *testing.T) {
	for _, tc := range uuidTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			got, err := ns.NewReader(tc.raw).ReadUUID()
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %v, want %v", got, tc.value)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteUUID(tc.value); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("got %x, want %x", buf.Bytes(), tc.raw)
			}
		})

		t.Run(tc.name+" from string", func(t *testing.T) {
			got, err := ns.UUIDFromString(tc.str)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got != tc.value {
				t.Errorf("got %v, want %v", got, tc.value)
			}
		})

		t.Run(tc.name+" to string", func(t *testing.T) {
			if got := tc.value.String(); got != tc.str {
				t.Errorf("got %q, want %q", got, tc.str)
			}
		})
	}
}

func TestUUID_ParseErrors(t *testing.T) {
	invalid := []string{
		"550e8400",                              // too short
		"550e8400-e29b-41d4-a716-44665544000g",  // invalid hex
		"550e8400-e29b-41d4-a716-4466554400000", // too long
		"550e8400e-29b-41d4-a716-446655440000",  // misplaced hyphen
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",  // non-hex
		"550e8400-e29b41d4a716446655440000",     // partial hyphens
	}
	for _, s := range invalid {
		if _, err := ns.UUIDFromString(s); err == nil {
			t.Errorf("UUIDFromString(%q) should error", s)
		}
	}
}

func TestUUID_IsNil(t *testing.T) {
	if !(ns.UUID{}).IsNil() {
		t.Error("zero UUID.IsNil() should be true")
	}
	nonNil := ns.UUID{0x01}
	if nonNil.IsNil() {
		t.Error("non-nil UUID.IsNil() should be false")
	}
}

func TestUUID_Text(t *testing.T) {
	want := ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}
	data, err := json.Marshal(struct{ ID ns.UUID }{want})
	if err != nil || string(data) != `{"ID":"550e8400-e29b-41d4-a716-446655440000"}` {
		t.Fatalf("marshal = %s, %v", data, err)
	}
	for _, raw := range []string{`"550e8400-e29b-41d4-a716-446655440000"`, `"550e8400e29b41d4a716446655440000"`} {
		var got ns.UUID
		if err := json.Unmarshal([]byte(raw), &got); err != nil || got != want {
			t.Errorf("unmarshal %s = %v, %v", raw, got, err)
		}
	}
	for _, raw := range []string{`""`, `"zz"`, `[85,14,132,0,226,155,65,212,167,22,68,102,85,68,0,0]`} {
		var got ns.UUID
		if err := json.Unmarshal([]byte(raw), &got); err == nil {
			t.Errorf("unmarshal %s accepted", raw)
		}
	}
}

func TestOfflinePlayerUUID(t *testing.T) {
	for name, want := range map[string]string{
		"Notch": "b50ad385-829d-3141-a216-7e7d7539ba7f",
		"Steve": "5627dd98-e6be-3c21-b8a8-e92344183641",
	} {
		if got := ns.OfflinePlayerUUID(name).String(); got != want {
			t.Fatalf("OfflinePlayerUUID(%q) = %s, want %s", name, got, want)
		}
	}
}
