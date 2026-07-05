package crypto_test

import (
	"bytes"
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

func newEnabledEncryption(t *testing.T, sharedSecret []byte) *crypto.Encryption {
	t.Helper()
	e := crypto.NewEncryption()
	e.SetSharedSecret(sharedSecret)
	if err := e.EnableEncryption(); err != nil {
		t.Fatalf("EnableEncryption: %v", err)
	}
	return e
}

func TestEncryptionEncrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		secret, _ := hex.DecodeString(tc.sharedSecret)
		plaintext, _ := hex.DecodeString(tc.plaintext)

		e := newEnabledEncryption(t, secret)
		got := hex.EncodeToString(e.Encrypt(plaintext))
		if got != tc.ciphertext {
			t.Errorf("test %d: encryption failed\nexpected: %s\ngot:      %s", i, tc.ciphertext, got)
		}
	}
}

func TestEncryptionDecrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		secret, _ := hex.DecodeString(tc.sharedSecret)
		ciphertext, _ := hex.DecodeString(tc.ciphertext)

		e := newEnabledEncryption(t, secret)
		got := hex.EncodeToString(e.Decrypt(ciphertext))
		if got != tc.plaintext {
			t.Errorf("test %d: decryption failed\nexpected: %s\ngot:      %s", i, tc.plaintext, got)
		}
	}
}

func TestEncryptionChunkedStream(t *testing.T) {
	// CFB8 is a stateful stream: chunked encryption must equal one-shot encryption
	secret, _ := hex.DecodeString(cfb8TestCases[0].sharedSecret)
	plaintext, _ := hex.DecodeString(cfb8TestCases[1].plaintext)

	oneShot := newEnabledEncryption(t, secret).Encrypt(plaintext)

	chunked := newEnabledEncryption(t, secret)
	var got []byte
	for i := 0; i < len(plaintext); i += 7 {
		got = append(got, chunked.Encrypt(plaintext[i:min(i+7, len(plaintext))])...)
	}

	if !bytes.Equal(got, oneShot) {
		t.Errorf("chunked encryption mismatch\nexpected: %x\ngot:      %x", oneShot, got)
	}

	decrypted := newEnabledEncryption(t, secret).Decrypt(got)
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round-trip mismatch\nexpected: %x\ngot:      %x", plaintext, decrypted)
	}
}
