package net_structures

import (
	"bytes"
	"testing"

	"github.com/go-mclib/protocol/nbt"
)

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
	decoded, err := readBuf.ReadSlot()
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
	decoded, err := readBuf.ReadSlot()
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

func TestSlot_WithDamageComponent(t *testing.T) {
	slot := NewSlot(100, 1) // some tool
	slot.AddComponent(&DamageComponent{Damage: 50})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Add) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(decoded.Components.Add))
	}

	dmg, ok := decoded.Components.Add[0].(*DamageComponent)
	if !ok {
		t.Fatalf("Expected DamageComponent, got %T", decoded.Components.Add[0])
	}
	if dmg.Damage != 50 {
		t.Errorf("Damage mismatch: got %d, want %d", dmg.Damage, 50)
	}
}

func TestSlot_WithMultipleComponents(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(&DamageComponent{Damage: 25})
	slot.AddComponent(&MaxStackSizeComponent{MaxStackSize: 1})
	slot.AddComponent(&UnbreakableComponent{ShowInTooltip: true})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Add) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(decoded.Components.Add))
	}

	// verify damage component
	dmg := decoded.GetComponent(ComponentDamage)
	if dmg == nil {
		t.Error("Missing DamageComponent")
	} else if dmg.(*DamageComponent).Damage != 25 {
		t.Errorf("Damage mismatch: got %d, want %d", dmg.(*DamageComponent).Damage, 25)
	}

	// verify max stack size
	maxStack := decoded.GetComponent(ComponentMaxStackSize)
	if maxStack == nil {
		t.Error("Missing MaxStackSizeComponent")
	} else if maxStack.(*MaxStackSizeComponent).MaxStackSize != 1 {
		t.Errorf("MaxStackSize mismatch")
	}

	// verify unbreakable
	unbreakable := decoded.GetComponent(ComponentUnbreakable)
	if unbreakable == nil {
		t.Error("Missing UnbreakableComponent")
	} else if !unbreakable.(*UnbreakableComponent).ShowInTooltip {
		t.Error("ShowInTooltip should be true")
	}
}

func TestSlot_WithCustomName(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(&CustomNameComponent{
		Name: TextComponent{
			Text:  "Epic Sword",
			Color: "gold",
		},
	})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	name := decoded.GetComponent(ComponentCustomName)
	if name == nil {
		t.Fatal("Missing CustomNameComponent")
	}
	customName := name.(*CustomNameComponent)
	if customName.Name.Text != "Epic Sword" {
		t.Errorf("Name.Text mismatch: got %q, want %q", customName.Name.Text, "Epic Sword")
	}
	if customName.Name.Color != "gold" {
		t.Errorf("Name.Color mismatch: got %q, want %q", customName.Name.Color, "gold")
	}
}

func TestSlot_WithLore(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(&LoreComponent{
		Lines: []TextComponent{
			{Text: "Line 1", Color: "gray"},
			{Text: "Line 2", Color: "dark_gray"},
		},
	})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	lore := decoded.GetComponent(ComponentLore)
	if lore == nil {
		t.Fatal("Missing LoreComponent")
	}
	loreComp := lore.(*LoreComponent)
	if len(loreComp.Lines) != 2 {
		t.Fatalf("Expected 2 lore lines, got %d", len(loreComp.Lines))
	}
	if loreComp.Lines[0].Text != "Line 1" {
		t.Errorf("Lore line 0 mismatch: got %q, want %q", loreComp.Lines[0].Text, "Line 1")
	}
}

func TestSlot_WithEnchantments(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(&EnchantmentsComponent{
		Enchantments: map[VarInt]VarInt{
			1: 3, // sharpness 3
			2: 1, // smite 1
		},
		ShowInTooltip: true,
	})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	ench := decoded.GetComponent(ComponentEnchantments)
	if ench == nil {
		t.Fatal("Missing EnchantmentsComponent")
	}
	enchComp := ench.(*EnchantmentsComponent)
	if len(enchComp.Enchantments) != 2 {
		t.Fatalf("Expected 2 enchantments, got %d", len(enchComp.Enchantments))
	}
	if enchComp.Enchantments[1] != 3 {
		t.Errorf("Sharpness level mismatch: got %d, want %d", enchComp.Enchantments[1], 3)
	}
	if !enchComp.ShowInTooltip {
		t.Error("ShowInTooltip should be true")
	}
}

func TestSlot_WithCustomData(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.AddComponent(&CustomDataComponent{
		Data: nbt.Compound{
			"CustomKey":  nbt.String("custom value"),
			"CustomInt":  nbt.Int(42),
			"NestedData": nbt.Compound{"inner": nbt.Byte(1)},
		},
	})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	customData := decoded.GetComponent(ComponentCustomData)
	if customData == nil {
		t.Fatal("Missing CustomDataComponent")
	}
	dataComp := customData.(*CustomDataComponent)
	compound, ok := dataComp.Data.(nbt.Compound)
	if !ok {
		t.Fatalf("Expected nbt.Compound, got %T", dataComp.Data)
	}
	if compound.GetString("CustomKey") != "custom value" {
		t.Errorf("CustomKey mismatch")
	}
	if compound.GetInt("CustomInt") != 42 {
		t.Errorf("CustomInt mismatch")
	}
}

