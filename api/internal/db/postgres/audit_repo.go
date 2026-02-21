// api/internal/db/postgres/audit_repo.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"kari/api/internal/core/domain"
)

// ==============================================================================
// 1. Repository Struct
// ==============================================================================

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// ==============================================================================
// 2. Tenant Audit Logging (User-Facing Actions)
// ==============================================================================

/**
 * LogActivity records a specific action taken by a user within a tenant context.
 * This satisfies the SLA for non-repudiation.
 */
func (r *AuditRepository) LogActivity(ctx context.Context, entry domain.AuditEntry) error {
	query := `
		INSERT INTO tenant_logs (
			id, tenant_id, user_id, action, resource_type, resource_id, metadata, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		uuid.New(),
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.Metadata, // Stored as JSONB for platform flexibility
		entry.IPAddress,
		time.Now().UTC(),
	)
	return err
}

// ==============================================================================
// 3. System Alerting (Infrastructure Health)
// ==============================================================================

/**
 * CreateAlert generates a proactive system notification.
 * These are the records surfaced in the "Action Center" UI widget.
 */
func (r *AuditRepository) CreateAlert(ctx context.Context, alert domain.SystemAlert) error {
	query := `
		INSERT INTO system_alerts (
			id, severity, category, resource_id, message, error_details, is_resolved, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		uuid.New(),
		alert.Severity,
		alert.Category,
		alert.ResourceID,
		alert.Message,
		alert.ErrorDetails,
		false,
		time.Now().UTC(),
	)
	return err
}

/**
 * GetUnresolvedAlerts fetches proactive notifications for the Action Center.
 * Implements strict data filtering to ensure only high-priority health data is returned.
 */
func (r *AuditRepository) GetUnresolvedAlerts(ctx context.Context) ([]domain.SystemAlert, error) {
	query := `
		SELECT id, severity, category, resource_id, message, created_at
		FROM system_alerts
		WHERE is_resolved = false
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []domain.SystemAlert
	for rows.Next() {
		var a domain.SystemAlert
		if err := rows.Scan(&a.ID, &a.Severity, &a.Category, &a.ResourceID, &a.Message, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

/**
 * ResolveAlert marks a system issue as addressed, clearing it from the UI.
 */
func (r *AuditRepository) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `UPDATE system_alerts SET is_resolved = true, resolved_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), alertID)
	return err
}

// ==============================================================================
// 4. Platform-Agnostic Querying (SLA)
// ==============================================================================

/**
 * GetTenantLogs provides paginated, filtered access to audit history.
 */
func (r *AuditRepository) GetTenantLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.AuditEntry, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, metadata, created_at
		FROM tenant_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditEntry
	for rows.Next() {
		var l domain.AuditEntry
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.ResourceType, &l.ResourceID, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
