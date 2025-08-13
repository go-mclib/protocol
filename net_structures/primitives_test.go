package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/go-mclib/protocol/net_structures"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Boolean
		want []byte
	}{
		{"true", true, []byte{0x01}},
		{"false", false, []byte{0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Boolean.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Boolean.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.Boolean
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalBoolean() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("Boolean.Unmarshal() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestByte(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Byte
		want []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"positive", 127, []byte{0x7F}},
		{"negative", -128, []byte{0x80}},
		{"minus one", -1, []byte{0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Byte.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Byte.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.Byte
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalByte() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("Byte.Unmarshal() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestUnsignedByte(t *testing.T) {
	tests := []struct {
		name string
		val  ns.UnsignedByte
		want []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"max", 255, []byte{0xFF}},
		{"middle", 128, []byte{0x80}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("UnsignedByte.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("UnsignedByte.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.UnsignedByte
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalUnsignedByte() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalUnsignedByte() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestShort(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Short
		want []byte
	}{
		{"zero", 0, []byte{0x00, 0x00}},
		{"positive", 32767, []byte{0x7F, 0xFF}},
		{"negative", -32768, []byte{0x80, 0x00}},
		{"minus one", -1, []byte{0xFF, 0xFF}},
		{"256", 256, []byte{0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Short.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Short.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.Short
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalShort() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalShort() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestUnsignedShort(t *testing.T) {
	tests := []struct {
		name string
		val  ns.UnsignedShort
		want []byte
	}{
		{"zero", 0, []byte{0x00, 0x00}},
		{"max", 65535, []byte{0xFF, 0xFF}},
		{"middle", 32768, []byte{0x80, 0x00}},
		{"256", 256, []byte{0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("UnsignedShort.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("UnsignedShort.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.UnsignedShort
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalUnsignedShort() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalUnsignedShort() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Int
		want []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"positive", 2147483647, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"negative", -2147483648, []byte{0x80, 0x00, 0x00, 0x00}},
		{"minus one", -1, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"256", 256, []byte{0x00, 0x00, 0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Int.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Int.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.Int
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalInt() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalInt() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestLong(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Long
		want []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"positive", 9223372036854775807, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"negative", -9223372036854775808, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"minus one", -1, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Long.Marshal() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Long.Marshal() = %v, want %v", got, tt.want)
			}

			var unmarshaled ns.Long
			_, err = unmarshaled.FromBytes(got)
			if err != nil {
				t.Errorf("UnmarshalLong() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalLong() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Float
	}{
		{"zero", 0.0},
		{"one", 1.0},
		{"negative", -1.0},
		{"pi", 3.14159},
		{"large", 1e10},
		{"small", 1e-10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Float.Marshal() error = %v", err)
			}
			var unmarshaled ns.Float
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalFloat() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalFloat() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestDouble(t *testing.T) {
	tests := []struct {
		name string
		val  ns.Double
	}{
		{"zero", 0.0},
		{"one", 1.0},
		{"negative", -1.0},
		{"pi", 3.141592653589793},
		{"large", 1e100},
		{"small", 1e-100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.val.ToBytes()
			if err != nil {
				t.Errorf("Double.Marshal() error = %v", err)
			}
			var unmarshaled ns.Double
			_, err = unmarshaled.FromBytes(marshaled)
			if err != nil {
				t.Errorf("UnmarshalDouble() error = %v", err)
			}
			if unmarshaled != tt.val {
				t.Errorf("UnmarshalDouble() = %v, want %v", unmarshaled, tt.val)
			}
		})
	}
}

func TestPrimitivesErrorCases(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		// Test Boolean with insufficient data
		var temp ns.Boolean
		_, err := temp.FromBytes(ns.ByteArray{})
		if err == nil {
			t.Error("Boolean.Unmarshal() should error on empty data")
		}

		// Test Short with insufficient data
		var temp2 ns.Short
		_, err = temp2.FromBytes(ns.ByteArray{0x01})
		if err == nil {
			t.Error("Short.Unmarshal() should error on insufficient data")
		}

		// Test Int with insufficient data
		var temp3 ns.Int
		_, err = temp3.FromBytes(ns.ByteArray{0x01, 0x02, 0x03})
		if err == nil {
			t.Error("Int.Unmarshal() should error on insufficient data")
		}

		// Test Long with insufficient data
		var temp4 ns.Long
		_, err = temp4.FromBytes(ns.ByteArray{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07})
		if err == nil {
			t.Error("Long.Unmarshal() should error on insufficient data")
		}
	})
}

func TestBooleanInterface(t *testing.T) {
	val := ns.Boolean(true)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Boolean.Marshal() error = %v", err)
	}

	var result ns.Boolean
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Boolean.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Boolean interface roundtrip: got %v, want %v", result, val)
	}
}

func TestIntInterface(t *testing.T) {
	val := ns.Int(42)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Int.Marshal() error = %v", err)
	}

	var result ns.Int
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Int.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Int interface roundtrip: got %v, want %v", result, val)
	}
}

func TestFloatInterface(t *testing.T) {
	val := ns.Float(3.14159)
	data, err := val.ToBytes()
	if err != nil {
		t.Errorf("Float.Marshal() error = %v", err)
	}

	var result ns.Float
	_, err = result.FromBytes(data)
	if err != nil {
		t.Errorf("Float.Unmarshal() error = %v", err)
	}
	if result != val {
		t.Errorf("Float interface roundtrip: got %v, want %v", result, val)
	}
}

func TestGenericMarshalUnmarshal(t *testing.T) {
	t.Run("Boolean", func(t *testing.T) {
		val := ns.Boolean(true)
		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Marshal(Boolean) error = %v", err)
		}

		var result ns.Boolean
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Unmarshal(Boolean) error = %v", err)
		}
		if result != val {
			t.Errorf("Generic Boolean roundtrip: got %v, want %v", result, val)
		}
	})

	t.Run("Int", func(t *testing.T) {
		val := ns.Int(42)
		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Marshal(Int) error = %v", err)
		}

		var result ns.Int
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Unmarshal(Int) error = %v", err)
		}
		if result != val {
			t.Errorf("Generic Int roundtrip: got %v, want %v", result, val)
		}
	})

	t.Run("Float", func(t *testing.T) {
		val := ns.Float(3.14159)
		data, err := val.ToBytes()
		if err != nil {
			t.Errorf("Marshal(Float) error = %v", err)
		}

		var result ns.Float
		_, err = result.FromBytes(data)
		if err != nil {
			t.Errorf("Unmarshal(Float) error = %v", err)
		}
		if result != val {
			t.Errorf("Generic Float roundtrip: got %v, want %v", result, val)
		}
	})
}
