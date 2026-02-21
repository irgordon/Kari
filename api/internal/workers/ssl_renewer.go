// api/internal/workers/ssl_renewer.go
package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"kari/api/internal/config"
	"kari/api/internal/core/domain"
	"kari/api/internal/core/services"
	"kari/api/internal/core/utils"
)

// ==============================================================================
// 1. Worker Struct (Dependency Injection)
// ==============================================================================

type SSLRenewer struct {
	Config       *config.Config
	DB           domain.DomainRepository
	SSLService   *services.SSLService
	AuditService domain.AuditService
	Logger       *slog.Logger
}

func NewSSLRenewer(
	cfg *config.Config,
	db domain.DomainRepository,
	sslService *services.SSLService,
	auditService domain.AuditService,
	logger *slog.Logger,
) *SSLRenewer {
	return &SSLRenewer{
		Config:       cfg,
		DB:           db,
		SSLService:   sslService,
		AuditService: auditService,
		Logger:       logger,
	}
}

// ==============================================================================
// 2. Lifecycle Management (Graceful Shutdowns)
// ==============================================================================

func (w *SSLRenewer) Start(ctx context.Context) {
	w.Logger.Info("🛡️ SSL Auto-Renewal Worker started")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	w.checkAndRenew(ctx)

	for {
		select {
		case <-ctx.Done():
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

	domains, err := w.DB.GetDomainsWithActiveSSL(ctx)
	if err != nil {
		w.Logger.Error("Failed to fetch domains for SSL check", slog.String("error", err.Error()))
		return
	}

	renewCount := 0
	failCount := 0

	for _, dom := range domains {
		// INJECTED: Read the path dynamically from config instead of hardcoding /etc/kari/ssl
		certPath := fmt.Sprintf("%s/%s/fullchain.pem", w.Config.SSLStorageDir, dom.DomainName)
		
		expiresAt, err := utils.GetCertExpiration(certPath)
		if err != nil {
			w.Logger.Warn("Could not parse certificate, skipping", 
				slog.String("domain", dom.DomainName), 
				slog.String("error", err.Error()),
			)
			continue
		}

		daysUntilExpiry := time.Until(expiresAt).Hours() / 24

		if daysUntilExpiry <= 30 {
			w.Logger.Info("♻️ Certificate expiring soon, initiating renewal", 
				slog.String("domain", dom.DomainName),
				slog.Float64("days_left", daysUntilExpiry),
			)

			err := w.SSLService.ProvisionCertificate(ctx, dom.UserID, dom.ID)
			if err != nil {
				w.Logger.Error("Failed to renew certificate", 
					slog.String("domain", dom.DomainName),
					slog.String("error", err.Error()),
				)
				
				w.AuditService.LogSystemAlert(
					ctx, 
					"ssl_renewal_failed", 
					"ssl", 
					dom.ID, 
					err, 
					"critical",
				)
				
				failCount++
				continue
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
