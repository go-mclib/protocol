package net_structures

import (
	"bytes"
	"testing"
)

// testSlotDecoder decodes known simple component types for testing.
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
		return nil, nil
	}
	return w.Bytes(), nil
}

// Slot wire format:
//   VarInt count (0 = empty, no further data)
//   VarInt itemID
//   VarInt addCount
//   VarInt removeCount
//   [addCount × (VarInt componentID + component data)]
//   [removeCount × VarInt componentID]

var slotTestCases = []struct {
	name     string
	raw      []byte
	expected Slot
}{
	{
		name:     "empty slot",
		raw:      []byte{0x00},
		expected: Slot{Count: 0},
	},
	{
		name: "stone x64 no components",
		// count=64 (0x40), itemID=1 (0x01), add=0, remove=0
		raw:      []byte{0x40, 0x01, 0x00, 0x00},
		expected: Slot{Count: 64, ItemID: 1},
	},
	{
		name: "diamond x1 no components",
		// count=1, itemID=264 (0x88 0x02), add=0, remove=0
		raw:      []byte{0x01, 0x88, 0x02, 0x00, 0x00},
		expected: Slot{Count: 1, ItemID: 264},
	},
	{
		name: "item with one removed component",
		// count=1, itemID=1, add=0, remove=1, removeID=3 (damage)
		raw: []byte{0x01, 0x01, 0x00, 0x01, 0x03},
		expected: Slot{
			Count:  1,
			ItemID: 1,
			Components: SlotComponents{
				Remove: []VarInt{3},
			},
		},
	},
	{
		name: "item with two removed components",
		// count=1, itemID=1, add=0, remove=2, removeIDs=3,12
		raw: []byte{0x01, 0x01, 0x00, 0x02, 0x03, 0x0c},
		expected: Slot{
			Count:  1,
			ItemID: 1,
			Components: SlotComponents{
				Remove: []VarInt{3, 12},
			},
		},
	},
	{
		name: "item with damage component",
		// count=1, itemID=100 (0x64), add=1, remove=0
		// component: id=3, data=VarInt(50)=0x32
		raw: []byte{0x01, 0x64, 0x01, 0x00, 0x03, 0x32},
		expected: Slot{
			Count:  1,
			ItemID: 100,
			Components: SlotComponents{
				Add: []RawSlotComponent{
					{ID: 3, Data: []byte{0x32}},
				},
			},
		},
	},
	{
		name: "item with max stack size component",
		// count=16, itemID=50 (0x32), add=1, remove=0
		// component: id=1, data=VarInt(16)=0x10
		raw: []byte{0x10, 0x32, 0x01, 0x00, 0x01, 0x10},
		expected: Slot{
			Count:  16,
			ItemID: 50,
			Components: SlotComponents{
				Add: []RawSlotComponent{
					{ID: 1, Data: []byte{0x10}},
				},
			},
		},
	},
	{
		name: "item with multiple components",
		// count=1, itemID=100, add=2, remove=1
		// comp1: id=3 (damage), data=VarInt(25)=0x19
		// comp2: id=1 (max stack), data=VarInt(1)=0x01
		// remove: id=4
		raw: []byte{0x01, 0x64, 0x02, 0x01, 0x03, 0x19, 0x01, 0x01, 0x04},
		expected: Slot{
			Count:  1,
			ItemID: 100,
			Components: SlotComponents{
				Add: []RawSlotComponent{
					{ID: 3, Data: []byte{0x19}},
					{ID: 1, Data: []byte{0x01}},
				},
				Remove: []VarInt{4},
			},
		},
	},
}

func TestSlot(t *testing.T) {
	for _, tc := range slotTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			buf := NewReader(tc.raw)
			got, err := buf.ReadSlot(testSlotDecoder)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if !slotEqual(got, tc.expected) {
				t.Errorf("decode mismatch:\n  got:  %+v\n  want: %+v", got, tc.expected)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			buf := NewWriter()
			if err := tc.expected.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func slotEqual(a, b Slot) bool {
	if a.Count != b.Count || a.ItemID != b.ItemID {
		return false
	}
	if len(a.Components.Add) != len(b.Components.Add) {
		return false
	}
	for i := range a.Components.Add {
		if a.Components.Add[i].ID != b.Components.Add[i].ID {
			return false
		}
		if !bytes.Equal(a.Components.Add[i].Data, b.Components.Add[i].Data) {
			return false
		}
	}
	if len(a.Components.Remove) != len(b.Components.Remove) {
		return false
	}
	for i := range a.Components.Remove {
		if a.Components.Remove[i] != b.Components.Remove[i] {
			return false
		}
	}
	return true
}

func TestSlot_GetComponent(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(3, []byte{0x32})
	slot.AddComponent(5, []byte{0x01, 0x02})

	if comp := slot.GetComponent(3); comp == nil || comp.ID != 3 {
		t.Error("GetComponent(3) failed")
	}
	if comp := slot.GetComponent(999); comp != nil {
		t.Error("GetComponent(999) should return nil")
	}
}
