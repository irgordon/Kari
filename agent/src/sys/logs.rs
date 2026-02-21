// agent/src/sys/logs.rs

use crate::sys::traits::LogManager;
use async_trait::async_trait;
use tokio::fs;
use tokio::process::Command;

pub struct LinuxLogManager {
    logrotate_dir: String, // Injected path
}

impl LinuxLogManager {
    pub fn new(logrotate_dir: String) -> Self {
        Self { logrotate_dir }
    }
}

#[async_trait]
impl LogManager for LinuxLogManager {
    async fn configure_logrotate(&self, domain_name: &str, log_dir: &str) -> Result<(), String> {
        // INJECTED: Dynamically construct the config path
        let config_path = format!("{}/kari-{}", self.logrotate_dir, domain_name);

        let logrotate_config = format!(
            r#"{log_dir}/*.log {{
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 640 root root
    sharedscripts
    postrotate
        if [ -f /var/run/nginx.pid ]; then
            kill -USR1 `cat /var/run/nginx.pid`
        fi
    endscript
}}
"#,
            log_dir = log_dir
        );

        fs::write(&config_path, logrotate_config)
            .await
            .map_err(|e| format!("Failed to write logrotate config: {}", e))?;

        let chmod_out = Command::new("chmod")
            .args(["644", &config_path])
            .output()
            .await
            .map_err(|e| e.to_string())?;

        if !chmod_out.status.success() {
            return Err("Failed to secure logrotate config permissions".into());
        }

        Ok(())
    }
}
