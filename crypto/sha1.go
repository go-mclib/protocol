package crypto

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"strings"
)

// MinecraftSHA1 creates the SHA1 hash digest of the given Minecraft username, used for auth on server side.
// Original implementation: https://gist.github.com/toqueteos/5372776
func MinecraftSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return minecraftHexDigest(h.Sum(nil))
}

// MinecraftSHA1Builder provides a way to build Minecraft-style SHA1 hashes
type MinecraftSHA1Builder struct {
	hash.Hash
}

// NewMinecraftSHA1 creates a new Minecraft SHA1 builder
func NewMinecraftSHA1() *MinecraftSHA1Builder {
	return &MinecraftSHA1Builder{sha1.New()}
}

// HexDigest returns the Minecraft-style hex digest
func (m *MinecraftSHA1Builder) HexDigest() string {
	return minecraftHexDigest(m.Sum(nil))
}

// negative hashes are two's complemented and prefixed with '-', leading zeroes trimmed
func minecraftHexDigest(hash []byte) string {
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if res == "" {
		res = "0"
	}
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
