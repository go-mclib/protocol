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
	Data      ByteArray
	parsedNBT *NBT
	text      string
}

func (t TextComponent) ToBytes() (ByteArray, error) {
	return t.Data.ToBytes()
}

func (t *TextComponent) FromBytes(data ByteArray) (int, error) {
	var nbt NBT
	if bytesRead, err := nbt.FromBytes(data); err == nil {
		t.Data = data[:bytesRead]
		t.parsedNBT = &nbt
		t.text = nbt.ExtractTextFromNBT()
		return bytesRead, nil
	}

	var str String
	if strBytes, err := str.FromBytes(data); err == nil && strBytes > 0 {
		t.Data = data[:strBytes]
		if comp, perr := ParseTextComponentFromString(string(str)); perr == nil {
			t.text = comp.String()
		} else {
			t.text = string(str)
		}
		t.parsedNBT = nil
		return strBytes, nil
	}

	return 0, fmt.Errorf("failed to parse TextComponent: neither NBT nor String parsing succeeded")
}

// GetText returns the extracted text from the component
func (t TextComponent) GetText() string {
	return t.text
}

// GetNBT returns the parsed NBT data (may be nil)
func (t TextComponent) GetNBT() *NBT {
	return t.parsedNBT
}

// String returns the text content
func (t TextComponent) String() string {
	if t.text != "" {
		return t.text
	}
	return "<empty text component>"
}

// ToMap converts the TextComponent's NBT data to a map[string]any structure
func (t TextComponent) ToMap() (map[string]any, error) {
	if t.parsedNBT == nil {
		return nil, fmt.Errorf("no parsed NBT data available")
	}
	return t.parsedNBT.ToMap()
}

// FromMap sets the TextComponent from a map[string]any structure
func (t *TextComponent) FromMap(data map[string]any) error {
	if data == nil {
		t.Data = ByteArray{0x00}
		t.parsedNBT = nil
		t.text = ""
		return nil
	}

	nbt := NBT{}
	err := nbt.FromMap(data)
	if err != nil {
		return fmt.Errorf("failed to create NBT from map: %w", err)
	}

	nbtBytes, err := nbt.ToBytes()
	if err != nil {
		return fmt.Errorf("failed to encode NBT: %w", err)
	}

	t.Data = nbtBytes
	t.parsedNBT = &nbt
	t.text = nbt.ExtractTextFromNBT()

	return nil
}

// ToMapSafe converts the TextComponent to a map, returning an empty map if conversion fails
func (t TextComponent) ToMapSafe() map[string]any {
	result, err := t.ToMap()
	if err != nil {
		return make(map[string]any)
	}
	return result
}

// NewTextComponentFromMap creates a new TextComponent from a map
func NewTextComponentFromMap(data map[string]any) (TextComponent, error) {
	var tc TextComponent
	err := tc.FromMap(data)
	return tc, err
}

// NewTextComponentFromString creates a TextComponent from a simple text string
func NewTextComponentFromString(text string) TextComponent {
	data := map[string]any{"text": text}
	tc, _ := NewTextComponentFromMap(data)
	return tc
}

// Entity Metadata - miscellaneous information about an entity
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Entity_Metadata
type EntityMetadata struct {
	Data ByteArray
}

func (e EntityMetadata) ToBytes() (ByteArray, error) {
	return e.Data.ToBytes()
}

func (e *EntityMetadata) FromBytes(data ByteArray) (int, error) {
	return e.Data.FromBytes(data)
}

// Slot - an item stack in an inventory or container
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Slot
type Slot struct {
	Data ByteArray
}

func (s Slot) ToBytes() (ByteArray, error) {
	return s.Data.ToBytes()
}

func (s *Slot) FromBytes(data ByteArray) (int, error) {
	return s.Data.FromBytes(data)
}

// HashedSlot - similar to Slot but with hashed data components
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Hashed_Slot
type HashedSlot struct {
	Data ByteArray
}

