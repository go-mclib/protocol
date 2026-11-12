package crypto_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"github.com/go-mclib/protocol/crypto"
)

// AES-128-CFB8 vectors with the shared secret used as both key and IV,
// as in the Minecraft protocol; ciphertexts computed with openssl enc -aes-128-cfb8
var cfb8TestCases = []struct {
	sharedSecret, plaintext, ciphertext string
}{
	{
		"000102030405060708090a0b0c0d0e0f",
		"6bc1bee22e409f96e93d7e117393172a",
		"614622e1d1e39ca6e652427481aab2ca",
	},
	{
		"000102030405060708090a0b0c0d0e0f",
		"0ecbd6d36cd12962ce671b4d96fb95aaa902096aeac366e13a6ae57c05d48673cf320c626689d05548f65fd6a108630c1d4e3aab543b006823c7a9422e97c0431587537c384f99a11488ffd9b2e9b46f49005a7e5cef64e27e2de3cf3fb87c1524766601",
		"046a019612b5d7624a498b89d05d5074b0835025e28a12a0455a99c958e5ee4c6a954619ac3bee65d514af339fdf1467a5e16f06612433a8ff2307fab66ad8be251280eaec482b8de94c15785900c2ecdf269c4bac0bf96de0bfafbf45ad27a4e60d2cd9",
	},
}

func newStream(t *testing.T, sharedSecret []byte, decrypt bool) cipher.Stream {
	t.Helper()
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	return crypto.NewCFB8(block, sharedSecret, decrypt)
}

func TestCFB8Encrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		secret, _ := hex.DecodeString(tc.sharedSecret)
		plaintext, _ := hex.DecodeString(tc.plaintext)

		encrypted := make([]byte, len(plaintext))
		newStream(t, secret, false).XORKeyStream(encrypted, plaintext)
		got := hex.EncodeToString(encrypted)
		if got != tc.ciphertext {
			t.Errorf("test %d: encryption failed\nexpected: %s\ngot:      %s", i, tc.ciphertext, got)
		}
	}
}

func TestCFB8Decrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		secret, _ := hex.DecodeString(tc.sharedSecret)
		ciphertext, _ := hex.DecodeString(tc.ciphertext)

		// decrypt in place: dst and src fully overlap
		newStream(t, secret, true).XORKeyStream(ciphertext, ciphertext)
		got := hex.EncodeToString(ciphertext)
		if got != tc.plaintext {
			t.Errorf("test %d: decryption failed\nexpected: %s\ngot:      %s", i, tc.plaintext, got)
		}
	}
}

func TestCFB8ChunkedStream(t *testing.T) {
	// CFB8 is a stateful stream: chunked encryption must equal one-shot encryption
	secret, _ := hex.DecodeString(cfb8TestCases[0].sharedSecret)
	plaintext, _ := hex.DecodeString(cfb8TestCases[1].plaintext)

	oneShot := make([]byte, len(plaintext))
	newStream(t, secret, false).XORKeyStream(oneShot, plaintext)

	chunked := newStream(t, secret, false)
	got := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += 7 {
		end := min(i+7, len(plaintext))
		chunked.XORKeyStream(got[i:end], plaintext[i:end])
	}

	if !bytes.Equal(got, oneShot) {
		t.Errorf("chunked encryption mismatch\nexpected: %x\ngot:      %x", oneShot, got)
	}

	decrypted := make([]byte, len(got))
	newStream(t, secret, true).XORKeyStream(decrypted, got)
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round-trip mismatch\nexpected: %x\ngot:      %x", plaintext, decrypted)
	}
}

func TestNewSharedSecret(t *testing.T) {
	a, b := crypto.NewSharedSecret(), crypto.NewSharedSecret()
	if len(a) != 16 || len(b) != 16 || bytes.Equal(a, b) {
		t.Fatalf("shared secrets %x %x", a, b)
	}
}