func TestSlot_WithRemovedComponents(t *testing.T) {
	slot := NewSlot(100, 1)
	slot.RemoveComponent(ComponentDamage)
	slot.RemoveComponent(ComponentEnchantments)

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Components.Remove) != 2 {
		t.Fatalf("Expected 2 removed components, got %d", len(decoded.Components.Remove))
	}
	if decoded.Components.Remove[0] != ComponentDamage {
		t.Errorf("First removed component should be damage")
	}
	if decoded.Components.Remove[1] != ComponentEnchantments {
		t.Errorf("Second removed component should be enchantments")
	}
}

func TestSlot_DyedColor(t *testing.T) {
	slot := NewSlot(100, 1) // leather armor
	slot.AddComponent(&DyedColorComponent{
		Color:         0xFF5500, // orange
		ShowInTooltip: true,
	})

	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	dyed := decoded.GetComponent(ComponentDyedColor)
	if dyed == nil {
		t.Fatal("Missing DyedColorComponent")
	}
	dyedComp := dyed.(*DyedColorComponent)
	if dyedComp.Color != 0xFF5500 {
		t.Errorf("Color mismatch: got %x, want %x", dyedComp.Color, 0xFF5500)
	}
	if !dyedComp.ShowInTooltip {
		t.Error("ShowInTooltip should be true")
	}
}

func TestSlot_RoundTrip(t *testing.T) {
	slot := NewSlot(100, 32)
	slot.AddComponent(&DamageComponent{Damage: 10})
	slot.AddComponent(&CustomNameComponent{
		Name: TextComponent{Text: "Test Item", Color: "blue"},
	})
	slot.AddComponent(&EnchantmentsComponent{
		Enchantments:  map[VarInt]VarInt{5: 2},
		ShowInTooltip: false,
	})
	slot.RemoveComponent(99) // some hypothetical component

	// encode
	buf := NewWriter()
	if err := slot.Encode(buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	encoded := buf.Bytes()

	// decode
	readBuf := NewReader(encoded)
	decoded, err := readBuf.ReadSlot()
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
	slot.AddComponent(&MaxDamageComponent{MaxDamage: 100})

	buf := NewWriter()
	if err := buf.WriteSlot(slot); err != nil {
		t.Fatalf("WriteSlot failed: %v", err)
	}

	readBuf := NewReader(buf.Bytes())
	decoded, err := readBuf.ReadSlot()
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

func TestNewSlotComponent_KnownTypes(t *testing.T) {
	tests := []struct {
		id       VarInt
		expected string
	}{
		{ComponentCustomData, "*net_structures.CustomDataComponent"},
		{ComponentMaxStackSize, "*net_structures.MaxStackSizeComponent"},
		{ComponentDamage, "*net_structures.DamageComponent"},
		{ComponentCustomName, "*net_structures.CustomNameComponent"},
		{ComponentLore, "*net_structures.LoreComponent"},
		{ComponentEnchantments, "*net_structures.EnchantmentsComponent"},
		{ComponentDyedColor, "*net_structures.DyedColorComponent"},
	}

	for _, tt := range tests {
		comp := NewSlotComponent(tt.id)
		typeName := typeNameOf(comp)
		if typeName != tt.expected {
			t.Errorf("NewSlotComponent(%d) = %s, want %s", tt.id, typeName, tt.expected)
		}
	}
}

func TestNewSlotComponent_UnknownType(t *testing.T) {
	comp := NewSlotComponent(9999) // unknown ID
	raw, ok := comp.(*RawComponent)
	if !ok {
		t.Errorf("Expected RawComponent for unknown ID, got %T", comp)
	}
	if raw.ID != 9999 {
		t.Errorf("RawComponent.ID = %d, want %d", raw.ID, 9999)
	}
}

func typeNameOf(v any) string {
	return typeName(v)
}

func typeName(v any) string {
	if v == nil {
		return "<nil>"
	}
	return typeNameStr(v)
}

func typeNameStr(v any) string {
	return typeNameFromInterface(v)
}

func typeNameFromInterface(v any) string {
	switch v.(type) {
	case *CustomDataComponent:
		return "*net_structures.CustomDataComponent"
	case *MaxStackSizeComponent:
		return "*net_structures.MaxStackSizeComponent"
	case *MaxDamageComponent:
		return "*net_structures.MaxDamageComponent"
	case *DamageComponent:
		return "*net_structures.DamageComponent"
	case *UnbreakableComponent:
		return "*net_structures.UnbreakableComponent"
	case *CustomNameComponent:
		return "*net_structures.CustomNameComponent"
	case *ItemNameComponent:
		return "*net_structures.ItemNameComponent"
	case *LoreComponent:
		return "*net_structures.LoreComponent"
	case *RarityComponent:
		return "*net_structures.RarityComponent"
	case *EnchantmentsComponent:
		return "*net_structures.EnchantmentsComponent"
	case *RepairCostComponent:
		return "*net_structures.RepairCostComponent"
	case *DyedColorComponent:
		return "*net_structures.DyedColorComponent"
	case *RawComponent:
		return "*net_structures.RawComponent"
	default:
		return "unknown"
	}
}
