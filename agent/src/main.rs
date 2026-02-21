use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use tokio::net::UnixListener;
use tokio::signal;
use tonic::transport::Server;
use tracing::{info, error, warn, debug};

mod config;
mod server;
mod sys;

use crate::config::AgentConfig;
use crate::server::kari_agent::system_agent_server::SystemAgentServer;
use crate::server::KariAgentService;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Core Telemetry
    tracing_subscriber::fmt::init();
    info!("🚀 Karı Rust Agent (The Muscle) initializing...");

    let config = AgentConfig::load();
    let socket_path = &config.socket_path;
    let socket_dir = Path::new(socket_path).parent().unwrap_or_else(|| Path::new("."));

    // 2. Filesystem Preparation
    if !socket_dir.exists() {
        fs::create_dir_all(socket_dir)?;
    }
    if Path::new(socket_path).exists() {
        debug!("Cleaning up stale socket at {}", socket_path);
        fs::remove_file(socket_path)?;
    }

    // 3. Bind and Secure the Socket
    let listener = UnixListener::bind(socket_path)?;
    
    // 🛡️ Lock down the file itself
    let mut perms = fs::metadata(socket_path)?.permissions();
    perms.set_mode(0o660); // rw-rw----
    fs::set_permissions(socket_path, perms)?;

    // 🛡️ Handover to the Go API User
    // We use numeric IDs to remain platform agnostic (no need to parse /etc/passwd)
    let uid = config.expected_api_uid;
    let gid = config.expected_api_gid;
    nix::unistd::chown(
        Path::new(socket_path),
        Some(nix::unistd::Uid::from_raw(uid)),
        Some(nix::unistd::Gid::from_raw(gid)),
    ).map_err(|e| format!("SLA Failure: Failed to chown socket: {}", e))?;

    // 4. Peer Credential Guard (Kernel-Level Auth)
    let incoming_stream = async_stream::stream! {
        loop {
            match listener.accept().await {
                Ok((stream, _)) => {
                    if let Ok(cred) = stream.peer_cred() {
                        if cred.uid() == uid || cred.uid() == 0 {
                            debug!("✅ Verified connection from UID {}", cred.uid());
                            yield Ok::<_, std::io::Error>(stream);
                        } else {
                            warn!("🚨 REJECTED: Unauthorized connection attempt from UID {}", cred.uid());
                        }
                    }
                }
                Err(e) => {
                    error!("Socket accept failure: {}", e);
                    yield Err(e);
                }
            }
        }
    };

    // 5. Dependency Injection
    let agent_service = KariAgentService::new(config);
    let grpc_server = Server::builder()
        .add_service(SystemAgentServer::new(agent_service))
        .serve_with_incoming(incoming_stream);

    info!("⚙️ Agent listening on {} [Target UID: {}]", socket_path, uid);

    // 6. Graceful Shutdown Listener
    tokio::select! {
        res = grpc_server => {
            if let Err(e) = res {
                error!("CRITICAL: Server crashed: {}", e);
            }
        }
        _ = signal::ctrl_c() => {
            info!("🛑 Shutdown signal received. Cleaning up socket...");
        }
    }

    // Post-shutdown hygiene
    if Path::new(socket_path).exists() {
        let _ = fs::remove_file(socket_path);
    }
    info!("👋 Agent shutdown complete.");

    Ok(())
}
