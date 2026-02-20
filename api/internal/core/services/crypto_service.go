package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type AESCryptoService struct {
	key []byte
}

func NewAESCryptoService(hexKey string) (*AESCryptoService, error) {
	if len(hexKey) != 32 {
		return nil, errors.New("encryption key must be exactly 32 bytes")
	}
	return &AESCryptoService{key: []byte(hexKey)}, nil
}

func (s *AESCryptoService) Encrypt(plaintext []byte) (string, error) {
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

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt extracts the nonce and decrypts the ciphertext
func (s *AESCryptoService) Decrypt(ciphertextBase64 string) ([]byte, error) {
	enc, err := base64.StdEncoding.DecodeString(ciphertextBase64)
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
	if len(enc) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
