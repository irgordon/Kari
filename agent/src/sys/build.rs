use crate::sys::traits::BuildManager;
use crate::server::kari_agent::LogChunk; 
use async_trait::async_trait;
use std::collections::HashMap;
use std::process::Stdio;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::Command;
use tokio::sync::mpsc;
use tonic::Status;
use tracing::{info, warn, error};

pub struct SystemBuildManager;

#[async_trait]
impl BuildManager for SystemBuildManager {
    async fn execute_build(
        &self,
        build_command: &str,
        working_dir: &str,
        run_as_user: &str,
        env_vars: &HashMap<String, String>,
        log_tx: mpsc::Sender<Result<LogChunk, Status>>,
        trace_id: String, 
    ) -> Result<(), String> {
        
        // 🛡️ 1. Identity Validation
        if run_as_user.is_empty() || !run_as_user.chars().all(|c| c.is_ascii_alphanumeric() || c == '-') {
            return Err("SECURITY VIOLATION: Suspicious username format".into());
        }

        // 🛡️ 2. Sandbox Execution
        // Using `runuser` ensures a clean environment drop.
        // `kill_on_drop` is our primary safety net for the Go Brain's context cancellation.
        let mut child = Command::new("runuser")
            .arg("-u").arg(run_as_user)
            .arg("--")
            .arg("sh").arg("-c").arg(build_command)
            .current_dir(working_dir)
            .envs(env_vars)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .kill_on_drop(true) 
            .spawn()
            .map_err(|e| format!("Failed to initiate build process: {}", e))?;

        let stdout = child.stdout.take().ok_or("STDOUT_UNAVAILABLE")?;
        let stderr = child.stderr.take().ok_or("STDERR_UNAVAILABLE")?;

        // 🛡️ 3. Concurrent Telemetry Tasks
        // We use .send().await to respect gRPC backpressure.
        let t_out = trace_id.clone();
        let tx_out = log_tx.clone();
        let stdout_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stdout).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let msg = format!("[OUT] {}\n", line);
                let chunk = LogChunk { content: msg, trace_id: t_out.clone() };
                if tx_out.send(Ok(chunk)).await.is_err() { break; } 
            }
        });

        let t_err = trace_id.clone();
        let tx_err = log_tx.clone();
        let stderr_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stderr).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let msg = format!("[ERR] {}\n", line);
                let chunk = LogChunk { content: msg, trace_id: t_err.clone() };
                if tx_err.send(Ok(chunk)).await.is_err() { break; }
            }
        });

        // 4. Lifecycle Synchronization
        let status = child.wait().await.map_err(|e| e.to_string())?;
        
        // Ensure all log buffers are flushed before returning control to server.rs
        let _ = tokio::join!(stdout_task, stderr_task);

        if !status.success() {
            let exit_desc = match status.code() {
                Some(code) => format!("Exit Code: {}", code),
                None => "Terminated by Signal (OOM/Abort)".to_string(),
            };
            return Err(format!("Build process failed: {}", exit_desc));
        }

        Ok(())
    }
}
