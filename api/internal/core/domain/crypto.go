package domain

type CryptoService interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertextBase64 string) ([]byte, error)
}
