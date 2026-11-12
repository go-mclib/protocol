package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type mojangKeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type mojangCertificate struct {
	ExpiresAt            string        `json:"expiresAt"`
	KeyPair              mojangKeyPair `json:"keyPair"`
	PublicKeySignature   string        `json:"publicKeySignature"`
	PublicKeySignatureV2 string        `json:"publicKeySignatureV2"`
	RefreshedAfter       string        `json:"refreshedAfter"`
}

// MojangCertificateData is the decoded chat-signing key pair of a player.
type MojangCertificateData struct {
	PrivateKey     *rsa.PrivateKey
	PublicKeyBytes []byte // SPKI DER format
	SignatureBytes []byte // Mojang signature V2 (512 bytes)
	ExpiryTime     time.Time
}

const mcCertificatesURL = "https://api.minecraftservices.com/player/certificates"

// FetchMojangCertificate retrieves the Mojang signing certificate for chat
// messages, using the client's HTTP client.
func (c *Client) FetchMojangCertificate(ctx context.Context, accessToken string) (*MojangCertificateData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mcCertificatesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	var cert mojangCertificate
	if err := doJSON(c.httpClient, req, "certificate request", &cert); err != nil {
		return nil, err
	}

	privateKey, err := parseRSAPrivateKey(cert.KeyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	publicKeyBytes, err := rsaPublicKeyDER(cert.KeyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	expiryTime, err := time.Parse(time.RFC3339Nano, cert.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate expiry time: %w", err)
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(cert.PublicKeySignatureV2)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Mojang signature: %w", err)
	}

	return &MojangCertificateData{
		PrivateKey:     privateKey,
		PublicKeyBytes: publicKeyBytes,
		SignatureBytes: signatureBytes,
		ExpiryTime:     expiryTime,
	}, nil
}

func parseRSAPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		return rsaPrivateKey, nil
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// rsaPublicKeyDER checks that the PEM holds an RSA public key and returns its DER bytes.
func rsaPublicKeyDER(publicKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	if publicKey, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if _, ok := publicKey.(*rsa.PublicKey); ok {
			return block.Bytes, nil
		}
	}
	if _, err := x509.ParsePKCS1PublicKey(block.Bytes); err != nil {
		return nil, err
	}
	return block.Bytes, nil
}
