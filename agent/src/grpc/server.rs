use crate::commands::site::SiteActivator;
use crate::commands::system_check::SystemChecker;
use crate::grpc::handlers::AgentHandlers;

pub struct GrpcServer {
    handlers: AgentHandlers,
}

impl GrpcServer {
    pub fn new(system_checker: SystemChecker, site_activator: SiteActivator) -> Self {
        Self {
            handlers: AgentHandlers::new(system_checker, site_activator),
        }
    }

    pub fn start(&self) -> Result<(), String> {
        let _ = &self.handlers;
        Ok(())
    }
}
