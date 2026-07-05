package net_structures

import (
	"fmt"
)

// IDSet represents a registry ID set, which can be either a tag reference
// or an inline list of IDs.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  Type (VarInt)  │  Data (depends on Type)             │
//	└─────────────────┴─────────────────────────────────────┘
//
// If Type = 0: followed by an Identifier (tag name)
// If Type > 0: followed by (Type - 1) VarInt IDs
type IDSet struct {
	// IsTag indicates whether this is a tag reference (true) or inline IDs (false).
	IsTag bool
	// TagName is the tag identifier (only valid if IsTag is true).
	TagName Identifier
	// IDs is the list of registry IDs (only valid if IsTag is false).
	IDs []VarInt
}

// Decode reads an IDSet from the buffer.
func (s *IDSet) Decode(buf *PacketBuffer) error {
	typeVal, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read id set type: %w", err)
	}

	if typeVal == 0 {
		// tag reference
		s.IsTag = true
		s.TagName, err = buf.ReadIdentifier()
		if err != nil {
			return fmt.Errorf("failed to read id set tag name: %w", err)
		}
	} else {
		// inline IDs
		s.IsTag = false
		count := int(typeVal - 1)
		s.IDs = make([]VarInt, count)
		for i := range s.IDs {
			s.IDs[i], err = buf.ReadVarInt()
			if err != nil {
				return fmt.Errorf("failed to read id set id %d: %w", i, err)
			}
		}
	}
	return nil
}

// Encode writes an IDSet to the buffer.
func (s *IDSet) Encode(buf *PacketBuffer) error {
	if s.IsTag {
		if err := buf.WriteVarInt(0); err != nil {
			return fmt.Errorf("failed to write id set type: %w", err)
		}
		if err := buf.WriteIdentifier(s.TagName); err != nil {
			return fmt.Errorf("failed to write id set tag name: %w", err)
		}
	} else {
		if err := buf.WriteVarInt(VarInt(len(s.IDs) + 1)); err != nil {
			return fmt.Errorf("failed to write id set type: %w", err)
		}
		for i, id := range s.IDs {
			if err := buf.WriteVarInt(id); err != nil {
				return fmt.Errorf("failed to write id set id %d: %w", i, err)
			}
		}
	}
	return nil
}

// XOrY represents a boolean-selected variant between two types.
// When IsX is true, the X value is encoded; otherwise, Y is encoded.
//
// Wire format:
//
//	┌──────────────────┬─────────────────────────────────────┐
//	│  IsX (Boolean)   │  X or Y (depending on IsX)          │
//	└──────────────────┴─────────────────────────────────────┘
//
// This pattern is relatively rare and is typically handled inline in packet
// definitions since both X and Y types must be known at compile time.
type XOrY[X, Y any] struct {
	IsX bool
	X   X
	Y   Y
}

// DecodeWith reads an XOrY using the provided decoders.
func (v *XOrY[X, Y]) DecodeWith(buf *PacketBuffer, decodeX ElementDecoder[X], decodeY ElementDecoder[Y]) error {
	isX, err := buf.ReadBool()
	if err != nil {
		return fmt.Errorf("failed to read x-or-y selector: %w", err)
	}
	v.IsX = bool(isX)

	if v.IsX {
		v.X, err = decodeX(buf)
		if err != nil {
			return fmt.Errorf("failed to read x-or-y x value: %w", err)
		}
	} else {
		v.Y, err = decodeY(buf)
		if err != nil {
			return fmt.Errorf("failed to read x-or-y y value: %w", err)
		}
	}
	return nil
}

// EncodeWith writes an XOrY using the provided encoders.
func (v XOrY[X, Y]) EncodeWith(buf *PacketBuffer, encodeX ElementEncoder[X], encodeY ElementEncoder[Y]) error {
	if err := buf.WriteBool(Boolean(v.IsX)); err != nil {
		return fmt.Errorf("failed to write x-or-y selector: %w", err)
	}
	if v.IsX {
		if err := encodeX(buf, v.X); err != nil {
			return fmt.Errorf("failed to write x-or-y x value: %w", err)
		}
	} else {
		if err := encodeY(buf, v.Y); err != nil {
			return fmt.Errorf("failed to write x-or-y y value: %w", err)
		}
	}
	return nil
}

// Get returns either the X value (if IsX) or Y value, along with the selector.
func (v XOrY[X, Y]) Get() (x X, y Y, isX bool) {
	return v.X, v.Y, v.IsX
}

// IDOrX represents a registry ID or an inline value.
// Used when a field can reference a registry entry by ID or define a value inline.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  ID (VarInt)    │  Value (X, only if ID = 0)          │
//	└─────────────────┴─────────────────────────────────────┘
//
// If ID = 0, the inline value follows.
// If ID > 0, it represents registry ID + 1 (actual ID is ID - 1).
type IDOrX[T any] struct {
	// IsInline indicates whether this contains an inline value (true) or a registry ID (false).
	IsInline bool
	// ID is the registry ID (only valid if IsInline is false).
	// Note: wire format uses ID+1, but this field stores the actual ID.
	ID VarInt
	// Value is the inline value (only valid if IsInline is true).
	Value T
}

// DecodeWith reads an IDOrX using the provided decoder for inline values.
func (x *IDOrX[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	id, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read id-or-x id: %w", err)
	}

	if id == 0 {
		x.IsInline = true
		x.Value, err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read id-or-x inline value: %w", err)
		}
	} else {
		x.IsInline = false
		x.ID = id - 1 // wire format is ID+1
	}
	return nil
}

// EncodeWith writes an IDOrX using the provided encoder for inline values.
func (x IDOrX[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if x.IsInline {
		if err := buf.WriteVarInt(0); err != nil {
			return fmt.Errorf("failed to write id-or-x id: %w", err)
		}
		if err := encode(buf, x.Value); err != nil {
			return fmt.Errorf("failed to write id-or-x inline value: %w", err)
		}
	} else {
		if err := buf.WriteVarInt(x.ID + 1); err != nil {
			return fmt.Errorf("failed to write id-or-x id: %w", err)
		}
	}
	return nil
}

// Get returns the ID (if reference) or -1 (if inline), and the inline value if present.
func (x IDOrX[T]) Get() (id VarInt, value T, isInline bool) {
	if x.IsInline {
		return -1, x.Value, true
	}
	return x.ID, x.Value, false
}
