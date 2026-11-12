package crypto

import (
	"cmp"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// MinecraftHexDigest returns Minecraft's signed hex SHA-1 digest of the
// concatenated parts, as used for the session server hash.
// Original implementation: https://gist.github.com/toqueteos/5372776
func MinecraftHexDigest(parts ...[]byte) string {
	h := sha1.New()
	for _, p := range parts {
		h.Write(p)
	}
	return minecraftHexDigest(h.Sum(nil))
}

// negative hashes are two's complemented and prefixed with '-', leading zeroes trimmed
func minecraftHexDigest(hash []byte) string {
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}
	res := cmp.Or(strings.TrimLeft(hex.EncodeToString(hash), "0"), "0")
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
