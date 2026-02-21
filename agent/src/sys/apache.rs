use async_trait::async_trait;
use tokio::fs;
use tokio::process::Command;
use std::path::Path;

#[async_trait]
pub trait ProxyManager: Send + Sync {
    async fn create_vhost(&self, domain: &str, target_port: u16) -> Result<(), String>;
    async fn remove_vhost(&self, domain: &str) -> Result<(), String>;
}

pub struct ApacheManager {
    sites_available: String,
    sites_enabled: String,
}

impl ApacheManager {
    pub fn new() -> Self {
        Self {
            sites_available: "/etc/apache2/sites-available".to_string(),
            sites_enabled: "/etc/apache2/sites-enabled".to_string(),
        }
    }

    async fn test_and_reload(&self) -> Result<(), String> {
        // 🛡️ SLA: Never reload a broken config
        let check = Command::new("apache2ctl")
            .arg("configtest")
            .output()
            .await
            .map_err(|e| e.to_string())?;

        if !check.status.success() {
            return Err("Apache config test failed. Rolling back.".into());
        }

        Command::new("systemctl")
            .args(["reload", "apache2"])
            .output()
            .await
            .map_err(|e| e.to_string())?;

        Ok(())
    }
}

#[async_trait]
impl ProxyManager for ApacheManager {
    async fn create_vhost(&self, domain: &str, target_port: u16) -> Result<(), String> {
        let config_path = format!("{}/{}.conf", self.sites_available, domain);
        let enabled_link = format!("{}/{}.conf", self.sites_enabled, domain);

        // 🛡️ Zero-Trust: Hardened VHost Template
        let content = format!(
            r#"<VirtualHost *:80>
    ServerName {domain}

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:{target_port}/
    ProxyPassReverse / http://127.0.0.1:{target_port}/

    # --- 🛡️ Security Headers ---
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-Frame-Options "SAMEORIGIN"
    
    ErrorLog ${{APACHE_LOG_DIR}}/{domain}_error.log
    CustomLog ${{APACHE_LOG_DIR}}/{domain}_access.log combined
</VirtualHost>"#,
            domain = domain,
            target_port = target_port
        );

        fs::write(&config_path, content).await.map_err(|e| e.to_string())?;

        // Enable the site via symlink if it doesn't exist
        if !Path::new(&enabled_link).exists() {
            fs::symlink(&config_path, &enabled_link).await.map_err(|e| e.to_string())?;
        }

        self.test_and_reload().await
    }

    async fn remove_vhost(&self, domain: &str) -> Result<(), String> {
        let enabled_link = format!("{}/{}.conf", self.sites_enabled, domain);
        let config_path = format!("{}/{}.conf", self.sites_available, domain);

        let _ = fs::remove_file(enabled_link).await;
        let _ = fs::remove_file(config_path).await;

        self.test_and_reload().await
    }
}