func (h HashedSlot) ToBytes() (ByteArray, error) {
	return h.Data.ToBytes()
}

func (h *HashedSlot) FromBytes(data ByteArray) (int, error) {
	return h.Data.FromBytes(data)
}

// FixedByteArray represents a ByteArray with a fixed length
type FixedByteArray struct {
	Data   ByteArray
	Length int // Expected fixed length
}

func NewFixedByteArray(length int) FixedByteArray {
	return FixedByteArray{
		Data:   make(ByteArray, length),
		Length: length,
	}
}

func (f FixedByteArray) ToBytes() (ByteArray, error) {
	if len(f.Data) != f.Length {
		return nil, fmt.Errorf("FixedByteArray length mismatch: expected %d, got %d", f.Length, len(f.Data))
	}
	return f.Data, nil
}

func (f *FixedByteArray) FromBytes(data ByteArray) (int, error) {
	if len(data) < f.Length {
		return 0, fmt.Errorf("insufficient data for FixedByteArray: need %d bytes, have %d", f.Length, len(data))
	}
	f.Data = make(ByteArray, f.Length)
	copy(f.Data, data[:f.Length])
	return f.Length, nil
}

// Optional - wrapper for optional fields without a boolean prefix
// Presence must be known from context
type Optional[T any] struct {
	Present bool
	Value   T
}

func (o Optional[T]) ToBytes() (ByteArray, error) {
	if !o.Present {
		return ByteArray{}, nil
	}

	// Direct handling for ByteArray - just pass through
	if ba, ok := any(o.Value).(ByteArray); ok {
		return ba, nil
	}

	// Handle FixedByteArray
	if fba, ok := any(o.Value).(FixedByteArray); ok {
		return fba.ToBytes()
	}

	// Use marshaler interface
	if marshaler, ok := any(o.Value).(interface{ ToBytes() (ByteArray, error) }); ok {
		return marshaler.ToBytes()
	}

	return nil, fmt.Errorf("type %T does not implement ToBytes method", o.Value)
}

func (o *Optional[T]) FromBytes(data ByteArray) (int, error) {
	// For Optional without prefix, presence must be set externally
	if !o.Present {
		return 0, nil
	}

	// Handle FixedByteArray
	if fba, ok := any(&o.Value).(*FixedByteArray); ok && fba != nil {
		return fba.FromBytes(data)
	}

	// Direct handling for ByteArray - consume all remaining data
	// (unless it's a known fixed size, which should use FixedByteArray)
	if _, isBA := any(o.Value).(ByteArray); isBA {
		ba := make(ByteArray, len(data))
		copy(ba, data)
		if v, ok := any(ba).(T); ok {
			o.Value = v
		}
		return len(data), nil
	}

	// Use unmarshaler interface
	if unmarshaler, ok := any(&o.Value).(interface{ FromBytes(ByteArray) (int, error) }); ok {
		return unmarshaler.FromBytes(data)
	}

	return 0, fmt.Errorf("type %T does not implement FromBytes method", o.Value)
}

// PrefixedOptional - optional field prefixed with boolean
type PrefixedOptional[T any] struct {
	Present bool
	Value   T
}

func (p PrefixedOptional[T]) ToBytes() (ByteArray, error) {
	result, err := Boolean(p.Present).ToBytes()
	if err != nil {
		return nil, err
	}

	if !p.Present {
		return result, nil
	}

	// Direct handling for ByteArray
	if ba, ok := any(p.Value).(ByteArray); ok {
		result = append(result, ba...)
		return result, nil
	}

	// Handle FixedByteArray
	if fba, ok := any(p.Value).(FixedByteArray); ok {
		valueBytes, err := fba.ToBytes()
		if err != nil {
			return nil, err
		}
		result = append(result, valueBytes...)
		return result, nil
	}

	// Use marshaler interface
	if marshaler, ok := any(p.Value).(interface{ ToBytes() (ByteArray, error) }); ok {
		valueBytes, err := marshaler.ToBytes()
		if err != nil {
			return nil, err
		}
		result = append(result, valueBytes...)
		return result, nil
	}

	return nil, fmt.Errorf("type %T does not implement ToBytes method", p.Value)
}

