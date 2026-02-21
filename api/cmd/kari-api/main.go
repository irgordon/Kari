// api/cmd/kari-api/main.go
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
	"kari/api/internal/workers"
	"kari/api/internal/grpc/rustagent" // Generated from protobuf
)

func main() {
	// ==============================================================================
	// 1. Core Telemetry & Configuration
	// ==============================================================================
	
	// Initialize structured JSON logging for secure, parseable audit trails
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("🚀 Booting Karı API Orchestrator...")

	// Load dynamic configuration (No hardcoded paths)
	cfg := config.Load()

	// ==============================================================================
	// 2. Infrastructure Connections (The "Outbound" SLAs)
	// ==============================================================================

	// Initialize PostgreSQL Connection Pool
	dbPool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("✅ Connected to PostgreSQL")

	// Initialize gRPC connection to the root-level Rust Agent via Unix Domain Socket
	// We use an insecure dialer here because Unix sockets are physically isolated to the host OS.
	grpcDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", addr)
	}
	
	grpcConn, err := grpc.Dial(
		"/var/run/kari/agent.sock", 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(grpcDialer),
	)
	if err != nil {
		logger.Error("Failed to connect to Rust Agent", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer grpcConn.Close()
	
	// Instantiate the Protobuf client
	agentClient := rustagent.NewSystemAgentClient(grpcConn)
	logger.Info("✅ Connected to Rust System Agent")

	// ==============================================================================
	// 3. Dependency Injection (Wiring the Layers)
	// ==============================================================================

	// -- Repositories (Data Access) --
	appRepo := postgres.NewApplicationRepository(dbPool)
	domainRepo := postgres.NewDomainRepository(dbPool)
	auditRepo := postgres.NewAuditRepository(dbPool)
	userRepo := postgres.NewUserRepository(dbPool)
	roleRepo := postgres.NewRoleRepository(dbPool)

	// -- Core Services (Business Logic) --
	auditService := services.NewAuditService(auditRepo, logger)
	roleService := services.NewRoleService(roleRepo, logger)
	authService := services.NewAuthService(userRepo, logger, cfg.JWTSecret)
	
	sslService := services.NewSSLService(
		cfg, 
		domainRepo, 
		agentClient, 
		auditService, 
		logger,
	)
	
	appService := services.NewAppService(
		cfg,
		appRepo,
		domainRepo,
		agentClient,
		auditService,
		logger,
	)

	// -- HTTP Handlers (Transport Layer) --
	authHandler := handlers.NewAuthHandler(authService)
	appHandler := handlers.NewAppHandler(appService)
	domainHandler := handlers.NewDomainHandler(sslService, domainRepo)
	auditHandler := handlers.NewAuditHandler(auditService)
	wsHandler := handlers.NewWebSocketHandler(logger)

	// -- Middleware --
	authMiddleware := middleware.NewAuthMiddleware(authService, roleService, logger)

	// ==============================================================================
	// 4. Background Workers (Automated System Maintenance)
	// ==============================================================================

	// Create a root context for background workers that we can cancel on shutdown
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	sslRenewer := workers.NewSSLRenewer(cfg, domainRepo, sslService, auditService, logger)
	
	// Start the cron worker in an isolated Goroutine
	go sslRenewer.Start(workerCtx)

	// ==============================================================================
	// 5. HTTP Server Initialization
	// ==============================================================================

	// Construct the chi router with our deeply-layered security middleware
	routerConfig := router.RouterConfig{
		AuthHandler:    authHandler,
		AppHandler:     appHandler,
		DomainHandler:  domainHandler,
		AuditHandler:   auditHandler,
		WSHandler:      wsHandler,
		AuthMiddleware: authMiddleware,
		Logger:         logger,
	}
	
	mux := router.NewRouter(routerConfig)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,  // Mitigate Slowloris attacks
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ==============================================================================
	// 6. Graceful Shutdown & Signal Handling
	// ==============================================================================

	// Listen for OS interrupt signals (e.g., Ctrl+C, systemctl stop kari-api)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine so it doesn't block the signal listener
	go func() {
		logger.Info("🌐 HTTP Gateway listening", slog.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP Server crashed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Block main thread until a termination signal is received
	<-stop
	logger.Info("🛑 Termination signal received. Initiating graceful shutdown...")

	// 1. Cancel background workers (stops SSL renewal loops mid-sleep)
	cancelWorkers()

	// 2. Shut down the HTTP server (gives active requests 10 seconds to finish)
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP Server forced to shutdown", slog.String("error", err.Error()))
	} else {
		logger.Info("✅ HTTP Server stopped cleanly")
	}

	// dbPool and grpcConn will be cleanly closed by their defers at the end of main()
	logger.Info("👋 Karı Orchestrator shutdown complete.")
}
