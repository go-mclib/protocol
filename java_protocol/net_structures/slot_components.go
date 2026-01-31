package net_structures

import (
	"fmt"

	"github.com/go-mclib/protocol/nbt"
)

// Component type IDs as defined by the protocol.
// See: https://minecraft.wiki/w/Data_component_format
const (
	ComponentCustomData      VarInt = 0
	ComponentMaxStackSize    VarInt = 1
	ComponentMaxDamage       VarInt = 2
	ComponentDamage          VarInt = 3
	ComponentUnbreakable     VarInt = 4
	ComponentCustomName      VarInt = 7
	ComponentItemName        VarInt = 8
	ComponentItemModel       VarInt = 9
	ComponentLore            VarInt = 10
	ComponentRarity          VarInt = 11
	ComponentEnchantments    VarInt = 12
	ComponentRepairCost      VarInt = 17
	ComponentDyedColor       VarInt = 26
	ComponentCustomModelData VarInt = 24
)

// componentDecoders maps component IDs to factory functions.
var componentDecoders = map[VarInt]func() SlotComponent{
	ComponentCustomData:   func() SlotComponent { return &CustomDataComponent{} },
	ComponentMaxStackSize: func() SlotComponent { return &MaxStackSizeComponent{} },
	ComponentMaxDamage:    func() SlotComponent { return &MaxDamageComponent{} },
	ComponentDamage:       func() SlotComponent { return &DamageComponent{} },
	ComponentUnbreakable:  func() SlotComponent { return &UnbreakableComponent{} },
	ComponentCustomName:   func() SlotComponent { return &CustomNameComponent{} },
	ComponentItemName:     func() SlotComponent { return &ItemNameComponent{} },
	ComponentLore:         func() SlotComponent { return &LoreComponent{} },
	ComponentRarity:       func() SlotComponent { return &RarityComponent{} },
	ComponentEnchantments: func() SlotComponent { return &EnchantmentsComponent{} },
	ComponentRepairCost:   func() SlotComponent { return &RepairCostComponent{} },
	ComponentDyedColor:    func() SlotComponent { return &DyedColorComponent{} },
}

// NewSlotComponent creates a slot component for the given ID.
// Returns a RawComponent for unknown component types.
func NewSlotComponent(id VarInt) SlotComponent {
	if factory, ok := componentDecoders[id]; ok {
		return factory()
	}
	return &RawComponent{ID: id}
}

// RegisterComponentDecoder registers a custom component decoder.
func RegisterComponentDecoder(id VarInt, factory func() SlotComponent) {
	componentDecoders[id] = factory
}

// -----------------------------------------------------------------------------
// RawComponent - fallback for unknown components
// -----------------------------------------------------------------------------

// RawComponent stores unknown component data as raw bytes for passthrough.
type RawComponent struct {
	ID   VarInt
	Data []byte
}

func (c *RawComponent) ComponentID() VarInt { return c.ID }

func (c *RawComponent) Encode(buf *PacketBuffer) error {
	_, err := buf.Write(c.Data)
	return err
}

func (c *RawComponent) Decode(buf *PacketBuffer) error {
	// we can't know the size of unknown components without a length prefix,
	// so this is a best-effort that reads until EOF or stores empty
	// in practice, this should only be used when the full packet data is available
	// and can be parsed completely
	c.Data = nil
	return nil
}

// -----------------------------------------------------------------------------
// CustomDataComponent - arbitrary NBT data (ID 0)
// -----------------------------------------------------------------------------

// CustomDataComponent holds arbitrary NBT data attached to an item.
type CustomDataComponent struct {
	Data nbt.Tag
}

func (*CustomDataComponent) ComponentID() VarInt { return ComponentCustomData }

func (c *CustomDataComponent) Encode(buf *PacketBuffer) error {
	if c.Data == nil {
		c.Data = nbt.Compound{}
	}
	data, err := nbt.Encode(c.Data, "", true)
	if err != nil {
		return err
	}
	_, err = buf.Write(data)
	return err
}

func (c *CustomDataComponent) Decode(buf *PacketBuffer) error {
	nbtReader := nbt.NewReaderFrom(buf.Reader())
	tag, _, err := nbtReader.ReadTag(true)
	if err != nil {
		return err
	}
	c.Data = tag
	return nil
}

// -----------------------------------------------------------------------------
// MaxStackSizeComponent - max stack size override (ID 1)
// -----------------------------------------------------------------------------

// MaxStackSizeComponent overrides the maximum stack size of an item.
type MaxStackSizeComponent struct {
	MaxStackSize VarInt
}

func (*MaxStackSizeComponent) ComponentID() VarInt { return ComponentMaxStackSize }

func (c *MaxStackSizeComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteVarInt(c.MaxStackSize)
}

func (c *MaxStackSizeComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadVarInt()
	c.MaxStackSize = v
	return err
}

// -----------------------------------------------------------------------------
// MaxDamageComponent - max damage/durability (ID 2)
// -----------------------------------------------------------------------------

// MaxDamageComponent sets the maximum damage (durability) of an item.
type MaxDamageComponent struct {
	MaxDamage VarInt
}

