package domain

import (
	"context"
	"time"
	"github.com/google/uuid"
)

// Application represents the core domain model
type Application struct {
	ID           uuid.UUID         `json:"id"`
	DomainID     uuid.UUID         `json:"domain_id"`
	AppType      string            `json:"app_type"`
	RepoURL      string            `json:"repo_url"`
	Branch       string            `json:"branch"`
	BuildCommand string            `json:"build_command"`
	StartCommand string            `json:"start_command"`
	EnvVars      map[string]string `json:"env_vars"` // Maps to Postgres JSONB
	Port         int               `json:"port"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ApplicationRepository defines the STRICT contract for data persistence.
// Notice there is absolutely no mention of SQL, rows, or connections here.
type ApplicationRepository interface {
	Create(ctx context.Context, app *Application) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Application, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateEnvVars(ctx context.Context, id uuid.UUID, envVars map[string]string) error
}
