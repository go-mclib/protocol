package wire

// Primitive type definitions for the Minecraft protocol.
// All multi-byte integers are big-endian unless otherwise specified.

// Boolean is a single byte (0x00 = false, 0x01 = true).
type Boolean bool

// Int8 is a signed 8-bit integer (-128 to 127).
type Int8 int8

// Uint8 is an unsigned 8-bit integer (0 to 255).
type Uint8 uint8

// Int16 is a big-endian signed 16-bit integer.
type Int16 int16

// Uint16 is a big-endian unsigned 16-bit integer.
type Uint16 uint16

// Int32 is a big-endian signed 32-bit integer.
type Int32 int32

// Int64 is a big-endian signed 64-bit integer.
type Int64 int64

// Float32 is a big-endian IEEE 754 single-precision float.
type Float32 float32

// Float64 is a big-endian IEEE 754 double-precision float.
type Float64 float64
