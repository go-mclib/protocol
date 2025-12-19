# Java ↔ Schema ↔ Go Type Mapping Table

This document defines the complete type mapping for the Minecraft protocol packet definitions.

## Version
- Minecraft: 1.21.11
- Protocol: 774
- Last Updated: 2025-12-19

---

## Primitive Types

| Java Method | ByteBufCodecs Constant | Schema Type | Go Type | Notes |
|------------|----------------------|-------------|---------|-------|
| `readBoolean()` | `BOOL` | `Boolean` | `ns.Boolean` | Single byte: 0x00 or 0x01 |
| `readByte()` | `BYTE` | `Byte` | `ns.Byte` | Signed 8-bit int (-128 to 127) |
| `readUnsignedByte()` | - | `UnsignedByte` | `ns.UnsignedByte` | Unsigned 8-bit int (0 to 255) |
| `readShort()` | `SHORT` | `Short` | `ns.Short` | Signed 16-bit int, big-endian |
| `readUnsignedShort()` | `UNSIGNED_SHORT` | `UnsignedShort` | `ns.UnsignedShort` | Unsigned 16-bit int, big-endian |
| `readInt()` | `INT` | `Int` | `ns.Int` | Signed 32-bit int, big-endian |
| `readLong()` | `LONG` | `Long` | `ns.Long` | Signed 64-bit int, big-endian |
| `readFloat()` | `FLOAT` | `Float` | `ns.Float` | 32-bit IEEE 754 float |
| `readDouble()` | `DOUBLE` | `Double` | `ns.Double` | 64-bit IEEE 754 double |
| `readVarInt()` | `VAR_INT` | `VarInt` | `ns.VarInt` | Variable-length signed 32-bit int |
| `readVarLong()` | `VAR_LONG` | `VarLong` | `ns.VarLong` | Variable-length signed 64-bit int |

## String Types

| Java Method | ByteBufCodecs Constant | Schema Type | Go Type | Notes |
|------------|----------------------|-------------|---------|-------|
| `readUtf()` | `STRING_UTF8` | `String` | `ns.String` | VarInt-prefixed UTF-8 string |
| `readUtf(maxLen)` | `stringUtf8(n)` | `String(maxLen)` | `ns.String` | Max length validated |
| `readIdentifier()` | - | `Identifier` | `ns.Identifier` | Namespaced ID (e.g., "minecraft:stone") |

## Specialized Primitives

| Java Method | ByteBufCodecs Constant | Schema Type | Go Type | Notes |
|------------|----------------------|-------------|---------|-------|
| - | `ROTATION_BYTE` | `Angle` | `ns.Angle` | Rotation in 1/256ths of full turn |
| `readUUID()` | - | `UUID` | `ns.UUID` | 128-bit UUID (16 bytes) |

## Position & Vector Types

| Java Method | ByteBufCodecs Constant | Schema Type | Go Type | Status |
|------------|----------------------|-------------|---------|--------|
| `readBlockPos()` | - | `Position` | `ns.Position` | ✅ Exists |
| `readChunkPos()` | - | `ChunkPos` | `ns.ChunkPos` | ⚠️ TODO: Add to net_structures |
| `readGlobalPos()` | - | `GlobalPos` | `ns.GlobalPos` | ⚠️ TODO: Add to net_structures |
| `readVec3()` | - | `Vec3` | `ns.Vec3` | ⚠️ TODO: Add to net_structures |
| `readLpVec3()` | - | `LpVec3` | `ns.LpVec3` | ⚠️ TODO: Add (length-prefixed Vec3) |
| `readVector3f()` | `VECTOR3F` | `Vector3f` | `ns.Vector3f` | ⚠️ TODO: Add to net_structures |
| `readQuaternion()` | `QUATERNIONF` | `Quaternionf` | `ns.Quaternionf` | ⚠️ TODO: Add to net_structures |

## Array Types

| Java Method | Schema Pattern | Go Type | Notes |
|------------|---------------|---------|-------|
| `readByteArray()` | `ByteArray` | `ns.ByteArray` | VarInt-prefixed byte array |
| `readByteArray(maxLen)` | `ByteArray(maxLen)` | `ns.ByteArray` | With max length |
| `readVarIntArray()` | `VarIntArray` | `ns.PrefixedArray[VarInt]` | VarInt count + elements |
| `readLongArray()` | `LongArray` | `ns.PrefixedArray[Long]` | VarInt count + elements |
| `readFixedSizeLongArray()` | `FixedLongArray(size)` | `ns.Array[Long]` | Fixed size, no prefix |
| `readIntIdList()` | `IntIdList` | `ns.PrefixedArray[VarInt]` | Special case of VarIntArray |

