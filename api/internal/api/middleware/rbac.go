package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"kari/api/internal/core/domain"
)

type contextKey string

const (
	UserKey  contextKey = "user_id"
	RoleKey  contextKey = "role_rank"
)

type RBACMiddleware struct {
	repo      domain.UserRepository
	jwtSecret []byte
}

func NewRBACMiddleware(repo domain.UserRepository, secret string) *RBACMiddleware {
	return &RBACMiddleware{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

// Authenticate verifies the JWT and injects the UserID into the context
func (m *RBACMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Token from Authorization Header or HttpOnly Cookie
		authHeader := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenStr == "" {
			// Fallback to cookie for SvelteKit SSR requests
			if cookie, err := r.Cookie("access_token"); err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// 2. Parse and Validate JWT
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			http.Error(w, "Malformed token subject", http.StatusUnauthorized)
			return
		}

		// 🛡️ 3. Real-time Security Check (Database Verification)
		// We verify the user is still 'active' in the DB to prevent 
		// "Ghost Access" from revoked accounts with long-lived tokens.
		user, err := m.repo.GetByID(r.Context(), userID)
		if err != nil || !user.IsActive {
			http.Error(w, "User account is suspended or removed", http.StatusForbidden)
			return
		}

		// 4. Inject into Context
		ctx := context.WithValue(r.Context(), UserKey, user.ID)
		ctx = context.WithValue(ctx, RoleKey, user.Role.Rank)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission ensures the user's role has the specific right for the resource
func (m *RBACMiddleware) RequirePermission(resource string, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserKey).(uuid.UUID)

			// 🛡️ 5. SLA Enforcement: Consult the Dynamic RBAC Store
			// We check if the role assigned to this user in the DB has the required permission.
			hasPerm, err := m.repo.HasPermission(r.Context(), userID, resource, action)
			if err != nil || !hasPerm {
				http.Error(w, "Insufficient permissions for this action", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}
