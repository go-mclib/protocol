package crypto_test

import (
	"testing"

	mc_crypto "github.com/go-mclib/protocol/crypto"
)

func TestOfflinePlayerUUID(t *testing.T) {
	for name, want := range map[string]string{
		"Notch": "b50ad385-829d-3141-a216-7e7d7539ba7f",
		"Steve": "5627dd98-e6be-3c21-b8a8-e92344183641",
	} {
		if got := mc_crypto.OfflinePlayerUUID(name).String(); got != want {
			t.Fatalf("OfflinePlayerUUID(%q) = %s, want %s", name, got, want)
		}
	}
}
