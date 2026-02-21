package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SystemProfile dictates the global behavior and safety boundaries of the panel.
type SystemProfile struct {
	ID                    uuid.UUID         `json:"id"`
	
	// Stack Governance
	DefaultStackRegistry  map[string]string `json:"stack_defaults"`  // e.g., {"php": "8.3", "node": "20"}
	SSLStrategy           string            `json:"ssl_strategy"`    // e.g., "letsencrypt"
	
	// 🛡️ Resource Jailing (SLA Enforcement)
	// These are passed to the Rust Agent to populate systemd cgroup settings.
	MaxMemoryPerAppMB     int               `json:"max_memory_per_app_mb"` 
	MaxCPUPercentPerApp   int               `json:"max_cpu_percent_per_app"`
	
	// 🛡️ Security & Identity Policies
	DefaultFirewallPolicy string            `json:"default_firewall_policy"`
	AppUserUIDRangeStart  int               `json:"app_user_uid_range_start"` // e.g., 5000
	AppUserUIDRangeEnd    int               `json:"app_user_uid_range_end"`   // e.g., 6000
	
	// Infrastructure Details
	AgentSocketPath       string            `json:"agent_socket_path"` // e.g., "/run/kari/agent.sock"
	BackupRetentionDays   int               `json:"backup_retention_days"`
	
	UpdatedAt             time.Time         `json:"updated_at"`
}

// SystemProfileRepository defines the interface for state persistence.
type SystemProfileRepository interface {
	// GetActiveProfile returns the singleton system configuration.
	GetActiveProfile(ctx context.Context) (*SystemProfile, error)
	
	// UpdateProfile mutates the system state and triggers a notification 
	// to the Rust Muscle if critical paths (like SocketPath) change.
	UpdateProfile(ctx context.Context, profile *SystemProfile) error
}
