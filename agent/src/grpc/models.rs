#[derive(Clone, Debug)]
pub struct RunSystemCheckRequest {
    pub server_id: String,
}

#[derive(Clone, Debug)]
pub struct RunSystemCheckResponse {
    pub distro: String,
    pub version: String,
    pub firewall_type: String,
    pub firewall_status: String,
}

#[derive(Clone, Debug)]
pub struct ActivateSiteRequest {
    pub site_id: String,
    pub domain: String,
    pub ipv4: String,
    pub ipv6: String,
    pub owner_uid: i32,
}

#[derive(Clone, Debug)]
pub struct ActivateSiteResponse {
    pub ok: bool,
}