func (p *PrefixedOptional[T]) FromBytes(data ByteArray) (int, error) {
	var present Boolean
	bytesRead, err := present.FromBytes(data)
	if err != nil {
		return 0, err
	}

	p.Present = bool(present)
	if !p.Present {
		return bytesRead, nil
	}

	// Handle FixedByteArray
	if fba, ok := any(&p.Value).(*FixedByteArray); ok && fba != nil {
		valueBytes, err := fba.FromBytes(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		return bytesRead + valueBytes, nil
	}

	// HACK: make this respect length defined in slice tags
	/*if ba, isBA := any(&p.Value).(*ByteArray); isBA {
		if len(data[bytesRead:]) >= 256 {
			*ba = data[bytesRead : bytesRead+256]
			return bytesRead + 256, nil
		}
		return 0, errors.New("PrefixedOptional[ByteArray] requires FixedByteArray for known sizes, or 256 bytes for signature data")
	}*/

	if unmarshaler, ok := any(&p.Value).(interface{ FromBytes(ByteArray) (int, error) }); ok {
		valueBytes, err := unmarshaler.FromBytes(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		return bytesRead + valueBytes, nil
	}

	return 0, fmt.Errorf("type %T does not implement FromBytes method", p.Value)
}

// Array - fixed-size array wrapper (no length prefix)
// For use when the array size is known from context
type Array[T any] []T

func (a Array[T]) ToBytes() (ByteArray, error) {
	var result ByteArray

	// Optimize for byte arrays - make them behave like ByteArray
	if _, isByte := any(a).(Array[Byte]); isByte {
		bytes := make([]byte, len(a))
		for i := range a {
			if b, ok := any(a[i]).(Byte); ok {
				bytes[i] = byte(b)
			}
		}
		return ByteArray(bytes), nil
	}

	// General case for other types
	for i, item := range a {
		if marshaler, ok := any(item).(interface{ ToBytes() (ByteArray, error) }); ok {
			itemBytes, err := marshaler.ToBytes()
			if err != nil {
				return nil, fmt.Errorf("error marshaling array item %d: %w", i, err)
			}
			result = append(result, itemBytes...)
		} else {
			return nil, fmt.Errorf("type %T does not implement ToBytes method", item)
		}
	}
	return result, nil
}

// FromBytesWithLength reads a fixed-size array with known length
func (a *Array[T]) FromBytesWithLength(data ByteArray, length int) (int, error) {
	*a = make(Array[T], length)

	// Optimize for byte arrays
	if _, isByte := any(*a).(Array[Byte]); isByte {
		if len(data) < length {
			return 0, errors.New("insufficient data for byte array")
		}
		for i := 0; i < length; i++ {
			if b, ok := any(Byte(data[i])).(T); ok {
				(*a)[i] = b
			}
		}
		return length, nil
	}

	// General case for other types
	offset := 0
	for i := 0; i < length; i++ {
		if unmarshaler, ok := any(&(*a)[i]).(interface{ FromBytes(ByteArray) (int, error) }); ok {
			itemBytes, err := unmarshaler.FromBytes(data[offset:])
			if err != nil {
				return 0, fmt.Errorf("error unmarshaling array item %d: %w", i, err)
			}
			offset += itemBytes
		} else {
			return 0, fmt.Errorf("type %T does not implement FromBytes method", (*a)[i])
		}
	}
	return offset, nil
}

// FromBytes consumes all available data
// Use FromBytesWithLength when you know the array size
func (a *Array[T]) FromBytes(data ByteArray) (int, error) {
	// Special case for byte arrays - consume all
	if _, isByte := any(*a).(Array[Byte]); isByte {
		*a = make(Array[T], len(data))
		for i := 0; i < len(data); i++ {
			if b, ok := any(Byte(data[i])).(T); ok {
				(*a)[i] = b
			}
		}
		return len(data), nil
	}

	// For other types, we can't know the length without context
	return 0, errors.New("Array.FromBytes requires length context for non-byte types, use FromBytesWithLength")
}

// PrefixedArray - length-prefixed array
// Behaves like PrefixedByteArray when T is Byte
type PrefixedArray[T any] []T

func (p PrefixedArray[T]) ToBytes() (ByteArray, error) {
	// Encode length prefix
	length := VarInt(len(p))
	result, err := length.ToBytes()
	if err != nil {
		return nil, err
	}

	// Optimize for byte arrays - make PrefixedArray[Byte] behave like PrefixedByteArray
	if _, isByte := any(p).(PrefixedArray[Byte]); isByte {
		bytes := make([]byte, len(p))
		for i := range p {
			if b, ok := any(p[i]).(Byte); ok {
				bytes[i] = byte(b)
			}
		}
		result = append(result, bytes...)
		return result, nil
	}

	// General case for other types
	for i, item := range p {
		if marshaler, ok := any(item).(interface{ ToBytes() (ByteArray, error) }); ok {
			itemBytes, err := marshaler.ToBytes()
			if err != nil {
				return nil, fmt.Errorf("error marshaling array item %d: %w", i, err)
			}
			result = append(result, itemBytes...)
		} else {
			return nil, fmt.Errorf("type %T does not implement ToBytes method", item)
		}
	}
	return result, nil
}

func (p *PrefixedArray[T]) FromBytes(data ByteArray) (int, error) {
	var length VarInt
	bytesRead, err := length.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if length < 0 {
		return 0, errors.New("negative array length")
	}

	// Optimize for byte arrays - make PrefixedArray[Byte] behave like PrefixedByteArray
	if _, isByte := any(*p).(PrefixedArray[Byte]); isByte {
		if len(data) < bytesRead+int(length) {
			return 0, errors.New("insufficient data for byte array")
		}
		*p = make(PrefixedArray[T], length)
		for i := 0; i < int(length); i++ {
			if b, ok := any(Byte(data[bytesRead+i])).(T); ok {
				(*p)[i] = b
			}
		}
		return bytesRead + int(length), nil
	}

	// General case for other types
	*p = make(PrefixedArray[T], length)
	offset := bytesRead

	for i := 0; i < int(length); i++ {
		if unmarshaler, ok := any(&(*p)[i]).(interface{ FromBytes(ByteArray) (int, error) }); ok {
			itemBytes, err := unmarshaler.FromBytes(data[offset:])
			if err != nil {
				return 0, fmt.Errorf("error unmarshaling array item %d: %w", i, err)
			}
			offset += itemBytes
		} else {
			return 0, fmt.Errorf("type %T does not implement FromBytes method", (*p)[i])
		}
	}

	return offset, nil
}

// Enum - represents an enum value
type Enum any

// IDor - either a registry ID or inline data
type IDor[T any] struct {
	IsID bool
	ID   VarInt
	Data T
}

func (i IDor[T]) ToBytes() (ByteArray, error) {
	var result ByteArray
	var err error

	if i.IsID {
		// ID + 1 for non-zero registry ID
		result, err = VarInt(i.ID + 1).ToBytes()
		if err != nil {
			return nil, err
		}
	} else {
		// 0 for inline data
		result, err = VarInt(0).ToBytes()
		if err != nil {
			return nil, err
		}

		// Use type assertion to check if T implements ToBytes
		if marshaler, ok := any(i.Data).(interface{ ToBytes() (ByteArray, error) }); ok {
			dataBytes, err := marshaler.ToBytes()
			if err != nil {
				return nil, err
			}
			result = append(result, dataBytes...)
		} else {
			return nil, fmt.Errorf("type %T does not implement ToBytes method", i.Data)
		}
	}

	return result, nil
}

func (i *IDor[T]) FromBytes(data ByteArray) (int, error) {
	var id VarInt
	bytesRead, err := id.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if id == 0 {
		// Inline data
		i.IsID = false

		// Use type assertion to check if T implements FromBytes
		if unmarshaler, ok := any(&i.Data).(interface{ FromBytes(ByteArray) (int, error) }); ok {
			dataBytes, err := unmarshaler.FromBytes(data[bytesRead:])
			if err != nil {
				return 0, err
			}
			return bytesRead + dataBytes, nil
		} else {
			return 0, fmt.Errorf("type %T does not implement FromBytes method", i.Data)
		}
	} else {
		// Registry ID
		i.IsID = true
		i.ID = id - 1 // Registry ID + 1 is stored
		return bytesRead, nil
	}
}

// IDSet - set of registry IDs
type IDSet struct {
	Type    VarInt
	TagName *Identifier // Optional identifier, present when Type is 0
	IDs     []VarInt    // Array of registry IDs, present when Type is not 0
}

func (i IDSet) ToBytes() (ByteArray, error) {
	result, err := i.Type.ToBytes()
	if err != nil {
		return nil, err
	}

	if i.Type == 0 {
		// Tag name
		if i.TagName == nil {
			return nil, errors.New("TagName is required when Type is 0")
		}
		nameBytes, err := i.TagName.ToBytes()
		if err != nil {
			return nil, err
		}
		result = append(result, nameBytes...)
	} else {
		// Array of IDs
		for _, id := range i.IDs {
			idBytes, err := id.ToBytes()
			if err != nil {
				return nil, err
			}
			result = append(result, idBytes...)
		}
	}

	return result, nil
}

func (i *IDSet) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := i.Type.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if i.Type == 0 {
		// Tag name
		var tagName Identifier
		i.TagName = &tagName
		nameBytes, err := i.TagName.FromBytes(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		return bytesRead + nameBytes, nil
	} else {
		// Array of IDs
		arraySize := int(i.Type) - 1
		if arraySize < 0 {
			return 0, errors.New("invalid IDSet type")
		}

		i.IDs = make([]VarInt, arraySize)
		offset := bytesRead

		for j := range arraySize {
			idBytes, err := i.IDs[j].FromBytes(data[offset:])
			if err != nil {
				return 0, err
			}
			offset += idBytes
		}

		return offset, nil
	}
}

// SoundEvent - parameters for a sound event
type SoundEvent struct {
	SoundName     Identifier
	HasFixedRange Boolean
	FixedRange    Optional[Float]
}

func (s SoundEvent) ToBytes() (ByteArray, error) {
	result, err := s.SoundName.ToBytes()
	if err != nil {
		return nil, err
	}

	fixedRangeBytes, err := s.HasFixedRange.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, fixedRangeBytes...)

	if bool(s.HasFixedRange) {
		s.FixedRange.Present = true
		rangeBytes, err := s.FixedRange.ToBytes()
		if err != nil {
			return nil, err
		}
		result = append(result, rangeBytes...)
	}

	return result, nil
}

func (s *SoundEvent) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := s.SoundName.FromBytes(data)
	if err != nil {
		return 0, err
	}

	fixedRangeBytes, err := s.HasFixedRange.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += fixedRangeBytes

	if bool(s.HasFixedRange) {
		s.FixedRange.Present = true
		rangeBytes, err := s.FixedRange.FromBytes(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += rangeBytes
	}

	return bytesRead, nil
}

// ChatType - parameters for direct chat
type ChatType struct {
	Chat      ChatDecoration
	Narration ChatDecoration
}

type ChatDecoration struct {
	TranslationKey String
	Parameters     PrefixedArray[VarInt] // 0: sender, 1: target, 2: content
	Style          NBT
}

func (c ChatType) ToBytes() (ByteArray, error) {
	result, err := c.Chat.ToBytes()
	if err != nil {
		return nil, err
	}

	narrationBytes, err := c.Narration.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, narrationBytes...)

	return result, nil
}

