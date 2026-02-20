// agent/src/sys/systemd.rs

use async_trait::async_trait;
use std::collections::HashMap;
use tokio::fs;
use tokio::process::Command;

pub struct ServiceConfig {
    pub service_name: String,
    pub username: String,
    pub working_directory: String,
    pub start_command: String,
    pub env_vars: HashMap<String, String>,
}

#[async_trait]
pub trait ServiceManager: Send + Sync {
    async fn write_unit_file(&self, config: &ServiceConfig) -> Result<(), String>;
    async fn reload_daemon(&self) -> Result<(), String>;
    async fn enable_and_start(&self, service_name: &str) -> Result<(), String>;
}

pub struct LinuxSystemdManager;

#[async_trait]
impl ServiceManager for LinuxSystemdManager {
    async fn write_unit_file(&self, config: &ServiceConfig) -> Result<(), String> {
        let path = format!("/etc/systemd/system/{}.service", config.service_name);
        
        let mut env_strings = String::new();
        for (k, v) in &config.env_vars {
            env_strings.push_str(&format!("Environment=\"{}={}\"\n", k, v));
        }

        // This template is pure 2026 security.
        let unit_content = format!(
            r#"[Unit]
Description=Kari Managed App: {service_name}
After=network.target

[Service]
Type=simple
User={username}
Group={username}
WorkingDirectory={workdir}
ExecStart={exec_start}
{env_block}
Restart=always
RestartSec=3

# --- 🛡️ Kari Ironclad Security Directives ---
# Prevent the app from gaining new privileges (no sudo escalations)
NoNewPrivileges=true

# Mount /usr, /boot, and /etc as Read-Only
ProtectSystem=full

# Give the app its own isolated /tmp folder that deletes on exit
PrivateTmp=true

# Hide home directories of other users on the server
ProtectHome=true

# Restrict kernel tuning and module loading
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
"#,
            service_name = config.service_name,
            username = config.username,
            workdir = config.working_directory,
            exec_start = config.start_command,
            env_block = env_strings
        );

        fs::write(&path, unit_content)
            .await
            .map_err(|e| format!("Failed to write systemd unit: {}", e))?;

        // Lock permissions of the unit file to root only
        Command::new("chmod").args(["644", &path]).output().await.map_err(|e| e.to_string())?;

        Ok(())
    }

    async fn reload_daemon(&self) -> Result<(), String> {
        let output = Command::new("systemctl").arg("daemon-reload").output().await.map_err(|e| e.to_string())?;
        if !output.status.success() {
            return Err("Failed to reload systemd daemon".into());
        }
        Ok(())
    }

    async fn enable_and_start(&self, service_name: &str) -> Result<(), String> {
        Command::new("systemctl").args(["enable", "--now", service_name]).output().await.map_err(|e| e.to_string())?;
        Ok(())
    }
}
