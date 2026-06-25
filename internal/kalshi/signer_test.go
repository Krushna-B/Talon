package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestSignerRoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	signer := &Signer{keyID: "test-key", key: key}

	cases := []struct {
		name        string
		timestampMs string
		method      string
		path        string
	}{
		{"ws handshake", "1700000000000", "GET", "/trade-api/ws/v2"},
		{"rest markets", "1700000000001", "GET", "/trade-api/v2/markets"},
		{"empty path", "1700000000002", "GET", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sig, err := signer.Sign(tc.timestampMs, tc.method, tc.path)
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}

			rawSig, err := base64.StdEncoding.DecodeString(sig)
			if err != nil {
				t.Fatalf("decoding signature: %v", err)
			}

			digest := sha256.Sum256([]byte(tc.timestampMs + tc.method + tc.path))
			err = rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, digest[:], rawSig, &rsa.PSSOptions{
				SaltLength: rsa.PSSSaltLengthEqualsHash,
				Hash:       crypto.SHA256,
			})
			if err != nil {
				t.Errorf("signature failed verification: %v", err)
			}
		})
	}
}
