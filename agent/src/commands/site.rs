use crate::grpc::models::{ActivateSiteRequest, ActivateSiteResponse};

#[derive(Clone)]
pub struct SiteActivator;

impl SiteActivator {
    pub fn new() -> Self {
        Self
    }

    pub fn activate(&self, request: ActivateSiteRequest) -> Result<ActivateSiteResponse, String> {
        if request.site_id.trim().is_empty() {
            return Err("site id is required".to_string());
        }
        if request.domain.trim().is_empty() {
            return Err("domain is required".to_string());
        }
        if request.owner_uid <= 0 {
            return Err("owner uid must be positive".to_string());
        }
        Ok(ActivateSiteResponse { ok: true })
    }
}
