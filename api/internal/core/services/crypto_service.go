package services

import (
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
// 🛡️ Zero-Trust: We store the hardware-accelerated cipher block, NOT the raw key bytes.
type AESCryptoService struct {
	// cipher.AEAD is inherently thread-safe for concurrent use by multiple Go routines.
	gcm cipher.AEAD
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

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM instance: %w", err)
	}

	return &AESCryptoService{gcm: aesGCM}, nil
}

// Encrypt secures the plaintext and attaches a cryptographic authentication tag.
func (s *AESCryptoService) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate cryptographic nonce: %w", err)
	}

	// Seal appends the authentication tag to the ciphertext automatically.
	// We prepend the nonce to the result so it can be extracted during decryption.
	ciphertext := s.gcm.Seal(nonce, nonce, plaintext, nil)
	
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt extracts the nonce, verifies the authentication tag, and decrypts the ciphertext.
func (s *AESCryptoService) Decrypt(ciphertextBase64 string) ([]byte, error) {
	enc, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, errors.New("failed to decode base64 ciphertext")
	}

	nonceSize := s.gcm.NonceSize()
	if len(enc) < nonceSize {
		return nil, errors.New("ciphertext too short: missing nonce")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	
	// 🛡️ Zero-Trust: Open cryptographically verifies the authentication tag BEFORE decrypting.
	// If the database was tampered with, this will throw an error and refuse to return corrupted data.
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid key or tampered ciphertext")
	}

	return plaintext, nil
}
