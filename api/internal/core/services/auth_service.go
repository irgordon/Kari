package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"kari/api/internal/config"
	"kari/api/internal/core/domain"
)

type AuthService struct {
	repo   domain.UserRepository
	config *config.Config
}

// KariClaims extends standard JWT claims with Kari-specific metadata
type KariClaims struct {
	Email string `json:"email"`
	Rank  int    `json:"rank"`
	jwt.RegisteredClaims
}

func NewAuthService(repo domain.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo:   repo,
		config: cfg,
	}
}

// GenerateTokenPair issues a short-lived Access Token and a long-lived Refresh Token
func (s *AuthService) GenerateTokenPair(user *domain.User) (string, string, error) {
	// 1. Access Token (15-30 Minutes) - Used for gRPC/REST authorization
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (7 Days) - Stored in DB to allow remote revocation
	refreshToken := uuid.New().String()
	err = s.repo.UpdateRefreshToken(context.Background(), user.ID, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	claims := KariClaims{
		Email: user.Email,
		Rank:  user.Role.Rank,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kari-brain",
			Audience:  []string{"kari-ui", "kari-agent"},
		},
	}

	// 🛡️ SLA: HS256 is used for symmetric signing via the platform-injected KARI_JWT_SECRET
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JwtSecret))
}

// Login validates credentials and initiates the token lifecycle
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials") // 🛡️ Zero-Trust: Generic error
	}

	// 🛡️ Platform Agnostic: Bcrypt is used for constant-time password comparison
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		return "", "", errors.New("account suspended")
	}

	return s.GenerateTokenPair(user)
}