func (c *ChatType) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := c.Chat.FromBytes(data)
	if err != nil {
		return 0, err
	}

	narrationBytes, err := c.Narration.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}

	return bytesRead + narrationBytes, nil
}

func (c ChatDecoration) ToBytes() (ByteArray, error) {
	result, err := c.TranslationKey.ToBytes()
	if err != nil {
		return nil, err
	}

	paramsBytes, err := c.Parameters.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, paramsBytes...)

	styleBytes, err := c.Style.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, styleBytes...)

	return result, nil
}

func (c *ChatDecoration) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := c.TranslationKey.FromBytes(data)
	if err != nil {
		return 0, err
	}

	paramsBytes, err := c.Parameters.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += paramsBytes

	styleBytes, err := c.Style.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += styleBytes

	return bytesRead, nil
}

// RecipeDisplay - recipe description for client
type RecipeDisplay struct {
	Data ByteArray
}

func (r RecipeDisplay) ToBytes() (ByteArray, error) {
	return r.Data.ToBytes()
}

func (r *RecipeDisplay) FromBytes(data ByteArray) (int, error) {
	return r.Data.FromBytes(data)
}

// SlotDisplay - recipe ingredient slot description
type SlotDisplay struct {
	Data ByteArray
}

func (s SlotDisplay) ToBytes() (ByteArray, error) {
	return s.Data.ToBytes()
}

