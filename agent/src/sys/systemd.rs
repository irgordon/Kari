// agent/src/sys/systemd.rs

use async_trait::async_trait;
use std::collections::HashMap;
use std::os::unix::fs::PermissionsExt;
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
    
    // 🛡️ Added missing SLA methods required by server.rs
    async fn start(&self, service_name: &str) -> Result<(), String>;
    async fn stop(&self, service_name: &str) -> Result<(), String>;
    async fn restart(&self, service_name: &str) -> Result<(), String>;
}

pub struct LinuxSystemdManager {
    systemd_dir: String,
}

impl LinuxSystemdManager {
    pub fn new(systemd_dir: String) -> Self {
        Self { systemd_dir }
    }

    /// Helper trait to ensure consistent error mapping from systemctl
    async fn execute_systemctl(&self, args: &[&str]) -> Result<(), String> {
        let output = Command::new("systemctl")
            .args(args)
            .output()
            .await
            .map_err(|e| format!("Failed to spawn systemctl: {}", e))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(format!("systemctl {} failed: {}", args[0], stderr));
        }
        Ok(())
    }
}

#[async_trait]
impl ServiceManager for LinuxSystemdManager {
    async fn write_unit_file(&self, config: &ServiceConfig) -> Result<(), String> {
        let path = format!("{}/{}.service", self.systemd_dir, config.service_name);
        
        let mut env_strings = String::new();
        for (k, v) in &config.env_vars {
            // 🛡️ 1. Zero-Trust Escape: Prevent Systemd Directive Injection
            // Strip any newlines and escape internal quotes to ensure the value
            // stays securely locked inside the Environment="" boundary.
            let safe_k = k.replace('\n', "");
            let safe_v = v.replace('\n', "").replace('"', "\\\"");
            env_strings.push_str(&format!("Environment=\"{}={}\"\n", safe_k, safe_v));
        }

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

# --- ⚖️ CGroup Resource Limits ---
CPUAccounting=true
CPUQuota=100%
MemoryAccounting=true
MemoryMax=512M
TasksMax=512

# --- 🛡️ Kari Ironclad Security Directives ---
NoNewPrivileges=true
ProtectSystem=full
PrivateTmp=true
ProtectHome=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

# 🛡️ 2. Incomplete Features Added: Network & Device Sandboxing
PrivateDevices=true
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX

[Install]
WantedBy=multi-user.target
"#,
            service_name = config.service_name,
            username = config.username,
            workdir = config.working_directory,
            exec_start = config.start_command,
            env_block = env_strings
        );

        // Write the file temporarily
        fs::write(&path, unit_content)
            .await
            .map_err(|e| format!("Failed to write systemd unit: {}", e))?;

        // 🛡️ 3. Native Kernel Syscalls (No `chmod` subprocess)
        let mut perms = tokio::fs::metadata(&path)
            .await
            .map_err(|e| e.to_string())?
            .permissions();
        perms.set_mode(0o644);
        tokio::fs::set_permissions(&path, perms)
            .await
            .map_err(|e| e.to_string())?;

        Ok(())
    }

    async fn reload_daemon(&self) -> Result<(), String> {
        self.execute_systemctl(&["daemon-reload"]).await
    }

    async fn enable_and_start(&self, service_name: &str) -> Result<(), String> {
        self.execute_systemctl(&["enable", "--now", service_name]).await
    }

    // 🛡️ 4. Fulfill the SLA Trait Contract
    async fn start(&self, service_name: &str) -> Result<(), String> {
        self.execute_systemctl(&["start", service_name]).await
    }

    async fn stop(&self, service_name: &str) -> Result<(), String> {
        self.execute_systemctl(&["stop", service_name]).await
    }

    async fn restart(&self, service_name: &str) -> Result<(), String> {
        self.execute_systemctl(&["restart", service_name]).await
    }
}
