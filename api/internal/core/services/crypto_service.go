package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AESCryptoService provides high-performance AES-256-GCM authenticated encryption.
// 🛡️ Zero-Trust: We store the hardware-accelerated AEAD interface, NOT the raw key bytes.
// 🛡️ SOLID: This struct satisfies domain.CryptoService for dependency injection.
type AESCryptoService struct {
	// cipher.AEAD is inherently thread-safe for concurrent use by multiple Go routines.
	aead cipher.AEAD
}

// NewAESCryptoService initializes the cipher block once during boot.
// It expects a 64-character hexadecimal string representing 32 bytes of raw entropy.
func NewAESCryptoService(hexKey string) (*AESCryptoService, error) {
	// 1. 🛡️ Cryptographic Integrity: Properly decode the hex string into raw bytes
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, errors.New("encryption key must be a valid hexadecimal string")
	}

	// 2. 🛡️ Zero-Trust: Enforce AES-256 (exactly 32 bytes of entropy)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes (got %d bytes)", len(keyBytes))
	}

	// 3. 🛡️ Performance: Initialize the AES cipher block ONCE
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	// 🛡️ Privacy: Best-effort memory hygiene for the decoded key slice
	defer func() {
		for i := range keyBytes {
			keyBytes[i] = 0
		}
	}()

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM instance: %w", err)
	}

	return &AESCryptoService{aead: aesGCM}, nil
}

// Encrypt secures the plaintext with AEAD (Authenticated Encryption with Associated Data).
// 🛡️ Zero-Trust: The associatedData (AAD) cryptographically binds the secret to a context
// (e.g., AppID), preventing cross-resource reuse even if the database is compromised.
func (s *AESCryptoService) Encrypt(ctx context.Context, plaintext []byte, associatedData []byte) (string, error) {
	nonceSize := s.aead.NonceSize()

	// 🛡️ SLA: Exact capacity pre-allocation — Seal appends without reallocation
	buf := make([]byte, nonceSize, nonceSize+len(plaintext)+s.aead.Overhead())

	if _, err := io.ReadFull(rand.Reader, buf[:nonceSize]); err != nil {
		return "", fmt.Errorf("[SLA ERROR] cryptographic nonce generation failed: %w", err)
	}

	// Seal appends the authentication tag to the ciphertext automatically.
	// The AAD is included in the authentication tag but NOT encrypted.
	ciphertext := s.aead.Seal(buf[:nonceSize], buf[:nonceSize], plaintext, associatedData)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt extracts the nonce, verifies the AAD-bound authentication tag, and decrypts.
// 🛡️ Zero-Trust: If the AAD does not match what was used during encryption, the
// authentication tag verification fails and this method returns an error immediately.
func (s *AESCryptoService) Decrypt(ctx context.Context, ciphertextBase64 string, associatedData []byte) ([]byte, error) {
	enc, err := base64.URLEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, errors.New("[SLA ERROR] failed to decode base64 ciphertext")
	}

	nonceSize := s.aead.NonceSize()
	if len(enc) < nonceSize {
		return nil, errors.New("[SLA ERROR] ciphertext too short: missing nonce")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	// 🛡️ Zero-Trust: Open cryptographically verifies the authentication tag BEFORE decrypting.
	// If the database was tampered with, or if the AAD context doesn't match, this fails.
	plaintext, err := s.aead.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		return nil, errors.New("decryption failed: integrity violation - potential tampering detected")
	}

	return plaintext, nil
}