func (*MaxDamageComponent) ComponentID() VarInt { return ComponentMaxDamage }

func (c *MaxDamageComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteVarInt(c.MaxDamage)
}

func (c *MaxDamageComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadVarInt()
	c.MaxDamage = v
	return err
}

// -----------------------------------------------------------------------------
// DamageComponent - current damage value (ID 3)
// -----------------------------------------------------------------------------

// DamageComponent holds the current damage value of an item.
type DamageComponent struct {
	Damage VarInt
}

func (*DamageComponent) ComponentID() VarInt { return ComponentDamage }

func (c *DamageComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteVarInt(c.Damage)
}

func (c *DamageComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadVarInt()
	c.Damage = v
	return err
}

// -----------------------------------------------------------------------------
// UnbreakableComponent - unbreakable flag (ID 4)
// -----------------------------------------------------------------------------

// UnbreakableComponent marks an item as unbreakable.
type UnbreakableComponent struct {
	ShowInTooltip bool
}

func (*UnbreakableComponent) ComponentID() VarInt { return ComponentUnbreakable }

func (c *UnbreakableComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteBool(Boolean(c.ShowInTooltip))
}

func (c *UnbreakableComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadBool()
	c.ShowInTooltip = bool(v)
	return err
}

// -----------------------------------------------------------------------------
// CustomNameComponent - custom display name (ID 7)
// -----------------------------------------------------------------------------

// CustomNameComponent sets a custom display name for an item.
type CustomNameComponent struct {
	Name TextComponent
}

func (*CustomNameComponent) ComponentID() VarInt { return ComponentCustomName }

func (c *CustomNameComponent) Encode(buf *PacketBuffer) error {
	return c.Name.Encode(buf)
}

func (c *CustomNameComponent) Decode(buf *PacketBuffer) error {
	return c.Name.Decode(buf)
}

// -----------------------------------------------------------------------------
// ItemNameComponent - item name override (ID 8)
// -----------------------------------------------------------------------------

// ItemNameComponent overrides the default item name (non-italicized).
type ItemNameComponent struct {
	Name TextComponent
}

func (*ItemNameComponent) ComponentID() VarInt { return ComponentItemName }

func (c *ItemNameComponent) Encode(buf *PacketBuffer) error {
	return c.Name.Encode(buf)
}

func (c *ItemNameComponent) Decode(buf *PacketBuffer) error {
	return c.Name.Decode(buf)
}

// -----------------------------------------------------------------------------
// LoreComponent - item lore lines (ID 10)
// -----------------------------------------------------------------------------

// LoreComponent holds the lore lines for an item.
type LoreComponent struct {
	Lines []TextComponent
}

func (*LoreComponent) ComponentID() VarInt { return ComponentLore }

func (c *LoreComponent) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(len(c.Lines))); err != nil {
		return err
	}
	for i, line := range c.Lines {
		if err := line.Encode(buf); err != nil {
			return fmt.Errorf("failed to encode lore line %d: %w", i, err)
		}
	}
	return nil
}

