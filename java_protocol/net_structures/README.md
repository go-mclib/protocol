# net_structures

Go implementation of Minecraft Java Edition protocol data types.

Based on the [Minecraft Wiki protocol documentation](https://minecraft.wiki/w/Java_Edition_protocol/Data_types) and the decompiled 1.21.11 client source code.

## Data Types

### Primitives

All multi-byte integers use **big-endian** byte order (same as Java/Netty). These map directly to Go's `encoding/binary.BigEndian`.

| Protocol Type | Go Type | Size | Notes |
| ------------- | ------- | ---- | ----- |
| Boolean | `Boolean` | 1 | `0x00` = false, `0x01` = true |
| Byte | `Int8` | 1 | Signed 8-bit |
| Unsigned Byte | `Uint8` | 1 | Unsigned 8-bit |
| Short | `Int16` | 2 | Signed 16-bit |
| Unsigned Short | `Uint16` | 2 | Unsigned 16-bit |
| Int | `Int32` | 4 | Signed 32-bit |
| Long | `Int64` | 8 | Signed 64-bit |
| Float | `Float32` | 4 | IEEE 754 single-precision |
| Double | `Float64` | 8 | IEEE 754 double-precision |

### Variable-Length Integers

| Protocol Type | Go Type | Max Size | Notes |
| ------------- | ------- | -------- | ----- |
| VarInt | `VarInt` | 5 bytes | 7-bit encoding with continuation bit |
| VarLong | `VarLong` | 10 bytes | Same encoding for 64-bit |

Encoding: each byte uses bits 0-6 for data, bit 7 as continuation flag (1 = more bytes follow).

```plain
0          -> [0x00]
127        -> [0x7f]
128        -> [0x80, 0x01]
-1         -> [0xff, 0xff, 0xff, 0xff, 0x0f]
```

### Strings

| Protocol Type | Go Type | Notes |
| ------------- | ------- | ----- |
| String | `String` | VarInt byte-length prefix + UTF-8 bytes |
| Identifier | `Identifier` | Same as String, format: `namespace:path` |

### Complex Types

| Protocol Type | Go Type | Notes |
| ------------- | ------- | ----- |
| Position | `Position` | Block coordinates packed into int64: X(26 bits) + Z(26 bits) + Y(12 bits) |
| UUID | `UUID` | 128-bit, stored as `[16]byte` |
| Angle | `Angle` | Rotation in 1/256 of a full turn (1 byte) |
| Byte Array | `ByteArray` | VarInt length prefix + raw bytes |

### Composite Types

These types handle common patterns like length-prefixed arrays, boolean-prefixed optionals, and bit sets.

| Protocol Type | Go Type | Wire Format |
| ------------- | ------- | ----------- |
| Prefixed Array | `PrefixedArray[T]` | VarInt length + elements |
| Prefixed Optional | `PrefixedOptional[T]` | Boolean + value (if true) |
| BitSet | `BitSet` | VarInt length (in longs) + int64 array |
| Fixed BitSet | `FixedBitSet` | ceil(n/8) bytes (no length prefix) |
| ID Set | `IDSet` | VarInt type + tag name or IDs |

## Usage

```go
import ns "github.com/go-mclib/protocol/java_protocol/net_structures"

// writing
buf := ns.NewWriter()
buf.WriteVarInt(25565)
buf.WriteString("localhost")
buf.WriteUint16(25565)
buf.WritePosition(ns.Position{X: 100, Y: 64, Z: -200})
data := buf.Bytes()

// reading
buf := ns.NewReader(data)
version, _ := buf.ReadVarInt()
address, _ := buf.ReadString(255)
port, _ := buf.ReadUint16()
pos, _ := buf.ReadPosition()

// low-level streaming (directly with net.Conn)
buf := ns.NewWriterTo(conn)
buf.WriteVarInt(0x00)

buf := ns.NewReaderFrom(conn)
packetID, _ := buf.ReadVarInt()
```

### Composite Types Usage

```go
// PrefixedArray - VarInt length-prefixed array
type MyPacket struct {
    Names ns.PrefixedArray[ns.String]
}

func (p *MyPacket) Read(buf *ns.PacketBuffer) error {
    return p.Names.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.String, error) {
        return b.ReadString(32767)
    })
}

func (p *MyPacket) Write(buf *ns.PacketBuffer) error {
    return p.Names.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.String) error {
        return b.WriteString(v)
    })
}

// PrefixedOptional - Boolean-prefixed optional
type MyPacket2 struct {
    Title ns.PrefixedOptional[ns.String]
}

// create optionals
title := ns.Some("Hello")       // present
noTitle := ns.None[ns.String]() // absent

// BitSet - dynamic bit set
bits := ns.NewBitSet(128)
bits.Set(5)
bits.Get(5) // true
bits.Encode(buf)

// FixedBitSet - fixed-size bit set
fixed := ns.NewFixedBitSet(20) // 20 bits = 3 bytes
fixed.Set(0)
fixed.Encode(buf)

// IDSet - registry ID set
tagSet := ns.NewTagIDSet("minecraft:climbable")
inlineSet := ns.NewInlineIDSet([]ns.VarInt{1, 2, 3})
```

### NBT (Named Binary Tag)

NBT is used for complex structured data in packets. The `nbt` package supports both file format (with root name) and **network format** (nameless root) used in packets. For communication with the server, use the network format.

#### Direct decoding

```go
import "github.com/go-mclib/protocol/nbt"

type EntityData struct {
    Name     string `nbt:"Name"`
    Position int64  `nbt:"Position"`
    OnGround bool   `nbt:"OnGround"`
}

type S2CSomePacket struct {
    EntityID   ns.VarInt
    Data       EntityData
    ExtraField ns.VarInt
}

func (p *S2CSomePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.EntityID, err = buf.ReadVarInt(); err != nil {
        return err
    }

    // use nbt.NewReaderFrom to read the NBT data, which stops at TAG_End
    nbtReader := nbt.NewReaderFrom(buf.Reader())
    tag, _, err := nbtReader.ReadTag(true) // true = network format
    if err != nil {
        return err
    }
    if err := nbt.UnmarshalTag(tag, &p.Data); err != nil {
        return err
    }
    p.ExtraField, err = buf.ReadVarInt()
    return err
}

func (p *S2CSomePacket) Write(buf *ns.PacketBuffer) error {
    if err := buf.WriteVarInt(p.EntityID); err != nil {
        return err
    }
    nbtData, err := nbt.MarshalNetwork(p.Data)
    if err != nil {
        return err
    }
    if _, err := buf.Write(nbtData); err != nil {
        return err
    }
    return buf.WriteVarInt(p.ExtraField)
}
```

#### Storing as `nbt.Tag` (lazy processing)

For packets where you want to defer NBT processing (maybe the NBT data is too large, or dynamic):

```go
type S2CSomePacket struct {
    EntityID ns.VarInt
    Data     nbt.Tag   // store as generic Tag
}

func (p *S2CSomePacket) Read(buf *ns.PacketBuffer) error {
    var err error
    if p.EntityID, err = buf.ReadVarInt(); err != nil {
        return err
    }
    nbtReader := nbt.NewReaderFrom(buf.Reader())
    p.Data, _, err = nbtReader.ReadTag(true)
    return err
}

// later, convert to struct to grab values when needed
// extra fields in the NBT data that are not present in
// the EntityData struct will be skipped
var entityData EntityData
err := nbt.UnmarshalTag(packet.Data, &entityData)
```

#### Empty/Optional NBT

Some packets use a single `TAG_End` byte (`0x00`) to indicate empty or absent NBT data. Check for `nbt.End{}` type after reading:

```go
tag, _, err := nbtReader.ReadTag(true)
if _, isEmpty := tag.(nbt.End); isEmpty {
    // no NBT data present
}
```

### Text Component

Text components are used for chat messages, item names, titles, and other formatted text. Since 1.20.3+, they are encoded as binary NBT over the network (not JSON or SNBT). Simple text-only components use NBT String tags, complex components use NBT Compound tags.

```go
// simple text
tc := ns.NewTextComponent("Hello, World!")

// with style
bold := true
tc := ns.TextComponent{
    Text:  "Styled text",
    Color: "red",
    Bold:  &bold,
}

// translatable
tc := ns.NewTranslateComponent("chat.type.text",
    ns.NewTextComponent("Player"),
    ns.NewTextComponent("Hello"),
)

// with children
tc := ns.TextComponent{
    Text: "Hello, ",
    Extra: []ns.TextComponent{
        {Text: "World", Color: "gold"},
        {Text: "!"},
    },
}

// read/write
buf.WriteTextComponent(tc)
tc, _ := buf.ReadTextComponent()
```

### Slot (Item Stack)

Slots represent item stacks with data components. Used in inventory packets, container interactions, etc.

Wire format:

- `VarInt count` - item count (0 = empty slot)
- `VarInt item_id` - registry ID (only if count > 0)
- `VarInt add_count` - components to add
- `VarInt remove_count` - components to remove
- Component data (+96 different components, we will implement them incrementally as needed, probably separating it into a separate package at that point)...

```go
// empty slot
slot := ns.EmptySlot()

// basic item
slot := ns.NewSlot(1, 64) // stone, 64 count

// with components
slot := ns.NewSlot(100, 1)
slot.AddComponent(&ns.DamageComponent{Damage: 50})
slot.AddComponent(&ns.CustomNameComponent{
    Name: ns.TextComponent{Text: "Epic Sword", Color: "gold"},
})
slot.AddComponent(&ns.EnchantmentsComponent{
    Enchantments:  map[ns.VarInt]ns.VarInt{1: 5}, // sharpness 5
    ShowInTooltip: true,
})

// remove default components
slot.RemoveComponent(ns.ComponentDamage)

// read/write
buf.WriteSlot(slot)
slot, _ := buf.ReadSlot()

// get components
if dmg := slot.GetComponent(ns.ComponentDamage); dmg != nil {
    damage := dmg.(*ns.DamageComponent).Damage
}
```

#### Implemented Components

| ID | Constant | Type | Description |
| -- | -------- | ---- | ----------- |
| 0 | `ComponentCustomData` | `CustomDataComponent` | Arbitrary NBT data |
| 1 | `ComponentMaxStackSize` | `MaxStackSizeComponent` | Max stack size override |
| 2 | `ComponentMaxDamage` | `MaxDamageComponent` | Max durability |
| 3 | `ComponentDamage` | `DamageComponent` | Current damage |
| 4 | `ComponentUnbreakable` | `UnbreakableComponent` | Unbreakable flag |
| 7 | `ComponentCustomName` | `CustomNameComponent` | Custom display name |
| 8 | `ComponentItemName` | `ItemNameComponent` | Item name override |
| 10 | `ComponentLore` | `LoreComponent` | Lore lines |
| 11 | `ComponentRarity` | `RarityComponent` | Item rarity |
| 12 | `ComponentEnchantments` | `EnchantmentsComponent` | Enchantments |
| 17 | `ComponentRepairCost` | `RepairCostComponent` | Anvil repair cost |
| 26 | `ComponentDyedColor` | `DyedColorComponent` | Leather armor color |

Unknown component types are stored as `RawComponent` for passthrough.

## References

- [Minecraft Wiki - Data Types](https://minecraft.wiki/w/Java_Edition_protocol/Data_types)
- [Minecraft Wiki - Protocol](https://minecraft.wiki/w/Java_Edition_protocol)
