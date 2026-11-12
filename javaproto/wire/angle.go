package wire

import "math"

// Angle represents a rotation angle.
//
// Stored as a single byte where 256 units = 360 degrees (full rotation).
// Each unit represents 1/256 of a full turn (1.40625 degrees).
//
// Examples:
//
//	0   = 0°
//	64  = 90°
//	128 = 180°
//	192 = 270°
type Angle uint8

// AngleFromDegrees converts degrees to an Angle.
// Like vanilla's (byte) Mth.floor(degrees * 256 / 360), the result wraps;
// NaN and infinities map to 0.
func AngleFromDegrees(degrees float64) Angle {
	if math.IsNaN(degrees) || math.IsInf(degrees, 0) {
		return 0
	}
	return Angle(uint8(int64(math.Floor(math.Mod(degrees, 360) * 256 / 360))))
}

// Degrees converts the Angle to degrees (0-360).
func (a Angle) Degrees() float64 {
	return float64(a) * 360 / 256
}

// Radians converts the Angle to radians (0-2π).
func (a Angle) Radians() float64 {
	return float64(a) * 2 * math.Pi / 256
}
