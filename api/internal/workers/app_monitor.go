package workers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"kari/api/internal/core/domain"
	"math/rand"
)

// AppMonitor implements the proactive heartbeat logic
type AppMonitor struct {
	repo       domain.ApplicationRepository
	auditRepo  domain.AuditRepository
	httpClient *http.Client
	logger     *slog.Logger
	interval   time.Duration
}

func NewAppMonitor(
	repo domain.ApplicationRepository,
	audit domain.AuditRepository,
	logger *slog.Logger,
) *AppMonitor {
	return &AppMonitor{
		repo:      repo,
		auditRepo: audit,
		logger:    logger,
		interval:  1 * time.Minute,
		httpClient: &http.Client{
			// 🛡️ SLA: Strict timeout prevents worker from hanging on zombie apps
			Timeout: 5 * time.Second,
		},
	}
}

// Start initiates the background loop with graceful shutdown support
func (m *AppMonitor) Start(ctx context.Context) {
	m.logger.Info("Starting Proactive AppMonitor Worker")
	
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping AppMonitor Worker")
			return
		case <-ticker.C:
			m.performHealthChecks(ctx)
		}
	}
}

func (m *AppMonitor) performHealthChecks(ctx context.Context) {
	// 1. Fetch all active applications from the Muscle
	apps, err := m.repo.ListAllActive(ctx)
	if err != nil {
		m.logger.Error("Failed to fetch apps for health check", slog.Any("error", err))
		return
	}

	for _, app := range apps {
		// 🛡️ SLA: Add random Jitter to prevent "Thundering Herd"
		jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
		
		go func(a domain.Application) {
			time.Sleep(jitter)
			m.checkAppHealth(ctx, a)
		}(app)
	}
}

func (m *AppMonitor) checkAppHealth(ctx context.Context, app domain.Application) {
	// 🛡️ Platform Agnostic Check
	// We check the local loopback port where the app is jailed.
	url := fmt.Sprintf("http://127.0.0.1:%d/health", app.Port)
	
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := m.httpClient.Do(req)

	isUp := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	// 🛡️ State-Transition Logic (Efficiency)
	// Only update the database/Action Center if the status has actually changed.
	if !isUp && app.Status == "running" {
		m.handleAppFailure(ctx, app, err)
	} else if isUp && app.Status == "failed" {
		m.handleAppRecovery(ctx, app)
	}
}

func (m *AppMonitor) handleAppFailure(ctx context.Context, app domain.Application, err error) {
	m.logger.Warn("App health check failed", 
		slog.String("app", app.Name), 
		slog.Any("error", err))

	// 1. Mark as failed in DB
	_ = m.repo.UpdateStatus(ctx, app.ID, "failed")

	// 2. Escalate to the Action Center
	_ = m.auditRepo.CreateAlert(ctx, &domain.SystemAlert{
		Severity: "critical",
		Category: "uptime",
		Message:  fmt.Sprintf("Application %s is unreachable on port %d", app.Name, app.Port),
		Metadata: map[string]any{"app_id": app.ID, "error": fmt.Sprintf("%v", err)},
	})
}

func (m *AppMonitor) handleAppRecovery(ctx context.Context, app domain.Application) {
	m.logger.Info("App recovered", slog.String("app", app.Name))
	_ = m.repo.UpdateStatus(ctx, app.ID, "running")
}
