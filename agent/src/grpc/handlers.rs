use crate::commands::site::SiteActivator;
use crate::commands::system_check::SystemChecker;
use crate::grpc::models::{
    ActivateSiteRequest, ActivateSiteResponse, RunSystemCheckRequest, RunSystemCheckResponse,
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
        self.system_checker.run(request.server_id)
    }

    pub fn activate_site(
        &self,
        request: ActivateSiteRequest,
    ) -> Result<ActivateSiteResponse, String> {
        self.site_activator.activate(request)
    }
}
