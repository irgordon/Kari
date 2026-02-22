use async_trait::async_trait;
use crate::proto::agent_v1::{AgentResponse, FirewallPolicy, JobIntent, SslPayload};
use std::io;

/// 🛡️ SLA: ProxyManager defines the interface for reverse-proxy orchestration.
/// It abstractly handles Nginx/Apache configuration without the Brain needing 
/// to know the specific underlying server.
#[async_trait]
pub trait ProxyManager: Send + Sync {
    /// Configures a new virtual host for a jail.
    async fn configure_vhost(&self, app_id: &str, domain: &str, port: u32) -> Result<(), io::Error>;
    
    /// Installs and validates SSL certificates. 
    /// Implementations must zeroize private keys after writing to disk.
    async fn install_ssl(&self, payload: SslPayload) -> Result<(), io::Error>;
    
    /// Performs a hot-reload of the proxy service (e.g., systemctl reload nginx).
    async fn reload(&self) -> Result<(), io::Error>;
}

/// 🛡️ Zero-Trust: FirewallManager enforces network boundaries at the host level.
#[async_trait]
pub trait FirewallManager: Send + Sync {
    /// Applies a specific firewall intent (ALLOW/DENY) using nftables or iptables.
    async fn apply_policy(&self, policy: FirewallPolicy) -> Result<(), io::Error>;
}

/// 🛡️ SOLID: JobScheduler handles the persistence and execution of recurring tasks.
#[async_trait]
pub trait JobScheduler: Send + Sync {
    /// Schedules a binary for execution based on a cron-style expression.
    /// Must validate that the 'run_as_user' exists to prevent privilege escalation.
    async fn schedule_job(&self, intent: JobIntent) -> Result<(), io::Error>;
    
    /// Removes a job from the system scheduler.
    async fn unschedule_job(&self, job_name: &str) -> Result<(), io::Error>;
}

/// 🛡️ Muscle Core: The primary interface for Jail Lifecycle Management.
#[async_trait]
pub trait JailManager: Send + Sync {
    /// Provisions a transient systemd-unit with resource constraints.
    async fn create_jail(
        &self, 
        app_id: &str, 
        command: &str, 
        env_vars: std::collections::HashMap<String, String>,
        memory_limit_mb: u32
    ) -> Result<AgentResponse, io::Error>;
    
    /// Tears down the jail and cleans up the cgroup/namespace.
    async fn destroy_jail(&self, app_id: &str) -> Result<(), io::Error>;
}