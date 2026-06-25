package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

type Signer struct {
	keyID string
	key   *rsa.PrivateKey
}

/*
*
Extract key from NewSigner
*
*/
func NewSigner(keyID, pemPath string) (*Signer, error) {
	data, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key from %s: %w", pemPath, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", pemPath)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is %T, want *rsa.PrivateKey", parsed)
	}
	return &Signer{keyID: keyID, key: key}, nil
}

func (s *Signer) Sign(timestampMs, method, path string) (string, error) {
	msg := timestampMs + method + path
	digest := sha256.Sum256([]byte(msg))

	sig, err := rsa.SignPSS(rand.Reader, s.key, crypto.SHA256, digest[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	})
	if err != nil {
		return "", fmt.Errorf("signing: %w", err)
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (s *Signer) KeyID() string { return s.keyID }
