package wire

import (
	"fmt"
)

// Slot represents an item stack with data components.
// Components are stored as raw bytes to keep this package protocol-level only.
// Callers should use a higher-level package to parse specific component types.
//
// Wire format:
//
//	┌──────────────────┬─────────────────┬─────────────────┬─────────────────┬──────────────────────────────────┐
//	│  Count (VarInt)  │  ItemID (VarInt)│  Add (VarInt)   │  Remove (VarInt)│  Components...                   │
//	└──────────────────┴─────────────────┴─────────────────┴─────────────────┴──────────────────────────────────┘
//
// If Count <= 0, the slot is empty and no further data is read.
// ItemID is the registry ID from minecraft:item.
// Add count is the number of components to add (with data).
// Remove count is the number of component type IDs to remove.
type Slot struct {
	Count      VarInt
	ItemID     VarInt         // only if Count > 0
	Components SlotComponents // only if Count > 0
}

// SlotComponents holds the component modifications for a slot.
type SlotComponents struct {
	Add    []RawSlotComponent // components with data (ID + raw bytes)
	Remove []VarInt           // component type IDs to remove
}

// RawSlotComponent stores a component as ID + raw bytes.
// This allows passthrough without parsing component internals.
type RawSlotComponent struct {
	ID   VarInt
	Data []byte
}

// EmptySlot returns an empty slot.
func EmptySlot() Slot {
	return Slot{Count: 0}
}

// NewSlot creates a slot with the given item and count.
func NewSlot(itemID VarInt, count VarInt) Slot {
	return Slot{
		Count:  count,
		ItemID: itemID,
	}
}

// IsEmpty returns true if the slot is empty.
func (s *Slot) IsEmpty() bool {
	return s.Count <= 0
}

// SlotDecoder is a function that decodes a component's data given its ID.
// Returns the raw bytes of the component. If the component format is unknown,
// return an error - there's no way to know where a component ends without
// understanding its format.
type SlotDecoder func(buf *PacketBuffer, componentID VarInt) ([]byte, error)

// Encode writes the slot to the buffer.
func (s *Slot) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(s.Count); err != nil {
		return fmt.Errorf("failed to write slot count: %w", err)
	}

	if s.Count <= 0 {
		return nil
	}

	if err := buf.WriteVarInt(s.ItemID); err != nil {
		return fmt.Errorf("failed to write slot item id: %w", err)
	}

	// write add count
	if err := buf.WriteVarInt(VarInt(len(s.Components.Add))); err != nil {
		return fmt.Errorf("failed to write slot add count: %w", err)
	}

	// write remove count
	if err := buf.WriteVarInt(VarInt(len(s.Components.Remove))); err != nil {
		return fmt.Errorf("failed to write slot remove count: %w", err)
	}

	// write added components
	for i, comp := range s.Components.Add {
		if err := buf.WriteVarInt(comp.ID); err != nil {
			return fmt.Errorf("failed to write component %d id: %w", i, err)
		}
		if _, err := buf.Write(comp.Data); err != nil {
			return fmt.Errorf("failed to write component %d data: %w", i, err)
		}
	}

	// write removed component IDs
	for i, id := range s.Components.Remove {
		if err := buf.WriteVarInt(id); err != nil {
			return fmt.Errorf("failed to write removed component %d id: %w", i, err)
		}
	}

	return nil
}

// ReadSlot reads a slot from the buffer using a decoder that knows component sizes.
// The decoder must return the raw bytes for each component.
func (pb *PacketBuffer) ReadSlot(decode SlotDecoder) (Slot, error) {
	var s Slot
	count, err := pb.ReadVarInt()
	if err != nil {
		return Slot{}, fmt.Errorf("failed to read slot count: %w", err)
	}
	s.Count = count

	if s.Count <= 0 {
		return s, nil
	}

	if s.ItemID, err = pb.ReadVarInt(); err != nil {
		return Slot{}, fmt.Errorf("failed to read slot item id: %w", err)
	}

	addCount, err := pb.readCount(1)
	if err != nil {
		return Slot{}, fmt.Errorf("failed to read slot add count: %w", err)
	}

	removeCount, err := pb.readCount(1)
	if err != nil {
		return Slot{}, fmt.Errorf("failed to read slot remove count: %w", err)
	}

	// read added components
	s.Components.Add = makeCapped[RawSlotComponent](addCount)
	for i := range addCount {
		compID, err := pb.ReadVarInt()
		if err != nil {
			return Slot{}, fmt.Errorf("failed to read component %d id: %w", i, err)
		}

		data, err := decode(pb, compID)
		if err != nil {
			return Slot{}, fmt.Errorf("failed to read component %d (id=%d): %w", i, compID, err)
		}
		s.Components.Add = append(s.Components.Add, RawSlotComponent{ID: compID, Data: data})
	}

	// read removed component IDs
	s.Components.Remove = makeCapped[VarInt](removeCount)
	for i := range removeCount {
		id, err := pb.ReadVarInt()
		if err != nil {
			return Slot{}, fmt.Errorf("failed to read removed component %d id: %w", i, err)
		}
		s.Components.Remove = append(s.Components.Remove, id)
	}

	return s, nil
}

// WriteSlot writes a slot to the buffer.
func (pb *PacketBuffer) WriteSlot(s Slot) error {
	return s.Encode(pb)
}

// AddComponent adds a raw component to the slot.
func (s *Slot) AddComponent(id VarInt, data []byte) {
	s.Components.Add = append(s.Components.Add, RawSlotComponent{ID: id, Data: data})
}

// RemoveComponent marks a component type for removal.
func (s *Slot) RemoveComponent(id VarInt) {
	s.Components.Remove = append(s.Components.Remove, id)
}