func (c *LoreComponent) Decode(buf *PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	c.Lines = make([]TextComponent, count)
	for i := range c.Lines {
		if err := c.Lines[i].Decode(buf); err != nil {
			return fmt.Errorf("failed to decode lore line %d: %w", i, err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// RarityComponent - item rarity (ID 11)
// -----------------------------------------------------------------------------

// Rarity values.
const (
	RarityCommon   VarInt = 0
	RarityUncommon VarInt = 1
	RarityRare     VarInt = 2
	RarityEpic     VarInt = 3
)

// RarityComponent sets the rarity of an item.
type RarityComponent struct {
	Rarity VarInt
}

func (*RarityComponent) ComponentID() VarInt { return ComponentRarity }

func (c *RarityComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteVarInt(c.Rarity)
}

func (c *RarityComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadVarInt()
	c.Rarity = v
	return err
}

// -----------------------------------------------------------------------------
// EnchantmentsComponent - item enchantments (ID 12)
// -----------------------------------------------------------------------------

// EnchantmentsComponent holds the enchantments on an item.
type EnchantmentsComponent struct {
	Enchantments  map[VarInt]VarInt // enchantment ID -> level
	ShowInTooltip bool
}

func (*EnchantmentsComponent) ComponentID() VarInt { return ComponentEnchantments }

func (c *EnchantmentsComponent) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(len(c.Enchantments))); err != nil {
		return err
	}
	for id, level := range c.Enchantments {
		if err := buf.WriteVarInt(id); err != nil {
			return err
		}
		if err := buf.WriteVarInt(level); err != nil {
			return err
		}
	}
	return buf.WriteBool(Boolean(c.ShowInTooltip))
}

func (c *EnchantmentsComponent) Decode(buf *PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	c.Enchantments = make(map[VarInt]VarInt, count)
	for range count {
		id, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		level, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		c.Enchantments[id] = level
	}
	show, err := buf.ReadBool()
	c.ShowInTooltip = bool(show)
	return err
}

// -----------------------------------------------------------------------------
// RepairCostComponent - anvil repair cost (ID 17)
// -----------------------------------------------------------------------------

// RepairCostComponent holds the anvil repair cost of an item.
type RepairCostComponent struct {
	Cost VarInt
}

func (*RepairCostComponent) ComponentID() VarInt { return ComponentRepairCost }

func (c *RepairCostComponent) Encode(buf *PacketBuffer) error {
	return buf.WriteVarInt(c.Cost)
}

func (c *RepairCostComponent) Decode(buf *PacketBuffer) error {
	v, err := buf.ReadVarInt()
	c.Cost = v
	return err
}

// -----------------------------------------------------------------------------
// DyedColorComponent - leather armor dye color (ID 26)
// -----------------------------------------------------------------------------

// DyedColorComponent holds the dye color for leather armor.
type DyedColorComponent struct {
	Color         Int32
	ShowInTooltip bool
}

func (*DyedColorComponent) ComponentID() VarInt { return ComponentDyedColor }

func (c *DyedColorComponent) Encode(buf *PacketBuffer) error {
	if err := buf.WriteInt32(c.Color); err != nil {
		return err
	}
	return buf.WriteBool(Boolean(c.ShowInTooltip))
}

func (c *DyedColorComponent) Decode(buf *PacketBuffer) error {
	color, err := buf.ReadInt32()
	if err != nil {
		return err
	}
	c.Color = color
	show, err := buf.ReadBool()
	c.ShowInTooltip = bool(show)
	return err
}

// -----------------------------------------------------------------------------
// Helpers for raw component passthrough
// -----------------------------------------------------------------------------

// NewRawComponent creates a RawComponent from encoded bytes.
// Useful for proxying packets without fully parsing component data.
func NewRawComponent(id VarInt, data []byte) *RawComponent {
	return &RawComponent{ID: id, Data: data}
}

// EncodeComponentToBytes encodes a component to bytes.
func EncodeComponentToBytes(comp SlotComponent) ([]byte, error) {
	buf := NewWriter()
	if err := comp.Encode(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeComponentFromBytes decodes a component from bytes.
func DecodeComponentFromBytes(id VarInt, data []byte) (SlotComponent, error) {
	comp := NewSlotComponent(id)
	buf := NewReader(data)
	if err := comp.Decode(buf); err != nil {
		return nil, err
	}
	return comp, nil
}

// CaptureRawComponent reads a component and captures its raw bytes.
// This is useful for proxying when you want to inspect components but
// also preserve their exact encoding.
func CaptureRawComponent(buf *PacketBuffer, id VarInt) (*RawComponent, SlotComponent, error) {
	// for components we know, capture the bytes by encoding after decode
	comp := NewSlotComponent(id)

	// if it's already a raw component, we can't capture it
	if raw, ok := comp.(*RawComponent); ok {
		return raw, raw, nil
	}

	// decode the component
	if err := comp.Decode(buf); err != nil {
		return nil, nil, err
	}

	// re-encode to get the raw bytes
	data, err := EncodeComponentToBytes(comp)
	if err != nil {
		return nil, nil, err
	}

	return &RawComponent{ID: id, Data: data}, comp, nil
}

// DecodeSlotWithRawComponents decodes a slot while capturing raw component data.
// This is useful for proxies that need to inspect and potentially modify components.
func DecodeSlotWithRawComponents(buf *PacketBuffer) (Slot, [][]byte, error) {
	var slot Slot

	count, err := buf.ReadVarInt()
	if err != nil {
		return slot, nil, fmt.Errorf("failed to read slot count: %w", err)
	}
	slot.Count = count

	if slot.Count <= 0 {
		return slot, nil, nil
	}

	slot.ItemID, err = buf.ReadVarInt()
	if err != nil {
		return slot, nil, fmt.Errorf("failed to read slot item id: %w", err)
	}

	addCount, err := buf.ReadVarInt()
	if err != nil {
		return slot, nil, fmt.Errorf("failed to read slot add count: %w", err)
	}

	removeCount, err := buf.ReadVarInt()
	if err != nil {
		return slot, nil, fmt.Errorf("failed to read slot remove count: %w", err)
	}

	// read and capture added components
	slot.Components.Add = make([]SlotComponent, addCount)
	rawData := make([][]byte, addCount)

	for i := range slot.Components.Add {
		compID, err := buf.ReadVarInt()
		if err != nil {
			return slot, nil, fmt.Errorf("failed to read component %d id: %w", i, err)
		}

		// capture the component by reading into a buffer first
		raw, comp, err := CaptureRawComponent(buf, compID)
		if err != nil {
			return slot, nil, fmt.Errorf("failed to read component %d: %w", i, err)
		}

		slot.Components.Add[i] = comp
		rawData[i] = raw.Data
	}

	// read removed component IDs
	slot.Components.Remove = make([]VarInt, removeCount)
	for i := range slot.Components.Remove {
		slot.Components.Remove[i], err = buf.ReadVarInt()
		if err != nil {
			return slot, nil, fmt.Errorf("failed to read removed component %d id: %w", i, err)
		}
	}

	return slot, rawData, nil
}
