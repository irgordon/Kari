// api/internal/db/postgres/application_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kari/api/internal/core/domain"
)

// ApplicationRepo implements domain.ApplicationRepository using pgx/v5
type ApplicationRepo struct {
	pool *pgxpool.Pool
}

// NewApplicationRepo is the factory function complying with our DI standards
func NewApplicationRepo(pool *pgxpool.Pool) domain.ApplicationRepository {
	return &ApplicationRepo{pool: pool}
}

// Create inserts a new application with the required app_user identity for the Rust Muscle
func (r *ApplicationRepo) Create(ctx context.Context, app *domain.Application) error {
	// 🛡️ SLA Sync: We added app_user to the insert to support Systemd/Jailing logic
	query := `
		INSERT INTO applications (domain_id, repo_url, branch, build_command, start_command, env_vars, port, app_user, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	// 🛡️ pgx natively handles map[string]string -> JSONB, no manual Marshal needed.
	err := r.pool.QueryRow(ctx, query,
		app.DomainID,
		app.RepoURL,
		app.Branch,
		app.BuildCommand,
		app.StartCommand,
		app.EnvVars, 
		app.Port,
		app.AppUser, // Required for Rust Agent jail identity
		app.Status,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	return nil
}

// GetByID fetches an app with strict IDOR protection via domain-ownership join
func (r *ApplicationRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Application, error) {
	// 🛡️ Zero-Trust: Joined on domains.user_id to ensure the requester owns the resource
	query := `
		SELECT a.id, a.domain_id, a.repo_url, a.branch, a.build_command, a.start_command, a.env_vars, a.port, a.app_user, a.status, a.created_at, a.updated_at
		FROM applications a
		INNER JOIN domains d ON a.domain_id = d.id
		WHERE a.id = $1 AND d.user_id = $2
	`

	// pgx.RowToStructByName automatically maps the columns to the struct fields.
	// This reduces the scan-list maintenance burden as the domain model grows.
	rows, err := r.pool.Query(ctx, query, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Application])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound 
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return &app, nil
}
