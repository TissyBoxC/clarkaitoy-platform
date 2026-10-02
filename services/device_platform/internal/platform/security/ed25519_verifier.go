package security

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// Ed25519ProofVerifier validates device signatures using an Ed25519 public key
// encoded as standard or URL-safe base64.
type Ed25519ProofVerifier struct{}

// ValidatePublicKey checks that publicKey is a base64 Ed25519 public key.
func (Ed25519ProofVerifier) ValidatePublicKey(publicKey string) error {
	_, err := decodeDevicePublicKey(publicKey)
	return err
}

// Verify checks one signature over message using publicKey.
func (Ed25519ProofVerifier) Verify(
	publicKey string,
	message []byte,
	signature []byte,
) error {
	decodedKey, err := decodeDevicePublicKey(publicKey)
	if err != nil {
		return err
	}
	if len(signature) != ed25519.SignatureSize {
		return errors.New("device signature has an invalid length")
	}
	if !ed25519.Verify(ed25519.PublicKey(decodedKey), message, signature) {
		return errors.New("device signature verification failed")
	}
	return nil
}

func decodeDevicePublicKey(publicKey string) ([]byte, error) {
	decodedKey, err := decodeBase64(strings.TrimSpace(publicKey))
	if err != nil {
		return nil, fmt.Errorf("decode device public key: %w", err)
	}
	if len(decodedKey) != ed25519.PublicKeySize {
		return nil, errors.New("device public key has an invalid length")
	}
	return decodedKey, nil
}

func decodeBase64(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	return base64.RawURLEncoding.DecodeString(value)
}
