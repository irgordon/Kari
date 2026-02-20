// api/internal/api/middleware/auth.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"

	"kari/api/internal/core/domain"
)

// ==============================================================================
// 1. Dependency Injection Struct
// ==============================================================================

type AuthMiddleware struct {
	AuthService domain.AuthService
	RoleService domain.RoleService
	Logger      *slog.Logger
}

func NewAuthMiddleware(authService domain.AuthService, roleService domain.RoleService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService: authService,
		RoleService: roleService,
		Logger:      logger,
	}
}

// ==============================================================================
// 2. Security & Protocol Enforcers
// ==============================================================================

// EnforceTLS ensures that no plaintext traffic can interact with the API.
// It redirects HTTP to HTTPS and injects HSTS headers.
func EnforceTLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the connection is already TLS or if a trusted load balancer forwarded it as HTTPS
		isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

		// Allow localhost bypass for local development DX
		if !isHTTPS && !strings.HasPrefix(r.Host, "localhost:") {
			// Drop the connection and redirect to the secure protocol
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		// Inject HSTS Header (Strict-Transport-Security)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		next.ServeHTTP(w, r)
	})
}

// ==============================================================================
// 3. Identity & Access Management (IAM)
// ==============================================================================

// RequireAuthentication intercepts the HTTP request, extracts the JWT (from cookie or header),
// validates it, and injects the UserClaims into the request context.
func (m *AuthMiddleware) RequireAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// 1. Check for the Secure, HttpOnly cookie first (Browser UI flow)
			if cookie, err := r.Cookie("kari_access_token"); err == nil {
				tokenString = cookie.Value
			} else {
				// 2. Fallback to Authorization: Bearer header (CLI / Programmatic flow)
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				http.Error(w, `{"message": "Unauthorized: Missing token"}`, http.StatusUnauthorized)
				return
			}

			// 3. Validate the token cryptographically
			claims, err := m.AuthService.ValidateAccessToken(r.Context(), tokenString)
			if err != nil {
				m.Logger.Warn("Invalid access token attempt", slog.String("error", err.Error()), slog.String("ip", r.RemoteAddr))
				http.Error(w, `{"message": "Unauthorized: Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// 4. Inject the claims into the request context
			ctx := context.WithValue(r.Context(), domain.UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission intercepts the HTTP request and checks the user's granular rights.
func (m *AuthMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value(domain.UserContextKey).(*domain.UserClaims)
			if !ok {
				http.Error(w, `{"message": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Super Admin Bypass
			if userClaims.RoleName == "Super Admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Check specific permission against the user's assigned role
			hasPerm, err := m.RoleService.RoleHasPermission(r.Context(), userClaims.RoleID, resource, action)
			if err != nil || !hasPerm {
				m.Logger.Warn("Forbidden access attempt", 
					slog.String("user_id", userClaims.Subject.String()),
					slog.String("resource", resource),
					slog.String("action", action),
				)
				http.Error(w, `{"message": "Forbidden: missing permission `+resource+`:`+action+`"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ==============================================================================
// 4. Rate Limiting (In-Memory Token Bucket)
// ==============================================================================

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// getVisitorLimiter retrieves or creates a rate limiter for a specific IP address.
func getVisitorLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// Allow 10 requests per second, with bursts of up to 30
		limiter = rate.NewLimiter(10, 30)
		visitors[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware protects the API from brute-force and DDoS attacks.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the real IP extracted by chi's RealIP middleware
		ip := r.RemoteAddr 
		limiter := getVisitorLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, `{"message": "Too many requests. Please slow down."}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ==============================================================================
// 5. Observability
// ==============================================================================

// StructuredLogger wraps the standard HTTP handler to provide JSON-formatted request logging.
func StructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.String("ip", r.RemoteAddr),
				slog.Duration("latency", time.Since(start)),
				slog.String("trace_id", middleware.GetReqID(r.Context())),
			)
		})
	}
}
