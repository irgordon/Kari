// agent/src/config.rs

use std::env;

#[derive(Clone, Debug)]
pub struct AgentConfig {
    pub web_root: String,
    pub systemd_dir: String,
    pub logrotate_dir: String,
}

impl AgentConfig {
    pub fn load() -> Self {
        Self {
            web_root: env::var("KARI_WEB_ROOT").unwrap_or_else(|_| "/var/www".to_string()),
            systemd_dir: env::var("KARI_SYSTEMD_DIR").unwrap_or_else(|_| "/etc/systemd/system".to_string()),
            logrotate_dir: env::var("KARI_LOGROTATE_DIR").unwrap_or_else(|_| "/etc/logrotate.d".to_string()),
        }
    }
}
