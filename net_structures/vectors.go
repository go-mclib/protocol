package net_structures

import (
	"encoding/binary"
	"errors"
	"math"
)

// ChunkPos represents chunk coordinates (X, Z).
// Encoded as a single Long where X is in the upper 32 bits and Z is in the lower 32 bits.
//
// https://wiki.vg/Protocol#ChunkPos
type ChunkPos struct {
	X int32
	Z int32
}

func NewChunkPos(x, z int32) ChunkPos {
	return ChunkPos{X: x, Z: z}
}

func (c ChunkPos) ToBytes() (ByteArray, error) {
	// Pack: X in upper 32 bits, Z in lower 32 bits
	packed := (int64(c.X) << 32) | (int64(c.Z) & 0xFFFFFFFF)
	return Long(packed).ToBytes()
}

func (c *ChunkPos) FromBytes(data ByteArray) (int, error) {
	var packed Long
	bytesRead, err := packed.FromBytes(data)
	if err != nil {
		return 0, err
	}

	// Unpack: X from upper 32 bits, Z from lower 32 bits
	c.X = int32(int64(packed) >> 32)
	c.Z = int32(int64(packed))
	return bytesRead, nil
}

// Vec3 represents a 3D vector with double precision.
// Used for entity positions, velocities, etc.
//
// https://wiki.vg/Protocol#Vec3
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

func (v Vec3) ToBytes() (ByteArray, error) {
	data := make([]byte, 24) // 3 * 8 bytes
	binary.BigEndian.PutUint64(data[0:8], math.Float64bits(v.X))
	binary.BigEndian.PutUint64(data[8:16], math.Float64bits(v.Y))
	binary.BigEndian.PutUint64(data[16:24], math.Float64bits(v.Z))
	return data, nil
}

func (v *Vec3) FromBytes(data ByteArray) (int, error) {
	if len(data) < 24 {
		return 0, errors.New("insufficient data for Vec3")
	}
	v.X = math.Float64frombits(binary.BigEndian.Uint64(data[0:8]))
	v.Y = math.Float64frombits(binary.BigEndian.Uint64(data[8:16]))
	v.Z = math.Float64frombits(binary.BigEndian.Uint64(data[16:24]))
	return 24, nil
}

// Vector3f represents a 3D vector with single precision floats.
// Used for particle positions, rotations, etc.
//
// https://wiki.vg/Protocol#Vector3f
type Vector3f struct {
	X float32
	Y float32
	Z float32
}

func NewVector3f(x, y, z float32) Vector3f {
	return Vector3f{X: x, Y: y, Z: z}
}

func (v Vector3f) ToBytes() (ByteArray, error) {
	data := make([]byte, 12) // 3 * 4 bytes
	binary.BigEndian.PutUint32(data[0:4], math.Float32bits(v.X))
	binary.BigEndian.PutUint32(data[4:8], math.Float32bits(v.Y))
	binary.BigEndian.PutUint32(data[8:12], math.Float32bits(v.Z))
	return data, nil
}

func (v *Vector3f) FromBytes(data ByteArray) (int, error) {
	if len(data) < 12 {
		return 0, errors.New("insufficient data for Vector3f")
	}
	v.X = math.Float32frombits(binary.BigEndian.Uint32(data[0:4]))
	v.Y = math.Float32frombits(binary.BigEndian.Uint32(data[4:8]))
	v.Z = math.Float32frombits(binary.BigEndian.Uint32(data[8:12]))
	return 12, nil
}

// Quaternionf represents a rotation quaternion with 4 floats (x, y, z, w).
// Used for entity head rotations and other orientation data.
//
// https://wiki.vg/Protocol#Quaternionf
type Quaternionf struct {
	X float32
	Y float32
	Z float32
	W float32
}

func NewQuaternionf(x, y, z, w float32) Quaternionf {
	return Quaternionf{X: x, Y: y, Z: z, W: w}
}

func (q Quaternionf) ToBytes() (ByteArray, error) {
	data := make([]byte, 16) // 4 * 4 bytes
	binary.BigEndian.PutUint32(data[0:4], math.Float32bits(q.X))
	binary.BigEndian.PutUint32(data[4:8], math.Float32bits(q.Y))
	binary.BigEndian.PutUint32(data[8:12], math.Float32bits(q.Z))
	binary.BigEndian.PutUint32(data[12:16], math.Float32bits(q.W))
	return data, nil
}

func (q *Quaternionf) FromBytes(data ByteArray) (int, error) {
	if len(data) < 16 {
		return 0, errors.New("insufficient data for Quaternionf")
	}
	q.X = math.Float32frombits(binary.BigEndian.Uint32(data[0:4]))
	q.Y = math.Float32frombits(binary.BigEndian.Uint32(data[4:8]))
	q.Z = math.Float32frombits(binary.BigEndian.Uint32(data[8:12]))
	q.W = math.Float32frombits(binary.BigEndian.Uint32(data[12:16]))
	return 16, nil
}

// GlobalPos represents a position in a specific dimension.
// Encoded as Identifier (dimension) + Position (block coordinates).
//
// https://wiki.vg/Protocol#GlobalPos
type GlobalPos struct {
	Dimension Identifier
	Pos       Position
}

func NewGlobalPos(dimension Identifier, pos Position) GlobalPos {
	return GlobalPos{Dimension: dimension, Pos: pos}
}

func (g GlobalPos) ToBytes() (ByteArray, error) {
	result, err := g.Dimension.ToBytes()
	if err != nil {
		return nil, err
	}

	posBytes, err := g.Pos.ToBytes()
	if err != nil {
		return nil, err
	}

	result = append(result, posBytes...)
	return result, nil
}

func (g *GlobalPos) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := g.Dimension.FromBytes(data)
	if err != nil {
		return 0, err
	}

	posBytes, err := g.Pos.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}

	return bytesRead + posBytes, nil
}

// LpVec3 represents a length-prefixed, quantized Vec3.
// This is a highly compressed format used for entity movement/velocity.
// The encoding uses bit-packing and quantization to minimize bandwidth.
//
// Note: This is complex and may be implemented as ByteArray initially if not needed.
//
// https://wiki.vg/Protocol#LpVec3
type LpVec3 struct {
	Data ByteArray
}

func (l LpVec3) ToBytes() (ByteArray, error) {
	return l.Data.ToBytes()
}

func (l *LpVec3) FromBytes(data ByteArray) (int, error) {
	// TODO: Implement proper LpVec3 decoding
	// For now, treat as opaque ByteArray
	// The actual implementation requires complex bit manipulation
	// See: net/minecraft/network/LpVec3.java
	return l.Data.FromBytes(data)
}
