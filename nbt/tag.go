// Package nbt implements Minecraft's Named Binary Tag format.
//
// NBT is a tree data structure used by Minecraft for save files and network
// transmission. This package supports both file format (with root tag name)
// and network format (nameless root tag).
//
// Basic usage with struct tags (similar to encoding/json):
//
//	type Player struct {
//	    Name  string  `nbt:"name"`
//	    X     float64 `nbt:"x"`
//	    Y     float64 `nbt:"y"`
//	    Z     float64 `nbt:"z"`
//	    Items []Item  `nbt:"items"`
//	}
//
//	// Marshal to NBT
//	data, err := nbt.Marshal(player)
//
//	// Unmarshal from NBT
//	var player Player
//	err := nbt.Unmarshal(data, &player)
//
// For network transmission, use MarshalNetwork/UnmarshalNetwork which omit
// the root tag name.
package nbt

// Tag type IDs as defined by the NBT specification.
const (
	TagEnd       byte = 0
	TagByte      byte = 1
	TagShort     byte = 2
	TagInt       byte = 3
	TagLong      byte = 4
	TagFloat     byte = 5
	TagDouble    byte = 6
	TagByteArray byte = 7
	TagString    byte = 8
	TagList      byte = 9
	TagCompound  byte = 10
	TagIntArray  byte = 11
	TagLongArray byte = 12
)

// TagName returns a human-readable name for a tag type ID.
func TagName(id byte) string {
	switch id {
	case TagEnd:
		return "End"
	case TagByte:
		return "Byte"
	case TagShort:
		return "Short"
	case TagInt:
		return "Int"
	case TagLong:
		return "Long"
	case TagFloat:
		return "Float"
	case TagDouble:
		return "Double"
	case TagByteArray:
		return "ByteArray"
	case TagString:
		return "String"
	case TagList:
		return "List"
	case TagCompound:
		return "Compound"
	case TagIntArray:
		return "IntArray"
	case TagLongArray:
		return "LongArray"
	default:
		return "Unknown"
	}
}

// Tag is the interface implemented by all NBT tag types.
type Tag interface {
	// ID returns the tag type identifier.
	ID() byte

	// write encodes the tag payload (not including type ID or name).
	write(w *Writer) error
}

// MaxDepth is the default maximum nesting depth for NBT structures.
// This matches Minecraft's limit of 512.
const MaxDepth = 512

// MaxBytes is the default maximum bytes that can be read.
// Set to 0 for unlimited.
const MaxBytes int64 = 2 * 1024 * 1024 // 2 MB default
