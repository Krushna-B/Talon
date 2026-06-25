package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
)

func main() {
	privPath := flag.String("private", "private.pem", "output path for private key")
	pubPath := flag.String("public", "public.pem", "output path for private key")
	bits := flag.Int("bits", 2048, "RSA key size")
	flag.Parse()

	if err := run(*privPath, *pubPath, *bits); err != nil {
		fmt.Fprintln(os.Stderr, "talon-keygen:", err)
		os.Exit(1)
	}
}

/*
*
Create RSA Private and public key
*
*/
func run(privPath, pubPath string, bits int) error {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("generating rsa key: %w", err)
	}
	privDer, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshall private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDer})

	pubDer, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshall public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDer})
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return fmt.Errorf("writing private key: %w", err)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
		return fmt.Errorf("writing public key: %w", err)
	}

	fmt.Printf("wrote %s (0600) and %s (0644)\n", privPath, pubPath)

	return nil

}