func (s *SlotDisplay) FromBytes(data ByteArray) (int, error) {
	return s.Data.FromBytes(data)
}

// ChunkData - chunk data structure
type ChunkData struct {
	Heightmaps    PrefixedArray[ByteArray] // Heightmap data
	Data          PrefixedArray[Byte]      // Chunk section data
	BlockEntities PrefixedArray[BlockEntity]
}

type BlockEntity struct {
	PackedXZ UnsignedByte
	Y        Short
	Type     VarInt
	Data     NBT
}

func (c ChunkData) ToBytes() (ByteArray, error) {
	result, err := c.Heightmaps.ToBytes()
	if err != nil {
		return nil, err
	}

	dataBytes, err := c.Data.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, dataBytes...)

	blockEntityBytes, err := c.BlockEntities.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, blockEntityBytes...)

	return result, nil
}

func (c *ChunkData) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := c.Heightmaps.FromBytes(data)
	if err != nil {
		return 0, err
	}

	dataBytes, err := c.Data.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += dataBytes

	blockEntityBytes, err := c.BlockEntities.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += blockEntityBytes

	return bytesRead, nil
}

func (b BlockEntity) ToBytes() (ByteArray, error) {
	result, err := b.PackedXZ.ToBytes()
	if err != nil {
		return nil, err
	}

	yBytes, err := b.Y.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, yBytes...)

	typeBytes, err := b.Type.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, typeBytes...)

	dataBytes, err := b.Data.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, dataBytes...)

	return result, nil
}

