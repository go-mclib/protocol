package wire

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// UUID is a 128-bit universally unique identifier.
//
// Encoded as two big-endian 64-bit integers (most significant bits first).
type UUID [16]byte

// UUIDFromString parses a UUID from its string representation.
// Accepts formats: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
func UUIDFromString(s string) (UUID, error) {
	digits := s
	switch len(s) {
	case 32:
	case 36:
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			return UUID{}, fmt.Errorf("invalid UUID string %q", s)
		}
		digits = s[:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
	default:
		return UUID{}, fmt.Errorf("invalid UUID string length: %d", len(s))
	}

	var u UUID
	if _, err := hex.Decode(u[:], []byte(digits)); err != nil {
		return UUID{}, fmt.Errorf("invalid UUID hex: %w", err)
	}
	return u, nil
}

// String returns the UUID in standard hyphenated format.
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// IsNil returns true if this is the nil UUID (all zeros).
func (u UUID) IsNil() bool {
	return u == UUID{}
}

// MarshalText returns the hyphenated form, so JSON and other text encodings
// carry UUIDs as strings.
func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText parses the hyphenated or the 32-digit hex form.
func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := UUIDFromString(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// OfflinePlayerUUID returns Java's name-based UUID for OfflinePlayer:<name>,
// as offline-mode servers assign it (UUIDUtil.createOfflinePlayerUUID).
func OfflinePlayerUUID(name string) UUID {
	u := UUID(md5.Sum([]byte("OfflinePlayer:" + name)))
	u[6] = u[6]&0x0f | 0x30
	u[8] = u[8]&0x3f | 0x80
	return u
}
