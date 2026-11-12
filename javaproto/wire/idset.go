package wire

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
		return nil
	}

	// inline IDs: Type is the ID count plus one, and each ID takes at least one byte
	if typeVal < 0 {
		return fmt.Errorf("invalid id set type %d", typeVal)
	}
	count := int(typeVal) - 1
	if rem, ok := remaining(buf.reader); ok && count > rem {
		return fmt.Errorf("id set count %d exceeds remaining %d bytes", count, rem)
	}
	s.IsTag = false
	s.IDs = makeCapped[VarInt](count)
	for i := range count {
		id, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read id set id %d: %w", i, err)
		}
		s.IDs = append(s.IDs, id)
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
