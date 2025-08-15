package net_structures

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Angle - rotation angle in steps of 1/256 of a full turn
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Angle
type Angle byte

// NewAngle creates an Angle from a yaw angle in degrees
func NewAngle(yaw float64) Angle {
	return Angle(yaw * 256 / 360)
}

func (a Angle) ToBytes() (ByteArray, error) {
	return ByteArray{byte(a)}, nil
}

func (a *Angle) FromBytes(data ByteArray) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for Angle")
	}
	*a = Angle(data[0])
	return 1, nil
}

func (a Angle) ToYaw() float64 {
	return float64(a) * 360 / 256
}

// UUID - 128-bit universally unique identifier
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:UUID
type UUID [16]byte

// NewUUID creates a UUID from a string (with or without dashes)
func NewUUID(s string) (UUID, error) {
	var u UUID
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return u, fmt.Errorf("invalid UUID length: expected 32 hex characters, got %d", len(s))
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return u, fmt.Errorf("invalid UUID format: %w", err)
	}

	copy(u[:], bytes)
	return u, nil
}

func (u UUID) ToBytes() (ByteArray, error) {
	return ByteArray(u[:]), nil
}

func (u *UUID) FromBytes(data ByteArray) (int, error) {
	if len(data) < 16 {
		return 0, errors.New("insufficient data for UUID")
	}
	copy(u[:], data[:16])
	return 16, nil
}

// String returns the UUID as a formatted string (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// StringNoDashes returns the UUID as a hex string without dashes
func (u UUID) StringNoDashes() string {
	return hex.EncodeToString(u[:])
}

// ValidateUUID validates a UUID format string.
// Should be 32 hex characters (no dashes) or 36 characters (with dashes).
func ValidateUUID(uuid string) bool {
	if len(uuid) == 32 {
		// no dashes; validate hex
		for _, r := range uuid {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
		return true
	} else if len(uuid) == 36 {
		// dashes: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
			return false
		}

		// strip dashes and validate hex
		cleaned := uuid[0:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:36]
		return ValidateUUID(cleaned)
	}

	return false
}

// TeleportFlags - bit field for teleportation
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Teleport_Flags
type TeleportFlags uint32

func (f TeleportFlags) ToBytes() (ByteArray, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(f))
	return data, nil
}

func (f *TeleportFlags) FromBytes(data ByteArray) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("insufficient data for TeleportFlags")
	}
	*f = TeleportFlags(binary.BigEndian.Uint32(data))
	return 4, nil
}

// Text Component encoded as NBT
//
// https://minecraft.wiki/w/Text_component_format
type TextComponent struct {
	// TODO: Implement NBT handling (use https://github.com/Tnze/go-mc/nbt)
	Data ByteArray
}

// Entity Metadata - miscellaneous information about an entity
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Entity_Metadata
type EntityMetadata struct {
	// TODO: Implement entity metadata format
	Data ByteArray
}

// Slot - an item stack in an inventory or container
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Slot
type Slot struct {
	// TODO: Implement slot data format
	Data ByteArray
}

// HashedSlot - similar to Slot but with hashed data components
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Hashed_Slot
type HashedSlot struct {
	// TODO: Implement hashed slot format
	Data ByteArray
}

// NBT - Named Binary Tag
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:NBT
type NBT struct {
	// TODO: Implement NBT format
	Data ByteArray
}

// Optional - wrapper for optional fields
type Optional[T any] struct {
	Present bool
	Value   T
}

// PrefixedOptional - optional field prefixed with boolean
type PrefixedOptional[T any] struct {
	Present bool
	Value   T
}

// Array - fixed-size array wrapper
type Array[T any] []T

// PrefixedArray - length-prefixed array
type PrefixedArray[T any] struct {
	Length VarInt
	Data   []T
}

// Enum - represents an enum value
type Enum any

// IDor - either a registry ID or inline data
type IDor[T any] struct {
	IsID bool
	ID   VarInt
	Data T
}

// IDSet - set of registry IDs
type IDSet struct {
	// TODO: Implement ID set format
	Data ByteArray
}

// SoundEvent - parameters for a sound event
type SoundEvent struct {
	// TODO: Implement sound event format
	Data ByteArray
}

// ChatType - parameters for direct chat
type ChatType struct {
	// TODO: Implement chat type format
	Data ByteArray
}

// RecipeDisplay - recipe description for client
type RecipeDisplay struct {
	// TODO: Implement recipe display format
	Data ByteArray
}

// SlotDisplay - recipe ingredient slot description
type SlotDisplay struct {
	// TODO: Implement slot display format
	Data ByteArray
}

// ChunkData - chunk data structure
type ChunkData struct {
	// TODO: Implement chunk data format
	Data ByteArray
}

// LightData - light data structure
type LightData struct {
	// TODO: Implement light data format
	Data ByteArray
}

// Or - represents X or Y type
type Or[X, Y any] struct {
	IsX  bool
	XVal X
	YVal Y
}
