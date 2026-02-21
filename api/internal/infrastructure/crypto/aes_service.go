package crypto

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

type AESCryptoService struct {
	key []byte
}

// NewAESCryptoService expects a 64-character hex string (32 bytes of entropy).
func NewAESCryptoService(hexKey string) (*AESCryptoService, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding for key: %w", err)
	}

	if len(key) != 32 {
		return nil, errors.New("encryption key must be exactly 32 bytes (256 bits)")
	}

	return &AESCryptoService{key: key}, nil
}

// Encrypt implements Authenticated Encryption with Associated Data (AEAD).
func (s *AESCryptoService) Encrypt(ctx context.Context, plaintext []byte, associatedData []byte) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 🛡️ Seal(dst, nonce, plaintext, additionalData)
	// The nonce is prepended to the ciphertext for storage.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, associatedData)
	
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt verifies the authenticity of the ciphertext and the associated context.
func (s *AESCryptoService) Decrypt(ctx context.Context, ciphertextBase64 string, associatedData []byte) ([]byte, error) {
	data, err := base64.URLEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("malformed ciphertext: too short")
	}

	nonce, actualCiphertext := data[:nonceSize], data[nonceSize:]

	// 🛡️ Open(dst, nonce, ciphertext, additionalData)
	// If 'associatedData' (e.g. AppID) has been tampered with, this returns an error.
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, associatedData)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: integrity violation or invalid context")
	}

	return plaintext, nil
}