func (b *BlockEntity) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := b.PackedXZ.FromBytes(data)
	if err != nil {
		return 0, err
	}

	yBytes, err := b.Y.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += yBytes

	typeBytes, err := b.Type.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += typeBytes

	dataBytes, err := b.Data.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += dataBytes

	return bytesRead, nil
}

// LightData - light data structure
type LightData struct {
	SkyLightMask        BitSet
	BlockLightMask      BitSet
	EmptySkyLightMask   BitSet
	EmptyBlockLightMask BitSet
	SkyLightArrays      PrefixedArray[PrefixedArray[Byte]]
	BlockLightArrays    PrefixedArray[PrefixedArray[Byte]]
}

func (l LightData) ToBytes() (ByteArray, error) {
	result, err := l.SkyLightMask.ToBytes()
	if err != nil {
		return nil, err
	}

	blockMaskBytes, err := l.BlockLightMask.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, blockMaskBytes...)

	emptySkyBytes, err := l.EmptySkyLightMask.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, emptySkyBytes...)

	emptyBlockBytes, err := l.EmptyBlockLightMask.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, emptyBlockBytes...)

	skyArrayBytes, err := l.SkyLightArrays.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, skyArrayBytes...)

	blockArrayBytes, err := l.BlockLightArrays.ToBytes()
	if err != nil {
		return nil, err
	}
	result = append(result, blockArrayBytes...)

	return result, nil
}