## Generic Collection Types

| Java Method | Schema Pattern | Go Type | Notes |
|------------|---------------|---------|-------|
| `readList(decoder)` | `{"array": "prefixed", "element": T}` | `ns.PrefixedArray[T]` | VarInt count prefix |
| `readCollection(factory, decoder)` | `{"array": "prefixed", "element": T}` | `ns.PrefixedArray[T]` | Same as list |
| `readMap(keyDec, valDec)` | `{"map": {"key": K, "value": V}}` | ⚠️ TODO: `ns.PrefixedMap[K,V]` | VarInt count prefix |

## Optional & Conditional Types

| Java Method | Schema Pattern | Go Type | Notes |
|------------|---------------|---------|-------|
| `readOptional(decoder)` | `{"optional": T}` | `ns.PrefixedOptional[T]` | Boolean prefix + optional value |
| `readNullable(decoder)` | `{"nullable": T}` | `ns.PrefixedOptional[T]` | Same as optional |
| - | `{"conditional": T, "when": "field"}` | `ns.Optional[T]` | No boolean prefix, condition from another field |

## Enum Types

| Java Method | Schema Pattern | Go Type | Notes |
|------------|---------------|---------|-------|
| `readEnum(Class)` | `{"enum": "EnumName"}` | `ns.VarInt` | VarInt with ordinal value |
| `readById(intFunc)` | `{"enum": "EnumName"}` | `ns.VarInt` | Custom int→value mapping |
| `readEnumSet(Class)` | `{"enumSet": "EnumName"}` | ⚠️ TODO: `ns.EnumSet` | BitSet representation |

## Variant/Union Types (Either)

| Java Method | Schema Pattern | Go Type | Notes |
|------------|---------------|---------|-------|
| `readEither(leftDec, rightDec)` | `{"either": {"left": L, "right": R}}` | `ns.Or[L, R]` | Boolean discriminator |
| - | `{"variant": {...}}` | ⚠️ TODO: `ns.Variant[K,V]` | Enum discriminator + data |

## NBT & Structured Data

| Java Method | ByteBufCodecs Constant | Schema Type | Go Type | Notes |
|------------|----------------------|-------------|---------|-------|
| `readNbt()` | `COMPOUND_TAG` | `NBT` | `ns.NBT` | ✅ Exists |
| - | `TRUSTED_COMPOUND_TAG` | `NBT(trusted)` | `ns.NBT` | No size limits |
| - | `TAG` | `Tag` | `ns.NBT` | Any NBT tag type |
| `readWithCodec(ops, codec)` | - | `ByteArray` | `ns.ByteArray` | ⚠️ Fallback for complex codecs |

## Minecraft-Specific Types

| Java Method | Schema Type | Go Type | Status | Notes |
|------------|-------------|---------|--------|-------|
| - | `TextComponent` | `ns.TextComponent` | ✅ Exists | Chat/text formatting |
| - | `JSONTextComponent` | `ns.JSONTextComponent` | ✅ Exists | Legacy JSON format |
| - | `GameProfile` | `ns.GameProfile` | ✅ Exists | Player profile data |
| - | `BitSet` | `ns.BitSet` | ✅ Exists | Variable-length bitset |
| - | `FixedBitSet` | `ns.FixedBitSet` | ⚠️ TODO | Fixed-size bitset |
| - | `Slot` | `ns.Slot` | ⚠️ Use ByteArray | Inventory slot (complex) |
| - | `HashedSlot` | `ns.HashedSlot` | ⚠️ Use ByteArray | Slot with hash (complex) |
| - | `EntityMetadata` | `ns.EntityMetadata` | ⚠️ Use ByteArray | Entity metadata (complex) |
| - | `Particle` | - | ⚠️ Use ByteArray | Particle data (80+ variants) |

## Registry Types

