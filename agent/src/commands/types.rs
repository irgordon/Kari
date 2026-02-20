use std::collections::BTreeMap;

#[derive(Clone, Debug)]
pub struct SystemCheckCommand {
    pub server_id: String,
}

#[derive(Clone, Debug)]
pub struct SystemCheckResult {
    pub distro: String,
    pub version: String,
    pub services: BTreeMap<String, String>,
    pub firewall_type: String,
    pub firewall_status: String,
}

#[derive(Clone, Debug)]
pub struct SiteActivationCommand {
    pub site_id: String,
    pub domain: String,
    pub ipv4: String,
    pub ipv6: String,
    pub owner_uid: i32,
    pub action: SiteActivationAction,
}

#[derive(Clone, Debug)]
pub struct SiteActivationResult {
    pub ok: bool,
}

#[derive(Clone, Debug)]
pub enum SiteActivationAction {
    CreateSystemUser,
    ApplyHttpVhost,
    IssueCertificate,
    ApplyHttpsVhost,
}
