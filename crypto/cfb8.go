package crypto

// inspired by https://github.com/Tnze/go-mc/blob/076f723e3d1467e8bb11fc09dd29e8e92caf339f/net/CFB8/cfb8.go

import "crypto/cipher"

type cfb8 struct {
	block     cipher.Block
	blockSize int
	iv        []byte
	temp      []byte
	decrypt   bool
}

// XORKeyStream satisfies cipher.Stream.
func (c *cfb8) XORKeyStream(dst, src []byte) {
	for i := range src {
		copy(c.temp, c.iv)

		c.block.Encrypt(c.iv, c.iv)
		keystreamByte := c.iv[0]

		outputByte := src[i] ^ keystreamByte
		dst[i] = outputByte
		copy(c.iv, c.temp[1:])

		if c.decrypt {
			c.iv[c.blockSize-1] = src[i]
		} else {
			c.iv[c.blockSize-1] = outputByte
		}
	}
}

func newEncryptStream(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, false)
}

func newDecryptStream(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, true)
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) *cfb8 {
	ivCopy := make([]byte, len(iv))
	copy(ivCopy, iv)

	return &cfb8{
		block:     block,
		blockSize: block.BlockSize(),
		iv:        ivCopy,
		temp:      make([]byte, block.BlockSize()),
		decrypt:   decrypt,
	}
}
