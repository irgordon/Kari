package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"kari/api/internal/core/domain"
)

// Dummy hash to equalize timing attacks. This is a valid bcrypt hash of the word "dummy".
var dummyBcryptHash = []byte("$2a$10$wTf/0J/Q32r.5R7bU4X8uO4b2pE7Z9H5a0rY4q1w4s7c9d0x2z5eG")

// AuthService orchestrates secure login flows and session generation.
type AuthService struct {
	repo         domain.UserRepository
	tokenService *TokenService // 🛡️ SOLID: Inject the cryptographic engine
}

// NewAuthService creates a new authentication orchestrator.
func NewAuthService(repo domain.UserRepository, ts *TokenService) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: ts,
	}
}

// Login authenticates a user safely against timing and enumeration attacks.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// 1. 🛡️ Zero-Trust: Anti-Enumeration
		// Even if the user doesn't exist, we force the CPU to compute a bcrypt hash.
		// This guarantees the HTTP response takes ~100ms regardless of user existence.
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(password))
		return "", "", errors.New("invalid credentials")
	}

	// 2. Constant-time credential check
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		// 🛡️ Information Obfuscation: Do not tell the attacker the account is suspended.
		return "", "", errors.New("invalid credentials")
	}

	return s.GenerateTokenPair(ctx, user)
}

// GenerateTokenPair mints a stateless Access Token and a stateful, hashed Opaque Refresh Token.
func (s *AuthService) GenerateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	// 1. 🛡️ SOLID: Delegate stateless JWT minting to the TokenService
	// (Assuming we refactored TokenService to just output the access token string)
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// 2. 🛡️ Secure Opaque Refresh Token Generation
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate cryptographic entropy: %w", err)
	}
	
	// This is the raw string sent to the SvelteKit edge (and stored in the HttpOnly cookie)
	refreshTokenPlain := base64.URLEncoding.EncodeToString(b)

	// 3. 🛡️ Zero-Trust Storage: Hash before persistence
	// We use SHA-256 to hash the refresh token. Because refresh tokens are 32 bytes 
	// of raw entropy, they are mathematically immune to rainbow table attacks, 
	// so a fast hashing algorithm like SHA-256 (instead of bcrypt) is safe and performant.
	hash := sha256.Sum256([]byte(refreshTokenPlain))
	refreshTokenHash := hex.EncodeToString(hash[:])

	// We store the HASH in PostgreSQL, never the plaintext token.
	err = s.repo.UpdateRefreshToken(ctx, user.ID, refreshTokenHash)
	if err != nil {
		return "", "", fmt.Errorf("failed to persist refresh token hash: %w", err)
	}

	// We return the plaintext token to the handler so it can be sent to the user.
	return accessToken, refreshTokenPlain, nil
}
