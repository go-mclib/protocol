package crypto

// inspired by https://github.com/Tnze/go-mc/blob/076f723e3d1467e8bb11fc09dd29e8e92caf339f/net/CFB8/cfb8.go

import (
	"bytes"
	"crypto/cipher"
)

type cfb8 struct {
	block   cipher.Block
	iv      []byte
	temp    []byte
	decrypt bool
}

// XORKeyStream satisfies cipher.Stream; dst and src may overlap entirely.
func (c *cfb8) XORKeyStream(dst, src []byte) {
	last := len(c.iv) - 1
	for i := range src {
		in := src[i]
		copy(c.temp, c.iv)
		c.block.Encrypt(c.iv, c.iv)
		out := in ^ c.iv[0]
		dst[i] = out

		copy(c.iv, c.temp[1:])
		if c.decrypt {
			c.iv[last] = in
		} else {
			c.iv[last] = out
		}
	}
}

// NewCFB8 returns an AES/CFB8 stream as used by the Minecraft protocol. The
// protocol uses the shared secret as both the AES key and iv; iv is copied.
func NewCFB8(block cipher.Block, iv []byte, decrypt bool) cipher.Stream {
	return &cfb8{
		block:   block,
		iv:      bytes.Clone(iv),
		temp:    make([]byte, block.BlockSize()),
		decrypt: decrypt,
	}
}
