// agent/src/main.rs

use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use tokio::net::UnixListener;
use tonic::transport::Server;

mod config;
mod server;
mod sys;

use crate::config::AgentConfig;
use crate::server::kari_agent::system_agent_server::SystemAgentServer;
use crate::server::KariAgentService;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // ==============================================================================
    // 1. Configuration & Environment (Platform Agnostic)
    // ==============================================================================
    
    tracing_subscriber::fmt::init();
    let config = AgentConfig::load();
    
    let socket_path = config.socket_path.clone(); 
    // 🛡️ SLA Fix: Remove unwrap() to prevent panic if socket is in the current directory
    let socket_dir = Path::new(&socket_path).parent().unwrap_or_else(|| Path::new("."));

    // ==============================================================================
    // 2. Secure Socket Initialization & Ownership Transfer
    // ==============================================================================

    if !socket_dir.exists() {
        fs::create_dir_all(socket_dir)?;
    }

    if Path::new(&socket_path).exists() {
        fs::remove_file(&socket_path)?;
    }

    let listener = UnixListener::bind(&socket_path)?;
    let expected_api_uid = config.expected_api_uid;
    
    // 🛡️ DEFENSE IN DEPTH: Restrict socket permissions
    let mut perms = fs::metadata(&socket_path)?.permissions();
    perms.set_mode(0o660);
    fs::set_permissions(&socket_path, perms)?;

    // 🛡️ THE LOCKOUT FIX: Transfer ownership to the Go API user.
    // Because the agent is root, it can hand the file over to the unprivileged Brain.
    // If we don't do this, the Go API gets "Permission Denied" before gRPC even starts.
    let chown_status = std::process::Command::new("chown")
        .arg(&format!("{}", expected_api_uid))
        .arg(&socket_path)
        .status()?;

    if !chown_status.success() {
        return Err("FATAL: Failed to chown socket file for Go API access".into());
    }

    // ==============================================================================
    // 3. SLA Boundary: Kernel-Level Peer Credential Interceptor
    // ==============================================================================
    
    let incoming_stream = async_stream::stream! {
        loop {
            match listener.accept().await {
                Ok((stream, _)) => {
                    match stream.peer_cred() {
                        Ok(cred) => {
                            if cred.uid() == expected_api_uid || cred.uid() == 0 {
                                tracing::debug!("✅ Authenticated gRPC connection from UID: {}", cred.uid());
                                yield Ok::<_, std::io::Error>(stream);
                            } else {
                                tracing::warn!(
                                    "🚨 BLOCKED unauthorized socket connection attempt from UID: {} / GID: {}", 
                                    cred.uid(), cred.gid()
                                );
                            }
                        }
                        Err(e) => tracing::error!("Failed to read peer credentials: {}", e),
                    }
                }
                Err(e) => {
                    tracing::error!("Socket accept error: {}", e);
                    yield Err(e);
                }
            }
        }
    };

    // ==============================================================================
    // 4. Dependency Injection & Service Start
    // ==============================================================================

    let agent_service = KariAgentService::new(config);

    tracing::info!("⚙️ Kari Rust Agent (The Muscle) securely listening on {}", socket_path);

    Server::builder()
        .add_service(SystemAgentServer::new(agent_service))
        .serve_with_incoming(incoming_stream)
        .await?;

    Ok(())
}
