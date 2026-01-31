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

## Not Yet Implemented

These compound types are built on top of primitives and are not yet implemented:

| Type | Description |
| ---- | ----------- |
| Text Component | Chat/display text (JSON-based) |
| Slot | Item stack with data components |

For NBT, we'll probably use `github.com/Tnze/go-mc/nbt` (already a dependency).

## References

- [Minecraft Wiki - Data Types](https://minecraft.wiki/w/Java_Edition_protocol/Data_types)
- [Minecraft Wiki - Protocol](https://minecraft.wiki/w/Java_Edition_protocol)
