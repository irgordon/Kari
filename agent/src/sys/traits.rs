use async_trait::async_trait;
use std::net::IpAddr;
use std::collections::HashMap;
use tokio::sync::mpsc;
use tonic::Status;
use crate::server::kari_agent::LogChunk;
use crate::sys::secrets::ProviderCredential;

// ==============================================================================
// 1. GitOps & Source Control (Transient Auth)
// ==============================================================================

#[async_trait]
pub trait GitManager: Send + Sync {
    /// Clones a repository into a target directory.
    /// 🛡️ ssh_key: An optional private key provided by the Go Brain. 
    /// If Some, it must be used for this operation only and never persisted.
    async fn clone_repo(
        &self, 
        repo_url: &str, 
        branch: &str, 
        target_dir: &str,
        ssh_key: Option<&str> 
    ) -> Result<(), String>;
}

// ==============================================================================
// 2. Build & Execution (Telemetry-Aware)
// ==============================================================================

#[async_trait]
pub trait BuildManager: Send + Sync {
    /// Executes a build command within an unprivileged jail.
    /// 🛡️ log_tx: A streaming channel to pipe stdout/stderr back to the gRPC stream.
    async fn execute_build(
        &self,
        build_command: &str,
        working_dir: &str,
        run_as_user: &str,
        env_vars: &HashMap<String, String>,
        log_tx: mpsc::Sender<Result<LogChunk, Status>>,
        trace_id: String,
    ) -> Result<(), String>;
}

// ==============================================================================
// 3. Firewall Abstraction (Type-Safe & Zero-Trust)
// ==============================================================================

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FirewallAction { Allow, Deny, Reject }

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Protocol { Tcp, Udp, Both }

pub struct FirewallPolicy {
    pub action: FirewallAction,
    pub port: u16,
    pub protocol: Protocol,
    pub source_ip: Option<IpAddr>,
}

#[async_trait]
pub trait FirewallManager: Send + Sync {
    async fn apply_policy(&self, policy: &FirewallPolicy) -> Result<(), String>;
}

// ==============================================================================
// 4. SSL Engine Abstraction (Memory Safe)
// ==============================================================================

pub struct SslPayload {
    pub domain_name: String,
    pub fullchain_pem: Vec<u8>,
    /// 🛡️ Zero-Copy Secret. Handled by the ProviderCredential wrapper 
    /// to ensure memory is zeroized after use.
    pub privkey_pem: ProviderCredential, 
}

#[async_trait]
pub trait SslEngine: Send + Sync {
    async fn install_certificate(&self, payload: SslPayload) -> Result<(), String>;
}
