package wire

// Position represents a block position in the world.
//
// Encoded as a 64-bit integer with the following bit layout:
//   - X: 26 bits (signed, bits 38-63)
//   - Z: 26 bits (signed, bits 12-37)
//   - Y: 12 bits (signed, bits 0-11)
//
// This allows coordinates:
//   - X, Z: -33554432 to 33554431
//   - Y: -2048 to 2047
type Position struct {
	X, Y, Z int
}

// Pack encodes the position into a 64-bit integer.
func (p Position) Pack() int64 {
	return ((int64(p.X) & 0x3FFFFFF) << 38) |
		((int64(p.Z) & 0x3FFFFFF) << 12) |
		(int64(p.Y) & 0xFFF)
}

// GlobalPos represents a position in a specific dimension.
// Used for things like death locations.
//
// Wire format:
//
//	┌─────────────────────────┬─────────────────────────┐
//	│  Dimension (Identifier)  │  Position (Int64)       │
//	└─────────────────────────┴─────────────────────────┘
type GlobalPos struct {
	Dimension Identifier
	Pos       Position
}

// unpackPosition decodes a 64-bit integer into a Position, sign-extending
// the 26-bit x/z and 12-bit y fields.
func unpackPosition(val int64) Position {
	x := int(val >> 38)
	z := int(val << 26 >> 38)
	y := int(val << 52 >> 52)

	if x >= 1<<25 {
		x -= 1 << 26
	}
	if z >= 1<<25 {
		z -= 1 << 26
	}
	if y >= 1<<11 {
		y -= 1 << 12
	}

	return Position{X: x, Y: y, Z: z}
}
