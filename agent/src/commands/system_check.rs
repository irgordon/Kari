use crate::grpc::models::RunSystemCheckResponse;

#[derive(Clone)]
pub struct SystemChecker;

impl SystemChecker {
    pub fn new() -> Self {
        Self
    }

    pub fn run(&self, _server_id: String) -> Result<RunSystemCheckResponse, String> {
        Ok(RunSystemCheckResponse {
            distro: "ubuntu".to_string(),
            version: "22.04".to_string(),
            firewall_type: "ufw".to_string(),
            firewall_status: "active".to_string(),
        })
    }
}
