package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"kari/api/internal/api/handlers"
	"kari/api/internal/api/middleware"
	"kari/api/internal/api/router"
	"kari/api/internal/config"
	"kari/api/internal/core/services"
	"kari/api/internal/db/postgres"
	"kari/api/internal/infrastructure/crypto"
	"kari/api/internal/workers"
	"kari/api/internal/grpc/rustagent" 
)

func main() {
	// ==============================================================================
	// 1. Core Telemetry & Configuration
	// ==============================================================================
	
	// Structured JSON logging is mandatory for 2026 production observability
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("🚀 Booting Karı API Orchestrator (The Brain)...")

	// Load environment-driven configuration
	cfg := config.Load()

	// ==============================================================================
	// 2. Outbound Infrastructure SLAs
	// ==============================================================================

	// Initialize PostgreSQL (The State Layer)
	dbPool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("FATAL: Database connectivity failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize gRPC link to the Rust Muscle over Unix Domain Socket
	// UDS provides the lowest latency and highest security for on-host communication.
	grpcDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", addr)
	}
	
	grpcConn, err := grpc.Dial(
		cfg.AgentSocketPath, // e.g., "/var/run/kari/agent.sock"
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(grpcDialer),
	)
	if err != nil {
		logger.Error("FATAL: gRPC link to Rust Agent failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer grpcConn.Close()
	
	agentClient := rustagent.NewSystemAgentClient(grpcConn)

	// ==============================================================================
	// 3. Hardened Dependency Injection
	// ==============================================================================

	// -- 🛡️ Security Engine --
	// Instantiate AEAD Crypto Service for context-bound secret management
	cryptoService, err := crypto.NewAESCryptoService(cfg.MasterKeyHex)
	if err != nil {
		logger.Error("FATAL: Cryptographic engine initialization failed", slog.Any("error", err))
		os.Exit(1)
	}

	// -- Repositories --
	appRepo    := postgres.NewApplicationRepository(dbPool)
	auditRepo  := postgres.NewAuditRepository(dbPool)
	userRepo   := postgres.NewUserRepository(dbPool)

	// -- Core Services --
	// Auth handles hashing and JWT signing via the configured secret
	authService := services.NewAuthService(userRepo, logger, cfg)
	
	// EnvVarService uses the cryptoService to wrap/unwrap app secrets
	envVarService := services.NewEnvVarService(appRepo, cryptoService, logger)
	
	// AppService orchestrates the high-level GitOps/Deployment lifecycle
	appService := services.NewApplicationService(
		appRepo,
		auditRepo,
		agentClient,
		logger,
		envVarService,
	)

	// -- HTTP Transport Handlers --
	authHandler   := handlers.NewAuthHandler(authService)
	appHandler    := handlers.NewAppHandler(appService, envVarService)
	auditHandler  := handlers.NewAuditHandler(auditRepo)
	
	// -- Middleware --
	// AuthMiddleware protects routes based on the JWT state and User Rank
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// ==============================================================================
	// 4. Background Workers (Proactive Feedback Loop)
	// ==============================================================================

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// Proactive Health Monitor: Checks app availability every 60s with 10-worker concurrency
	appMonitor := workers.NewAppMonitor(appRepo, auditRepo, logger, 1*time.Minute)
	go appMonitor.Start(workerCtx)

	// ==============================================================================
	// 5. HTTP Gateway Lifecycle
	// ==============================================================================

	mux := router.NewRouter(router.RouterConfig{
		AuthHandler:    authHandler,
		AppHandler:     appHandler,
		AuditHandler:   auditHandler,
		AuthMiddleware: authMiddleware,
		Logger:         logger,
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ==============================================================================
	// 6. Signal Handling & Graceful Exit
	// ==============================================================================

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("🌐 HTTP Gateway active", slog.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("CRITICAL: Server crashed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("🛑 Termination signal received. Flushing buffers...")

	// 1. Terminate background processes (Stop monitoring/cron tasks)
	cancelWorkers()

	// 2. Drain active HTTP connections (10-second window)
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("ERROR: Forced shutdown occurred", slog.Any("error", err))
	}

	logger.Info("✅ Karı Orchestrator shutdown complete. Stay efficient.")
}
