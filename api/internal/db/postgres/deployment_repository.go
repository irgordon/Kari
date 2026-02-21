package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"kari/api/internal/core/domain"
)

type PostgresDeploymentRepository struct {
	db *sql.DB
}

func NewPostgresDeploymentRepository(db *sql.DB) *PostgresDeploymentRepository {
	return &PostgresDeploymentRepository{db: db}
}

// ClaimNextPending 🛡️ Zero-Trust Concurrency
// Uses 'SKIP LOCKED' to allow multiple Brain instances to process the queue without conflicts.
func (r *PostgresDeploymentRepository) ClaimNextPending(ctx context.Context) (*domain.Deployment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE deployments
		SET status = $1, updated_at = NOW()
		WHERE id = (
			SELECT id FROM deployments
			WHERE status = 'PENDING'
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, app_id, domain_name, repo_url, branch, build_command, target_port, encrypted_ssh_key;
	`

	d := &domain.Deployment{}
	err = tx.QueryRowContext(ctx, query, domain.StatusRunning).Scan(
		&d.ID, &d.AppID, &d.DomainName, &d.RepoURL, &d.Branch, 
		&d.BuildCommand, &d.TargetPort, &d.EncryptedSSHKey,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Queue is empty
		}
		return nil, fmt.Errorf("db: failed to claim deployment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return d, nil
}

// AppendLog 🛡️ SLA Visibility
// Writes a log chunk to the database for the Kari Panel UI to consume.
func (r *PostgresDeploymentRepository) AppendLog(ctx context.Context, deploymentID string, content string) error {
	query := `INSERT INTO deployment_logs (deployment_id, content) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, deploymentID, content)
	return err
}

// UpdateStatus 🛡️ State Machine Integrity
func (r *PostgresDeploymentRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	query := `UPDATE deployments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
