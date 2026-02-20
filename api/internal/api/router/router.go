// api/internal/api/router/router.go
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"kari/api/internal/api/handlers"
	auth_middleware "kari/api/internal/api/middleware"
)

// RouterConfig defines the strict dependencies required to build the API routing tree.
type RouterConfig struct {
	AuthHandler    *handlers.AuthHandler
	AppHandler     *handlers.AppHandler
	DomainHandler  *handlers.DomainHandler
	AuditHandler   *handlers.AuditHandler
	WSHandler      *handlers.WebSocketHandler
	AuthMiddleware *auth_middleware.AuthMiddleware
	Logger         *slog.Logger
}

// NewRouter constructs the Chi multiplexer, attaches global middleware, and wires all endpoints.
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// =========================================================================
	// 1. Global Gateway Middleware Pipeline
	// =========================================================================

	// Injects a unique trace_id into every request context for logging and audit trails
	r.Use(middleware.RequestID)
	
	// Extracts the true client IP (respecting X-Forwarded-For if behind a load balancer)
	r.Use(middleware.RealIP)
	
	// Structured JSON logging for every HTTP request
	r.Use(auth_middleware.StructuredLogger(cfg.Logger))
	
	// Catches panic() in any handler and returns a 500 error instead of crashing the Go daemon
	r.Use(middleware.Recoverer)
	
	// Failsafe: No HTTP request is allowed to hang for more than 60 seconds
	r.Use(middleware.Timeout(60 * time.Second))

	// In-memory token bucket rate limiting
	r.Use(auth_middleware.RateLimitMiddleware)

	// 🔒 Force all connections to use TLS/SSL and inject HSTS headers
	r.Use(auth_middleware.EnforceTLS)

	// Strict CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300, 
	}))

	// =========================================================================
	// 2. API v1 Routing Tree
	// =========================================================================

	r.Route("/api/v1", func(r chi.Router) {

		// ---------------------------------------------------------------------
		// Public Routes (No Auth Required)
		// ---------------------------------------------------------------------
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", cfg.AuthHandler.Login)
			r.Post("/auth/refresh", cfg.AuthHandler.Refresh)
			r.Post("/webhooks/github", cfg.AppHandler.HandleGitHubWebhook)
		})

		// ---------------------------------------------------------------------
		// Protected Routes (Requires a Valid JWT)
		// ---------------------------------------------------------------------
		r.Group(func(r chi.Router) {
			r.Use(cfg.AuthMiddleware.RequireAuthentication())

			// --- Domains & SSL ---
			r.Route("/domains", func(r chi.Router) {
				r.With(cfg.AuthMiddleware.RequirePermission("domains", "read")).
					Get("/", cfg.DomainHandler.List)
				
				r.With(cfg.AuthMiddleware.RequirePermission("domains", "write")).
					Post("/", cfg.DomainHandler.Create)
				
				r.With(cfg.AuthMiddleware.RequirePermission("domains", "delete")).
					Delete("/{id}", cfg.DomainHandler.Delete)
				
				r.With(cfg.AuthMiddleware.RequirePermission("domains", "write")).
					Post("/{id}/ssl", cfg.DomainHandler.ProvisionSSL)
			})

			// --- Applications & Deployments ---
			r.Route("/applications", func(r chi.Router) {
				r.With(cfg.AuthMiddleware.RequirePermission("applications", "read")).
					Get("/", cfg.AppHandler.List)
				
				r.With(cfg.AuthMiddleware.RequirePermission("applications", "write")).
					Post("/", cfg.AppHandler.Create)
				
				r.With(cfg.AuthMiddleware.RequirePermission("applications", "read")).
					Get("/{id}", cfg.AppHandler.GetByID)
				
				r.With(cfg.AuthMiddleware.RequirePermission("applications", "write")).
					Put("/{id}/env", cfg.AppHandler.UpdateEnv)
				
				r.With(cfg.AuthMiddleware.RequirePermission("applications", "deploy")).
					Post("/{id}/deploy", cfg.AppHandler.TriggerDeploy)
			})

			// --- Privacy-First Observability & Audit Logs ---
			r.With(cfg.AuthMiddleware.RequirePermission("audit_logs", "read")).
				Get("/audit", cfg.AuditHandler.HandleGetTenantLogs)

			r.With(cfg.AuthMiddleware.RequirePermission("server", "manage")).
				Get("/admin/alerts", cfg.AuditHandler.HandleGetAdminAlerts)

			// --- WebSocket Real-Time Terminal Streaming ---
			r.With(cfg.AuthMiddleware.RequirePermission("applications", "read")).
				Get("/ws/deployments/{trace_id}", cfg.WSHandler.StreamDeploymentLogs)
		})
	})

	// Health Check / Ping Endpoint for Uptime Monitors (e.g., Uptime Kuma)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	return r
}
