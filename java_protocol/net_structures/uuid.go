package net_structures

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// UUID is a 128-bit universally unique identifier.
//
// Encoded as two big-endian 64-bit integers (most significant bits first).
type UUID [16]byte

// NilUUID is the zero UUID (all zeros).
var NilUUID = UUID{}

// Encode writes the UUID to w.
func (u UUID) Encode(w io.Writer) error {
	_, err := w.Write(u[:])
	return err
}

// DecodeUUID reads a UUID from r.
func DecodeUUID(r io.Reader) (UUID, error) {
	var u UUID
	if _, err := io.ReadFull(r, u[:]); err != nil {
		return UUID{}, err
	}
	return u, nil
}

// UUIDFromString parses a UUID from its string representation.
// Accepts formats: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
func UUIDFromString(s string) (UUID, error) {
	// Remove hyphens if present
	clean := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			clean = append(clean, s[i])
		}
	}

	if len(clean) != 32 {
		return UUID{}, fmt.Errorf("invalid UUID string length: %d", len(clean))
	}

	var u UUID
	_, err := hex.Decode(u[:], clean)
	if err != nil {
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
	return u == NilUUID
}

// ValidateUUID checks if a string is a valid UUID format.
// Accepts formats: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" (36 chars) or "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" (32 chars)
func ValidateUUID(uuid string) bool {
	if len(uuid) == 36 && strings.Count(uuid, "-") == 4 {
		return true
	}
	if len(uuid) == 32 && strings.Count(uuid, "-") == 0 {
		return true
	}
	return false
}
