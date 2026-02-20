// agent/src/sys/logs.rs

use crate::sys::traits::LogManager;
use async_trait::async_trait;
use tokio::fs;
use tokio::process::Command;

pub struct LinuxLogManager;

#[async_trait]
impl LogManager for LinuxLogManager {
    async fn configure_logrotate(&self, domain_name: &str, log_dir: &str) -> Result<(), String> {
        let config_path = format!("/etc/logrotate.d/kari-{}", domain_name);

        // Standard Linux logrotate template
        // - daily: Rotate logs every day
        // - rotate 14: Keep 14 days of history
        // - compress: GZIP old logs to save space
        // - delaycompress: Don't compress yesterday's log yet (in case apps are still writing to it)
        // - create 640 root root: Secure permissions for new log files
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

        // 1. Write the configuration file
        fs::write(&config_path, logrotate_config)
            .await
            .map_err(|e| format!("Failed to write logrotate config: {}", e))?;

        // 2. Lock permissions to root only (Secure by Design)
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
