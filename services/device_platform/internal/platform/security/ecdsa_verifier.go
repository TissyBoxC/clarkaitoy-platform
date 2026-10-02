package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const ecdsaP256SignatureSize = 64

// ECDSAProofVerifier validates device signatures using an uncompressed P-256
// public key encoded as standard or URL-safe base64.
type ECDSAProofVerifier struct{}

// ValidatePublicKey checks that publicKey is an uncompressed P-256 key.
func (ECDSAProofVerifier) ValidatePublicKey(publicKey string) error {
	_, err := decodeECDSAPublicKey(publicKey)
	return err
}

// Verify checks one r || s signature over SHA-256(message).
func (ECDSAProofVerifier) Verify(
	publicKey string,
	message []byte,
	signature []byte,
) error {
	decodedKey, err := decodeECDSAPublicKey(publicKey)
	if err != nil {
		return err
	}
	if len(signature) != ecdsaP256SignatureSize {
		return errors.New("device signature has an invalid length")
	}
	digest := sha256.Sum256(message)
	signatureR := new(big.Int).SetBytes(signature[:32])
	signatureS := new(big.Int).SetBytes(signature[32:])
	if !ecdsa.Verify(decodedKey, digest[:], signatureR, signatureS) {
		return errors.New("device signature verification failed")
	}
	return nil
}

func decodeECDSAPublicKey(publicKey string) (*ecdsa.PublicKey, error) {
	decodedKey, err := decodeBase64(strings.TrimSpace(publicKey))
	if err != nil {
		return nil, fmt.Errorf("decode device public key: %w", err)
	}
	if len(decodedKey) != 65 || decodedKey[0] != 4 {
		return nil, errors.New("device public key has an invalid length")
	}
	curve := elliptic.P256()
	x := new(big.Int).SetBytes(decodedKey[1:33])
	y := new(big.Int).SetBytes(decodedKey[33:])
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("device public key is not on P-256")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func decodeBase64(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	return base64.RawURLEncoding.DecodeString(value)
}
