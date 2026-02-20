// agent/src/sys/jail.rs

use async_trait::async_trait;
use tokio::process::Command;
use std::path::Path;

#[async_trait]
pub trait JailManager: Send + Sync {
    /// Creates a unique Linux user with no login shell
    async fn provision_app_user(&self, username: &str) -> Result<(), String>;
    
    /// Locks down a directory so only the app user (and root) can access it
    async fn secure_directory(&self, path: &str, username: &str) -> Result<(), String>;
}

pub struct LinuxJailManager;

#[async_trait]
impl JailManager for LinuxJailManager {
    async fn provision_app_user(&self, username: &str) -> Result<(), String> {
        // 1. Check if user already exists
        let check = Command::new("id").arg("-u").arg(username).output().await;
        if let Ok(output) = check {
            if output.status.success() {
                return Ok(()); // User exists, idempotent success
            }
        }

        // 2. Create an unprivileged system user with NO login shell (/bin/false)
        let output = Command::new("useradd")
            .args(["--system", "--shell", "/bin/false", username])
            .output()
            .await
            .map_err(|e| format!("Failed to execute useradd: {}", e))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(format!("Failed to create user {}: {}", username, stderr));
        }

        Ok(())
    }

    async fn secure_directory(&self, path: &str, username: &str) -> Result<(), String> {
        if !Path::new(path).exists() {
            tokio::fs::create_dir_all(path)
                .await
                .map_err(|e| format!("Failed to create app directory: {}", e))?;
        }

        // Chown to the unprivileged user
        let chown_out = Command::new("chown")
            .args(["-R", &format!("{}:{}", username, username), path])
            .output()
            .await
            .map_err(|e| e.to_string())?;

        if !chown_out.status.success() {
            return Err("Failed to chown application directory".into());
        }

        // Chmod to 750: Owner can read/write/execute, Group can read/execute, Others get NOTHING.
        let chmod_out = Command::new("chmod")
            .args(["750", path])
            .output()
            .await
            .map_err(|e| e.to_string())?;

        if !chmod_out.status.success() {
            return Err("Failed to chmod application directory".into());
        }

        Ok(())
    }
}
