package net_structures

import (
	"fmt"
)

// Slot represents an item stack with data components.
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
	Add    []SlotComponent // components with data
	Remove []VarInt        // component type IDs to remove
}

// SlotComponent is the interface for item data components.
type SlotComponent interface {
	// ComponentID returns the component type ID.
	ComponentID() VarInt
	// Encode writes the component data to the buffer.
	Encode(buf *PacketBuffer) error
	// Decode reads the component data from the buffer.
	Decode(buf *PacketBuffer) error
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
		if err := buf.WriteVarInt(comp.ComponentID()); err != nil {
			return fmt.Errorf("failed to write component %d id: %w", i, err)
		}
		if err := comp.Encode(buf); err != nil {
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

// Decode reads a slot from the buffer.
func (s *Slot) Decode(buf *PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot count: %w", err)
	}
	s.Count = count

	if s.Count <= 0 {
		return nil
	}

	s.ItemID, err = buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot item id: %w", err)
	}

	addCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot add count: %w", err)
	}

	removeCount, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read slot remove count: %w", err)
	}

	// read added components
	s.Components.Add = make([]SlotComponent, addCount)
	for i := range s.Components.Add {
		compID, err := buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read component %d id: %w", i, err)
		}

		comp := NewSlotComponent(compID)
		if err := comp.Decode(buf); err != nil {
			return fmt.Errorf("failed to read component %d (id=%d): %w", i, compID, err)
		}
		s.Components.Add[i] = comp
	}

	// read removed component IDs
	s.Components.Remove = make([]VarInt, removeCount)
	for i := range s.Components.Remove {
		s.Components.Remove[i], err = buf.ReadVarInt()
		if err != nil {
			return fmt.Errorf("failed to read removed component %d id: %w", i, err)
		}
	}

	return nil
}

// ReadSlot reads a slot from the buffer.
func (pb *PacketBuffer) ReadSlot() (Slot, error) {
	var slot Slot
	err := slot.Decode(pb)
	return slot, err
}

// WriteSlot writes a slot to the buffer.
func (pb *PacketBuffer) WriteSlot(s Slot) error {
	return s.Encode(pb)
}

// GetComponent returns the first component with the given ID, or nil if not found.
func (s *Slot) GetComponent(id VarInt) SlotComponent {
	for _, comp := range s.Components.Add {
		if comp.ComponentID() == id {
			return comp
		}
	}
	return nil
}

// AddComponent adds a component to the slot.
func (s *Slot) AddComponent(comp SlotComponent) {
	s.Components.Add = append(s.Components.Add, comp)
}

// RemoveComponent marks a component type for removal.
func (s *Slot) RemoveComponent(id VarInt) {
	s.Components.Remove = append(s.Components.Remove, id)
}
