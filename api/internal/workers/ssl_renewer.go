// api/internal/workers/ssl_renewer.go
package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"kari/api/internal/core/domain"
	"kari/api/internal/core/services"
	"kari/api/internal/core/utils"
)

// ==============================================================================
// 1. Worker Struct (Dependency Injection)
// ==============================================================================

type SSLRenewer struct {
	DB           domain.DomainRepository
	SSLService   *services.SSLService
	AuditService domain.AuditService
	Logger       *slog.Logger
}

func NewSSLRenewer(
	db domain.DomainRepository,
	sslService *services.SSLService,
	auditService domain.AuditService,
	logger *slog.Logger,
) *SSLRenewer {
	return &SSLRenewer{
		DB:           db,
		SSLService:   sslService,
		AuditService: auditService,
		Logger:       logger,
	}
}

// ==============================================================================
// 2. Lifecycle Management (Graceful Shutdowns)
// ==============================================================================

// Start kicks off the background cron job. It blocks, so it must be run in a goroutine
// from main.go. It listens to the system context to ensure graceful teardown.
func (w *SSLRenewer) Start(ctx context.Context) {
	w.Logger.Info("🛡️ SSL Auto-Renewal Worker started")

	// Wake up once every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run an initial check immediately on startup so admins don't have to wait
	// a full 24 hours to see if broken certificates fix themselves after a reboot.
	w.checkAndRenew(ctx)

	for {
		select {
		case <-ctx.Done():
			// The OS sent a SIGTERM (e.g., systemctl restart kari-api)
			w.Logger.Info("🛑 Shutting down SSL Auto-Renewal Worker gracefully")
			return
		case <-ticker.C:
			w.checkAndRenew(ctx)
		}
	}
}

// ==============================================================================
// 3. Core Worker Logic
// ==============================================================================

func (w *SSLRenewer) checkAndRenew(ctx context.Context) {
	w.Logger.Info("🔍 Running daily SSL expiration check...")

	// 1. Fetch all domains that currently have an active SSL
	domains, err := w.DB.GetDomainsWithActiveSSL(ctx)
	if err != nil {
		w.Logger.Error("Failed to fetch domains for SSL check", slog.String("error", err.Error()))
		return
	}

	renewCount := 0
	failCount := 0

	// 2. Iterate through each domain and check its expiration
	for _, dom := range domains {
		// Because the Rust Agent wrote this public certificate with 0644 permissions,
		// the unprivileged Go API can read it directly from the disk.
		certPath := fmt.Sprintf("/etc/kari/ssl/%s/fullchain.pem", dom.DomainName)
		
		expiresAt, err := utils.GetCertExpiration(certPath)
		if err != nil {
			w.Logger.Warn("Could not parse certificate, skipping", 
				slog.String("domain", dom.DomainName), 
				slog.String("error", err.Error()),
			)
			continue
		}

		// 3. Let's Encrypt recommends renewing 30 days before expiration.
		daysUntilExpiry := time.Until(expiresAt).Hours() / 24

		if daysUntilExpiry <= 30 {
			w.Logger.Info("♻️ Certificate expiring soon, initiating renewal", 
				slog.String("domain", dom.DomainName),
				slog.Float64("days_left", daysUntilExpiry),
			)

			// 4. Re-use our existing SSLService (SOLID: SRP)
			// We pass the domain's UserID to satisfy the service layer's IDOR checks, 
			// even though this is an automated system task.
			err := w.SSLService.ProvisionCertificate(ctx, dom.UserID, dom.ID)
			if err != nil {
				w.Logger.Error("Failed to renew certificate", 
					slog.String("domain", dom.DomainName),
					slog.String("error", err.Error()),
				)
				
				// 🚨 5. Proactive Alerting
				// This logs a critical alert to the database so the Admin sees it in the 
				// SvelteKit UI Action Center immediately, rather than waiting for a customer complaint.
				w.AuditService.LogSystemAlert(
					ctx, 
					"ssl_renewal_failed", 
					"ssl", 
					dom.ID, 
					err, 
					"critical",
				)
				
				failCount++
				continue // One failure shouldn't stop other domains from renewing
			}
			
			renewCount++
		}
	}

	if renewCount > 0 || failCount > 0 {
		w.Logger.Info("✅ SSL renewal sweep completed", 
			slog.Int("renewed_count", renewCount),
			slog.Int("failed_count", failCount),
		)
	} else {
		w.Logger.Info("✅ SSL renewal sweep completed. No renewals needed today.")
	}
}
