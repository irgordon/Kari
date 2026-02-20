use crate::commands::types::{SiteActivationAction, SiteActivationCommand, SiteActivationResult};

#[derive(Clone)]
pub struct SiteActivator;

impl SiteActivator {
    pub fn new() -> Self {
        Self
    }

    pub fn activate(&self, command: SiteActivationCommand) -> Result<SiteActivationResult, String> {
        if command.site_id.trim().is_empty() {
            return Err("site id is required".to_string());
        }
        if command.domain.trim().is_empty() {
            return Err("domain is required".to_string());
        }
        if command.owner_uid <= 0 {
            return Err("owner uid must be positive".to_string());
        }

        self.run_action(command.action)?;
        Ok(SiteActivationResult { ok: true })
    }

    fn run_action(&self, action: SiteActivationAction) -> Result<(), String> {
        match action {
            SiteActivationAction::CreateSystemUser => Ok(()),
            SiteActivationAction::ApplyHttpVhost => Ok(()),
            SiteActivationAction::IssueCertificate => Ok(()),
            SiteActivationAction::ApplyHttpsVhost => Ok(()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::types::SiteActivationCommand;

    #[test]
    fn activate_accepts_valid_command() {
        let activator = SiteActivator::new();
        let result = activator.activate(SiteActivationCommand {
            site_id: "site-1".to_string(),
            domain: "example.com".to_string(),
            ipv4: "1.2.3.4".to_string(),
            ipv6: "".to_string(),
            owner_uid: 1001,
            action: SiteActivationAction::ApplyHttpVhost,
        });

        assert!(result.is_ok());
    }
}
