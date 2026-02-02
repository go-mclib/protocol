package net_structures

import (
	"bytes"
	"testing"
)

// testSlotDecoder is a simple decoder for testing that handles a few known component types.
func testSlotDecoder(buf *PacketBuffer, id VarInt) ([]byte, error) {
	w := NewWriter()
	switch id {
	case 1: // max stack size - VarInt
		v, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(v)
	case 3: // damage - VarInt
		v, err := buf.ReadVarInt()
		if err != nil {
			return nil, err
		}
		w.WriteVarInt(v)
	case 4: // unbreakable - Boolean
		v, err := buf.ReadBool()
		if err != nil {
			return nil, err
		}
		w.WriteBool(v)
	default:
		// unknown component - can't decode without knowing size
		return nil, nil
	}
	return w.Bytes(), nil
}

func TestSlot_Empty(t *testing.T) {
	slot := EmptySlot()

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// empty slot should be a single VarInt 0
	if buf.Len() != 1 {
		t.Errorf("Empty slot should be 1 byte, got %d", buf.Len())
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !decoded.IsEmpty() {
		t.Error("Decoded slot should be empty")
	}
}

func TestSlot_ItemOnly(t *testing.T) {
	slot := NewSlot(1, 64) // stone, 64 count

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Count != 64 {
		t.Errorf("Count mismatch: got %d, want %d", decoded.Count, 64)
	}
	if decoded.ItemID != 1 {
		t.Errorf("ItemID mismatch: got %d, want %d", decoded.ItemID, 1)
	}
	if len(decoded.Components.Add) != 0 {
		t.Errorf("Unexpected components: %d", len(decoded.Components.Add))
	}
}

func TestSlot_WithRawComponent(t *testing.T) {
	slot := NewSlot(100, 1)

	// add damage component (ID 3) with VarInt value 50
	damageData := NewWriter()
	damageData.WriteVarInt(50)
	slot.AddComponent(3, damageData.Bytes())

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Add) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(decoded.Components.Add))
	}

	comp := decoded.Components.Add[0]
	if comp.ID != 3 {
		t.Errorf("Component ID mismatch: got %d, want 3", comp.ID)
	}

	// decode the raw damage value
	compBuf := NewReader(comp.Data)
	damage, err := compBuf.ReadVarInt()
	if err != nil {
		t.Fatalf("Failed to read damage: %v", err)
	}
	if damage != 50 {
		t.Errorf("Damage mismatch: got %d, want 50", damage)
	}
}

func TestSlot_WithMultipleComponents(t *testing.T) {
	slot := NewSlot(100, 1)

	// damage component (ID 3)
	damageData := NewWriter()
	damageData.WriteVarInt(25)
	slot.AddComponent(3, damageData.Bytes())

	// max stack size component (ID 1)
	stackData := NewWriter()
	stackData.WriteVarInt(1)
	slot.AddComponent(1, stackData.Bytes())

	// unbreakable component (ID 4)
	unbData := NewWriter()
	unbData.WriteBool(true)
	slot.AddComponent(4, unbData.Bytes())

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Add) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(decoded.Components.Add))
	}

	// verify damage component
	dmg := decoded.GetComponent(3)
	if dmg == nil {
		t.Error("Missing damage component")
	} else {
		dmgBuf := NewReader(dmg.Data)
		dmgVal, _ := dmgBuf.ReadVarInt()
		if dmgVal != 25 {
			t.Errorf("Damage mismatch: got %d, want 25", dmgVal)
		}
	}

	// verify max stack size
	maxStack := decoded.GetComponent(1)
	if maxStack == nil {
		t.Error("Missing max stack size component")
	} else {
		stackBuf := NewReader(maxStack.Data)
		stackVal, _ := stackBuf.ReadVarInt()
		if stackVal != 1 {
			t.Errorf("MaxStackSize mismatch: got %d, want 1", stackVal)
		}
	}

	// verify unbreakable
	unb := decoded.GetComponent(4)
	if unb == nil {
		t.Error("Missing unbreakable component")
	} else {
		unbBuf := NewReader(unb.Data)
		unbVal, _ := unbBuf.ReadBool()
		if !bool(unbVal) {
			t.Error("ShowInTooltip should be true")
		}
	}
}

func TestSlot_WithRemovedComponents(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.RemoveComponent(3)  // damage
	slot.RemoveComponent(12) // enchantments

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Remove) != 2 {
		t.Fatalf("Expected 2 removed components, got %d", len(decoded.Components.Remove))
	}
	if decoded.Components.Remove[0] != 3 {
		t.Errorf("First removed component should be 3 (damage)")
	}
	if decoded.Components.Remove[1] != 12 {
		t.Errorf("Second removed component should be 12 (enchantments)")
	}
}

func TestSlot_RoundTrip(t *testing.T) {
	slot := NewSlot(100, 32)

	// damage
	damageData := NewWriter()
	damageData.WriteVarInt(10)
	slot.AddComponent(3, damageData.Bytes())

	// max stack size
	stackData := NewWriter()
	stackData.WriteVarInt(64)
	slot.AddComponent(1, stackData.Bytes())

	slot.RemoveComponent(99)

	// encode
	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	encoded := buf.Bytes()

	// decode
	readBuf := NewReader(encoded)
	decoded, err := readBuf.ReadSlot(testSlotDecoder)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// re-encode
	buf2 := NewWriter()
	if err := decoded.Encode(buf2); err != nil {
		t.Fatalf("Re-encode failed: %v", err)
	}
	reencoded := buf2.Bytes()

	// compare
	if !bytes.Equal(encoded, reencoded) {
		t.Errorf("Round-trip encoding mismatch:\n  original:  %x\n  reencoded: %x", encoded, reencoded)
	}
}

func TestSlot_PacketBufferHelpers(t *testing.T) {
	slot := NewSlot(50, 16)

	// max damage
	maxDmgData := NewWriter()
	maxDmgData.WriteVarInt(100)
	slot.AddComponent(2, maxDmgData.Bytes()) // max damage = ID 2

	buf := NewWriter()
	if err := buf.WriteSlot(slot); err != nil {
		t.Fatalf("WriteSlot failed: %v", err)
	}

	// use a decoder that handles ID 2
	decoder := func(buf *PacketBuffer, id VarInt) ([]byte, error) {
		if id == 2 {
			w := NewWriter()
			v, err := buf.ReadVarInt()
			if err != nil {
				return nil, err
			}
			w.WriteVarInt(v)
			return w.Bytes(), nil
		}
		return nil, nil
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot(decoder)
	if err != nil {
		t.Fatalf("ReadSlot failed: %v", err)
	}

	if decoded.Count != slot.Count {
		t.Errorf("Count mismatch")
	}
	if decoded.ItemID != slot.ItemID {
		t.Errorf("ItemID mismatch")
	}
}

func TestSlot_GetComponent(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(3, []byte{0x32})
	slot.AddComponent(5, []byte{0x01, 0x02})

	comp := slot.GetComponent(3)
	if comp == nil {
		t.Fatal("GetComponent returned nil for existing component")
	}
	if comp.ID != 3 {
		t.Errorf("Component ID mismatch: got %d, want 3", comp.ID)
	}

	comp2 := slot.GetComponent(999)
	if comp2 != nil {
		t.Error("GetComponent should return nil for non-existent component")
	}
}
