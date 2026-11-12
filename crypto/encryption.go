package crypto

// https://minecraft.wiki/w/Protocol_encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
)

// NewSharedSecret returns a random 16-byte shared secret for the login
// encryption handshake; it is both the AES-128 key and the CFB8 IV.
func NewSharedSecret() []byte {
	secret := make([]byte, 16)
	_, _ = rand.Read(secret) // never fails since Go 1.24
	return secret
}

// EncryptWithPublicKey RSA-encrypts data with a DER-encoded public key,
// as used in the login encryption handshake.
func EncryptWithPublicKey(publicKeyBytes []byte, data []byte) ([]byte, error) {
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encrypted, nil
}

// DecryptWithPrivateKey RSA-decrypts login handshake data with the server's key.
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return decrypted, nil
}
