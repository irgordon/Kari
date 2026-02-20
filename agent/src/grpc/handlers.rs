use crate::commands::site::SiteActivator;
use crate::commands::system_check::SystemChecker;
use crate::commands::types::{
    SiteActivationAction as CommandAction, SiteActivationCommand, SystemCheckCommand,
};
use crate::grpc::models::{
    ActivateSiteRequest, ActivateSiteResponse, RunSystemCheckRequest, RunSystemCheckResponse,
    SiteActivationAction,
};

pub struct AgentHandlers {
    system_checker: SystemChecker,
    site_activator: SiteActivator,
}

impl AgentHandlers {
    pub fn new(system_checker: SystemChecker, site_activator: SiteActivator) -> Self {
        Self {
            system_checker,
            site_activator,
        }
    }

    pub fn run_system_check(
        &self,
        request: RunSystemCheckRequest,
    ) -> Result<RunSystemCheckResponse, String> {
        let result = self
            .system_checker
            .run(SystemCheckCommand { server_id: request.server_id })?;
        Ok(RunSystemCheckResponse {
            distro: result.distro,
            version: result.version,
            services: result.services,
            firewall_type: result.firewall_type,
            firewall_status: result.firewall_status,
        })
    }

    pub fn activate_site(
        &self,
        request: ActivateSiteRequest,
    ) -> Result<ActivateSiteResponse, String> {
        let result = self.site_activator.activate(SiteActivationCommand {
            site_id: request.site_id,
            domain: request.domain,
            ipv4: request.ipv4,
            ipv6: request.ipv6,
            owner_uid: request.owner_uid,
            action: map_action(request.action),
        })?;

        Ok(ActivateSiteResponse { ok: result.ok })
    }
}

fn map_action(action: SiteActivationAction) -> CommandAction {
    match action {
        SiteActivationAction::CreateSystemUser => CommandAction::CreateSystemUser,
        SiteActivationAction::ApplyHttpVhost => CommandAction::ApplyHttpVhost,
        SiteActivationAction::IssueCertificate => CommandAction::IssueCertificate,
        SiteActivationAction::ApplyHttpsVhost => CommandAction::ApplyHttpsVhost,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn run_system_check_maps_successfully() {
        let handlers = AgentHandlers::new(SystemChecker::new(), SiteActivator::new());
        let response = handlers
            .run_system_check(RunSystemCheckRequest {
                server_id: "srv-1".to_string(),
            })
            .expect("system check should succeed");

        assert_eq!(response.distro, "ubuntu");
    }

    #[test]
    fn activate_site_rejects_invalid_input() {
        let handlers = AgentHandlers::new(SystemChecker::new(), SiteActivator::new());
        let result = handlers.activate_site(ActivateSiteRequest {
            site_id: "".to_string(),
            domain: "example.com".to_string(),
            ipv4: "1.2.3.4".to_string(),
            ipv6: "".to_string(),
            owner_uid: 1001,
            action: SiteActivationAction::CreateSystemUser,
        });

        assert!(result.is_err());
    }
}
