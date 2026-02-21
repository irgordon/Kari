package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SystemProfile dictates the global behavior, resource limits, and safety boundaries 
// of the orchestration engine. 
// 🛡️ SLA: This struct strictly holds INTENT, not OS-level deployment paths.
type SystemProfile struct {
	ID                    uuid.UUID         `json:"id"`
	
	// 📦 Stack & Routing Governance
	DefaultStackRegistry  map[string]string `json:"stack_defaults"`  // e.g., {"php": "8.3", "node": "20"}
	SSLStrategy           string            `json:"ssl_strategy"`    // e.g., "letsencrypt_http01"
	
	// 🛡️ Resource Jailing (SLA Enforcement)
	// The Brain passes these to the Muscle to generate strict systemd cgroup constraints.
	MaxMemoryPerAppMB     int               `json:"max_memory_per_app_mb"` 
	MaxCPUPercentPerApp   int               `json:"max_cpu_percent_per_app"`
	
	// 🛡️ Security & Identity Policies
	DefaultFirewallPolicy string            `json:"default_firewall_policy"` // e.g., "deny_all_except_80_443"
	
	// The Brain acts as the ledger for UID assignments to prevent collisions
	AppUserUIDRangeStart  int               `json:"app_user_uid_range_start"` // e.g., 5000
	AppUserUIDRangeEnd    int               `json:"app_user_uid_range_end"`   // e.g., 6000
	
	// 💾 Backup & Retention
	BackupRetentionDays   int               `json:"backup_retention_days"`
	
	UpdatedAt             time.Time         `json:"updated_at"`
}

// SystemProfileRepository defines the interface for state persistence.
// 🛡️ SOLID: Driven by adapters (like PostgreSQL) in the outer layers.
type SystemProfileRepository interface {
	// GetActiveProfile returns the singleton system configuration.
	// We pass ctx to ensure database calls respect HTTP request timeouts.
	GetActiveProfile(ctx context.Context) (*SystemProfile, error)
	
	// UpdateProfile mutates the system state. 
	// Changes to resource limits here will be applied asynchronously by a background 
	// Go worker reconciling the state with the Rust Agent.
	UpdateProfile(ctx context.Context, profile *SystemProfile) error
}
