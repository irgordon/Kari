// agent/src/sys/build.rs

use crate::sys::traits::BuildManager;
// Note: Adjusted import to match our previous server.rs refactor
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
    ) -> Result<(), String> {
        
        // 🛡️ 1. Zero-Trust Input Validation
        if run_as_user.is_empty() || !run_as_user.chars().all(|c| c.is_ascii_alphanumeric() || c == '-') {
            return Err("SECURITY VIOLATION: Invalid username format".into());
        }

        // 🛡️ 2. Platform Agnostic Shell & Privilege Dropping
        // We pass the build command to `sh -c` to ensure POSIX compliance across all Linux distros.
        let mut child = Command::new("runuser")
            .arg("-u").arg(run_as_user)
            .arg("--")
            .arg("sh").arg("-c").arg(build_command)
            .current_dir(working_dir)
            .envs(env_vars) // Native environment block injection (No shell eval needed)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("Failed to spawn build process: {}", e))?;

        let stdout = child.stdout.take().expect("Failed to open stdout");
        let stderr = child.stderr.take().expect("Failed to open stderr");

        // 🛡️ 3. SLA Backpressure Relief Valve
        // Using `try_send` drops the chunk if the Go Brain/Svelte UI channel is full,
        // preventing the `child` process from blocking on full stdout pipes.
        let tx_out = log_tx.clone();
        let stdout_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stdout).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let chunk = LogChunk { content: format!("[STDOUT] {}\n", line), trace_id: String::new() };
                let _ = tx_out.try_send(Ok(chunk)); 
            }
        });

        let tx_err = log_tx.clone();
        let stderr_task = tokio::spawn(async move {
            let mut reader = BufReader::new(stderr).lines();
            while let Ok(Some(line)) = reader.next_line().await {
                let chunk = LogChunk { content: format!("[STDERR] {}\n", line), trace_id: String::new() };
                let _ = tx_err.try_send(Ok(chunk));
            }
        });

        // Wait for the binary execution to finish
        let status = child.wait().await.map_err(|e| e.to_string())?;

        // 🛡️ 4. Log Draining Guarantee
        // We MUST wait for the spawned buffer readers to finish flushing the pipes
        // before we return and drop the `log_tx` channel.
        let _ = tokio::join!(stdout_task, stderr_task);

        if !status.success() {
            return Err(format!("Build process exited with code: {}", status.code().unwrap_or(-1)));
        }

        Ok(())
    }
}