func (l *LightData) FromBytes(data ByteArray) (int, error) {
	bytesRead, err := l.SkyLightMask.FromBytes(data)
	if err != nil {
		return 0, err
	}

	blockMaskBytes, err := l.BlockLightMask.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += blockMaskBytes

	emptySkyBytes, err := l.EmptySkyLightMask.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += emptySkyBytes

	emptyBlockBytes, err := l.EmptyBlockLightMask.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += emptyBlockBytes

	skyArrayBytes, err := l.SkyLightArrays.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += skyArrayBytes

	blockArrayBytes, err := l.BlockLightArrays.FromBytes(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += blockArrayBytes

	return bytesRead, nil
}

// Or - represents X or Y type
type Or[X, Y any] struct {
	IsX  bool
	XVal X
	YVal Y
}

func (o Or[X, Y]) ToBytes() (ByteArray, error) {
	result, err := Boolean(o.IsX).ToBytes()
	if err != nil {
		return nil, err
	}

	if o.IsX {
		// Use type assertion to check if X implements ToBytes
		if marshaler, ok := any(o.XVal).(interface{ ToBytes() (ByteArray, error) }); ok {
			valueBytes, err := marshaler.ToBytes()
			if err != nil {
				return nil, err
			}
			result = append(result, valueBytes...)
		} else {
			return nil, fmt.Errorf("type %T does not implement ToBytes method", o.XVal)
		}
	} else {
		// Use type assertion to check if Y implements ToBytes
		if marshaler, ok := any(o.YVal).(interface{ ToBytes() (ByteArray, error) }); ok {
			valueBytes, err := marshaler.ToBytes()
			if err != nil {
				return nil, err
			}
			result = append(result, valueBytes...)
		} else {
			return nil, fmt.Errorf("type %T does not implement ToBytes method", o.YVal)
		}
	}

	return result, nil
}

func (o *Or[X, Y]) FromBytes(data ByteArray) (int, error) {
	var isX Boolean
	bytesRead, err := isX.FromBytes(data)
	if err != nil {
		return 0, err
	}

	o.IsX = bool(isX)

	if o.IsX {
		// Use type assertion to check if X implements FromBytes
		if unmarshaler, ok := any(&o.XVal).(interface{ FromBytes(ByteArray) (int, error) }); ok {
			valueBytes, err := unmarshaler.FromBytes(data[bytesRead:])
			if err != nil {
				return 0, err
			}
			return bytesRead + valueBytes, nil
		} else {
			return 0, fmt.Errorf("type %T does not implement FromBytes method", o.XVal)
		}
	} else {
		// Use type assertion to check if Y implements FromBytes
		if unmarshaler, ok := any(&o.YVal).(interface{ FromBytes(ByteArray) (int, error) }); ok {
			valueBytes, err := unmarshaler.FromBytes(data[bytesRead:])
			if err != nil {
				return 0, err
			}
			return bytesRead + valueBytes, nil
		} else {
			return 0, fmt.Errorf("type %T does not implement FromBytes method", o.YVal)
		}
	}
}
