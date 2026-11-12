package wire

import "strings"

// String is a UTF-8 encoded string with a VarInt length prefix (byte count).
//
// The length prefix indicates the number of bytes, not characters.
// Maximum length is 32767 characters (which can be up to ~130KB in UTF-8).
type String string

// Identifier is a namespaced location string.
//
// Format: "namespace:path" where:
//   - namespace: Only lowercase letters, digits, underscores, hyphens, and periods
//   - path: Same as namespace plus forward slashes
//   - If no colon, defaults to "minecraft" namespace
//
// Examples:
//
//	"minecraft:stone"
//	"minecraft:textures/block/stone.png"
//	"custom:my_item"
type Identifier string

// Namespace returns the namespace part of the identifier.
// Returns "minecraft" if no namespace is specified.
func (id Identifier) Namespace() string {
	namespace, _, ok := strings.Cut(string(id), ":")
	if !ok {
		return "minecraft"
	}
	return namespace
}

// Path returns the path part of the identifier.
func (id Identifier) Path() string {
	if _, path, ok := strings.Cut(string(id), ":"); ok {
		return path
	}
	return string(id)
}
