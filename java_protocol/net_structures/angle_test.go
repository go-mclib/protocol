package net_structures_test

import (
	"math"
	"testing"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

func TestAngleFromDegrees(t *testing.T) {
	tests := []struct {
		degrees  float64
		expected ns.Angle
	}{
		{0, 0},
		{90, 64},
		{180, 128},
		{270, 192},
		{360, 0},   // wraps around
		{-90, 192}, // wraps to 270°
		{45, 32},
		{1.40625, 1}, // exactly 1 unit
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := ns.AngleFromDegrees(tt.degrees)
			if got != tt.expected {
				t.Errorf("AngleFromDegrees(%v) = %d, want %d", tt.degrees, got, tt.expected)
			}
		})
	}
}

func TestAngleDegrees(t *testing.T) {
	tests := []struct {
		angle    ns.Angle
		expected float64
	}{
		{0, 0},
		{64, 90},
		{128, 180},
		{192, 270},
		{255, 358.59375}, // 255 * 360 / 256
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.angle.Degrees()
			if math.Abs(got-tt.expected) > 0.0001 {
				t.Errorf("Angle(%d).Degrees() = %v, want %v", tt.angle, got, tt.expected)
			}
		})
	}
}

func TestAngleRadians(t *testing.T) {
	tests := []struct {
		angle    ns.Angle
		expected float64
	}{
		{0, 0},
		{64, math.Pi / 2},      // 90°
		{128, math.Pi},         // 180°
		{192, 3 * math.Pi / 2}, // 270°
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.angle.Radians()
			if math.Abs(got-tt.expected) > 0.0001 {
				t.Errorf("Angle(%d).Radians() = %v, want %v", tt.angle, got, tt.expected)
			}
		})
	}
}

func TestAngleReadWrite(t *testing.T) {
	angles := []ns.Angle{0, 64, 128, 192, 255}

	for _, a := range angles {
		t.Run("", func(t *testing.T) {
			buf := ns.NewWriter()
			if err := buf.WriteAngle(a); err != nil {
				t.Fatalf("WriteAngle() error = %v", err)
			}

			if len(buf.Bytes()) != 1 {
				t.Errorf("WriteAngle() produced %d bytes, want 1", len(buf.Bytes()))
			}

			if buf.Bytes()[0] != byte(a) {
				t.Errorf("WriteAngle(%d) = %v, want %v", a, buf.Bytes()[0], byte(a))
			}

			reader := ns.NewReader(buf.Bytes())
			got, err := reader.ReadAngle()
			if err != nil {
				t.Fatalf("ReadAngle() error = %v", err)
			}

			if got != a {
				t.Errorf("ReadAngle() = %d, want %d", got, a)
			}
		})
	}
}
