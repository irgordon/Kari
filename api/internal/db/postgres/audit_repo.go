// api/internal/db/postgres/audit_repo.go
package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kari/api/internal/core/domain"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// ==============================================================================
// Dynamic Action Center Querying
// ==============================================================================

/**
 * GetFilteredAlerts builds a dynamic SQL query based on UI filters.
 * This ensures that as the Karı UI grows, the data layer remains flexible
 * without needing dozens of hardcoded "GetByStatus" methods.
 */
func (r *AuditRepository) GetFilteredAlerts(ctx context.Context, filter domain.AlertFilter) ([]domain.SystemAlert, error) {
	// 1. Base query focusing on actionable infrastructure health
	query := `SELECT id, severity, category, resource_id, message, is_resolved, created_at 
	          FROM system_alerts WHERE 1=1`
	
	var args []any
	argCount := 1

	// 2. Dynamic filter building (Open/Closed Principle)
	// We append conditions dynamically without modifying the base logic
	if filter.IsResolved != nil {
		query += fmt.Sprintf(" AND is_resolved = $%d", argCount)
		args = append(args, *filter.IsResolved)
		argCount++
	}

	if filter.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, filter.Severity)
		argCount++
	}

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, filter.Category)
		argCount++
	}

	// 3. Finalize ordering (Latest issues first)
	query += " ORDER BY created_at DESC"

	// 4. Execution with strict Context (SLA Compliance)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filtered alerts: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.SystemAlert])
}

// ==============================================================================
// Atomic Alert Resolution
// ==============================================================================

/**
 * ResolveAlert handles the transition of an alert from active to resolved.
 * We use an atomic UPDATE to ensure consistency across the distributed Brain.
 */
func (r *AuditRepository) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `
		UPDATE system_alerts 
		SET is_resolved = true, resolved_at = $1 
		WHERE id = $2 AND is_resolved = false
	`
	tag, err := r.pool.Exec(ctx, query, time.Now().UTC(), alertID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}
