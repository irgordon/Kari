use crate::commands::site::SiteActivator;
use crate::commands::system_check::SystemChecker;
use crate::grpc::server::GrpcServer;

pub fn run() -> Result<(), String> {
    let system_checker = SystemChecker::new();
    let site_activator = SiteActivator::new();
    let server = GrpcServer::new(system_checker, site_activator);
    server.start()
}
