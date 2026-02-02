package net_structures

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LpVec3 is a low-precision 3D vector used for entity velocity.
// Encodes 3 doubles in typically 6 bytes using 15-bit scaled values.
//
// Wire format (if any component has |value| >= 3.05e-5):
//
//	┌────────────────────────────────────────────────────────────┐
//	│  First 2 bytes (little-endian): scale + X low bits         │
//	│  Last 4 bytes (big-endian): X high + Y + Z bits            │
//	└────────────────────────────────────────────────────────────┘
//
// If all components are essentially zero (< 3.05e-5), sends single 0x00 byte.
type LpVec3 struct {
	X, Y, Z float64
}

// threshold below which values are considered zero
const lpThreshold = 3.0517578125e-5 // 1 / 32768

// Decode reads an LpVec3 from the buffer.
func (v *LpVec3) Decode(buf *PacketBuffer) error {
	// read first byte
	first, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read lpvec3 first byte: %w", err)
	}

	// zero vector
	if first == 0 {
		v.X, v.Y, v.Z = 0, 0, 0
		return nil
	}

	// read second byte (little-endian pair with first)
	second, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read lpvec3 second byte: %w", err)
	}

	// read remaining 4 bytes (big-endian)
	var rest [4]byte
	if _, err := buf.Read(rest[:]); err != nil {
		return fmt.Errorf("failed to read lpvec3 rest: %w", err)
	}

	// parse 48 bits total:
	// first 2 bytes are little-endian, next 4 are big-endian
	le16 := uint16(first) | uint16(second)<<8
	be32 := binary.BigEndian.Uint32(rest[:])

	// extract fields:
	// scale: 3 bits from le16 bits 0-2
	// x: 15 bits from le16 bits 3-15 (13 bits) + be32 bits 30-31 (2 bits)
	// y: 15 bits from be32 bits 15-29
	// z: 15 bits from be32 bits 0-14
	scale := int(le16 & 0x7)
	xLow := (le16 >> 3) & 0x1FFF  // 13 bits
	xHigh := (be32 >> 30) & 0x3   // 2 bits
	x := int16((xHigh << 13) | uint32(xLow))
	y := int16((be32 >> 15) & 0x7FFF)
	z := int16(be32 & 0x7FFF)

	// sign extend 15-bit values to 16-bit
	if x&0x4000 != 0 {
		x |= -0x4000
	}
	if y&0x4000 != 0 {
		y |= -0x4000
	}
	if z&0x4000 != 0 {
		z |= -0x4000
	}

	// decode using scale factor
	mult := math.Pow(2, float64(scale)) * lpThreshold
	v.X = float64(x) * mult
	v.Y = float64(y) * mult
	v.Z = float64(z) * mult

	return nil
}

// Encode writes an LpVec3 to the buffer.
func (v *LpVec3) Encode(buf *PacketBuffer) error {
	// check if essentially zero
	maxAbs := math.Max(math.Abs(v.X), math.Max(math.Abs(v.Y), math.Abs(v.Z)))
	if maxAbs < lpThreshold {
		return buf.WriteByte(0)
	}

	// determine scale factor (0-7)
	scale := 0
	for s := range 8 {
		maxVal := math.Pow(2, float64(s)) * lpThreshold * 16383 // 15-bit signed max
		if maxAbs <= maxVal {
			scale = s
			break
		}
		if s == 7 {
			scale = 7 // clamp to max
		}
	}

	// encode to 15-bit signed values
	mult := math.Pow(2, float64(scale)) * lpThreshold
	x := clampInt15(int16(math.Round(v.X / mult)))
	y := clampInt15(int16(math.Round(v.Y / mult)))
	z := clampInt15(int16(math.Round(v.Z / mult)))

	// pack into 48 bits:
	// le16: scale(3) + x_low(13)
	// be32: x_high(2) + y(15) + z(15)
	xVal := uint16(x) & 0x7FFF
	yVal := uint16(y) & 0x7FFF
	zVal := uint16(z) & 0x7FFF

	le16 := uint16(scale&0x7) | ((xVal & 0x1FFF) << 3)
	be32 := uint32(xVal>>13)<<30 | uint32(yVal)<<15 | uint32(zVal)

	// write little-endian pair
	if err := buf.WriteByte(byte(le16)); err != nil {
		return err
	}
	if err := buf.WriteByte(byte(le16 >> 8)); err != nil {
		return err
	}

	// write big-endian quad
	var rest [4]byte
	binary.BigEndian.PutUint32(rest[:], be32)
	_, err := buf.Write(rest[:])
	return err
}

func clampInt15(v int16) int16 {
	if v > 16383 {
		return 16383
	}
	if v < -16384 {
		return -16384
	}
	return v
}

// ReadLpVec3 reads an LpVec3 from the buffer.
func (pb *PacketBuffer) ReadLpVec3() (LpVec3, error) {
	var v LpVec3
	err := v.Decode(pb)
	return v, err
}

// WriteLpVec3 writes an LpVec3 to the buffer.
func (pb *PacketBuffer) WriteLpVec3(v LpVec3) error {
	return v.Encode(pb)
}
