use std::collections::BTreeMap;

use crate::commands::types::{SystemCheckCommand, SystemCheckResult};

#[derive(Clone)]
pub struct SystemChecker;

impl SystemChecker {
    pub fn new() -> Self {
        Self
    }

    pub fn run(&self, command: SystemCheckCommand) -> Result<SystemCheckResult, String> {
        if command.server_id.trim().is_empty() {
            return Err("server id is required".to_string());
        }

        let mut services = BTreeMap::new();
        services.insert("nginx".to_string(), "running".to_string());
        services.insert("php-fpm".to_string(), "running".to_string());

        Ok(SystemCheckResult {
            distro: "ubuntu".to_string(),
            version: "22.04".to_string(),
            services,
            firewall_type: "ufw".to_string(),
            firewall_status: "active".to_string(),
        })
    }
}