| Schema Pattern | Go Type | Status | Notes |
|---------------|---------|--------|-------|
| `{"registry": "entity_type"}` | `ns.Registry[EntityType]` | ⚠️ TODO | Registry-based ID |
| `{"registry": "block"}` | `ns.Registry[Block]` | ⚠️ TODO | Block registry ID |
| `{"registry": "item"}` | `ns.Registry[Item]` | ⚠️ TODO | Item registry ID |
| `{"idOr": T}` | `ns.IDor[T]` | ✅ Exists | Registry ID or inline data |
| `IDSet` | `ns.IDSet` | ✅ Exists | Set of registry IDs |

## Complex/Composite Types

| Schema Pattern | Go Type | Notes |
|---------------|---------|-------|
| `{"composite": {"fields": [...]}}` | Anonymous struct | Inline struct definition |
| `{"ref": "TypeName"}` | Named type | Reference to named type in schema |
| `SoundEvent` | `ns.SoundEvent` | ✅ Exists |
| `ChatType` | `ns.ChatType` | ✅ Exists |
| `ChunkData` | `ns.ChunkData` | ✅ Exists |
| `LightData` | `ns.LightData` | ✅ Exists |

## Special/Fallback Types

| Schema Type | Go Type | When to Use |
|------------|---------|-------------|
| `ByteArray` | `ns.ByteArray` | Unknown/complex structures |
| `RestOfPacket` | `ns.ByteArray` | Consume all remaining bytes |
| - | - | See blacklist configuration |

---

## Blacklist Configuration

Types that should fall back to `ByteArray` for now:

### By Type Name
- `EntityMetadata` - Complex, version-specific structure
- `Particle` - 80+ discriminated variants
- `SlotDisplay` - Recipe display data (complex)
- `RecipeDisplay` - Recipe description (complex)

### By Regex Pattern
- `.*Recipe.*Display` - All recipe display types
- `.*Commands.*` - Command system packets (very complex)

### Manual Packets
Packets with custom codec logic (not using StreamCodec.composite):
- TBD: Will identify during extraction phase

---

## Missing Types to Implement

Priority order for adding to `net_structures`:

### High Priority (needed for basic packets)
1. ✅ `ChunkPos` - Used in many chunk-related packets
2. ✅ `GlobalPos` - Used in respawn/death location
3. ✅ `Vec3` - Common in entity/movement packets
4. ✅ `LpVec3` - Length-prefixed variant
5. ✅ `Vector3f` - Float precision vector
6. ✅ `Quaternionf` - Rotation quaternion
7. ✅ `FixedBitSet` - Fixed-size bitset variant
8. ✅ `PrefixedMap[K,V]` - Generic map type

### Medium Priority
9. ⚠️ `EnumSet` - Used in various flag packets
10. ⚠️ `Variant[K,V]` - Discriminated union type
11. ⚠️ `Registry[T]` - Typed registry wrapper

### Low Priority (can use ByteArray initially)
12. Particle codecs
13. Entity metadata codecs
14. Recipe display codecs

---

## Schema JSON Examples

### Simple Packet
```json
{
  "name": "entityId",
  "type": "VarInt"
}
```

### Prefixed Array
```json
{
  "name": "dimensions",
  "type": {
    "array": "prefixed",
    "element": "Identifier"
  }
}
```

### Optional with Condition
```json
{
  "name": "deathLocation",
  "type": {
    "conditional": {
      "fields": [
        {"name": "dimension", "type": "Identifier"},
        {"name": "position", "type": "Position"}
      ]
    },
    "when": "hasDeathLocation"
  }
}
```

### Variant (Discriminated Union)
```json
{
  "name": "action",
  "type": {
    "variant": {
      "discriminator": "VarInt",
      "variants": {
        "0": {"name": "attack", "fields": []},
        "1": {"name": "interact", "fields": [
          {"name": "hand", "type": "VarInt"}
        ]},
        "2": {"name": "interact_at", "fields": [
          {"name": "location", "type": "Vec3"},
          {"name": "hand", "type": "VarInt"}
        ]}
      }
    }
  }
}
```

### Registry Type
```json
{
  "name": "entityType",
  "type": {
    "registry": "entity_type"
  }
}
```

---

## Next Steps

1. ✅ Create this mapping table
2. ⚠️ Implement missing types in `net_structures/` (ChunkPos, Vec3, etc.)
3. ⚠️ Define complete schema JSON format spec
4. ⚠️ Build Java parser to extract StreamCodec definitions
5. ⚠️ Generate schema JSON from decompiled source
6. ⚠️ Build Go code generator from schemas
7. ⚠️ Test with simple packets first (handshake, status)
