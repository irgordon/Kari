use crate::sys::traits::BuildManager;
use crate::server::kari_agent::LogChunk; 
use async_trait::async_trait;
use std::collections::HashMap;
use std::process::Stdio;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::Command;
use tokio::sync::mpsc;
use tonic::Status;

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
        trace_id: String, // 🛡️ SLA: Passed from server.rs for correlation
    ) -> Result<(), String> {
        
        // 🛡️ 1. Zero-Trust Validation
        if run_as_user.is_empty() || !run_as_user.chars().all(|c| c.is_ascii_alphanumeric() || c == '-') {
            return Err("SECURITY VIOLATION: Invalid username format".into());
        }

        // 🛡️ 2. Automatic Cleanup
        // kill_on_drop(true) ensures the build process terminates if the gRPC context is cancelled.
        let mut cmd = Command::new("runuser");
        cmd.arg("-u").arg(run_as_user)
           .arg("--")
           .arg("sh").arg("-c").arg(build_command)
           .current_dir(working_dir)
           .envs(env_vars)
           .stdout(Stdio::piped())
           .stderr(Stdio::piped())
           .kill_on_drop(true); 

        let mut child = cmd.spawn()
            .map_err(|e| format!("SLA Failure: Build spawn error: {}", e))?;

        let stdout = child.stdout.take().ok_or("Failed to open stdout")?;
        let stderr = child.stderr.take().ok_or("Failed to open stderr")?;

        // 🛡️ 3. Structured Telemetry Tasks
        let t_out = trace_id.clone();
        let tx_out = log_tx.clone();
        let stdout_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stdout).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let chunk = LogChunk { content: format!("{}\n", line), trace_id: t_out.clone() };
                // Using await here respects backpressure; try_send would drop lines.
                if tx_out.send(Ok(chunk)).await.is_err() { break; } 
            }
        });

        let t_err = trace_id.clone();
        let tx_err = log_tx.clone();
        let stderr_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stderr).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let chunk = LogChunk { content: format!("ERR: {}\n", line), trace_id: t_err.clone() };
                if tx_err.send(Ok(chunk)).await.is_err() { break; }
            }
        });

        // 4. Synchronization
        let status = child.wait().await.map_err(|e| e.to_string())?;
        
        // Ensure all logs are flushed before returning
        let _ = tokio::join!(stdout_task, stderr_task);

        if !status.success() {
            return Err(format!("Build exited with status: {}", status));
        }

        Ok(())
    }
}
