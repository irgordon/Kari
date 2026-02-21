// agent/src/sys/ssl.rs

use async_trait::async_trait;
use std::fs as std_fs;
use std::io::Write;
use std::os::unix::fs::{OpenOptionsExt, PermissionsExt};
use std::path::Path;
use tokio::fs as tokio_fs;

use crate::sys::traits::{SslEngine, SslPayload};

// ==============================================================================
// 1. Concrete Implementation (Linux Filesystem)
// ==============================================================================

pub struct LinuxSslEngine {
    ssl_storage_dir: String, 
}

impl LinuxSslEngine {
    pub fn new(ssl_storage_dir: String) -> Self {
        Self { ssl_storage_dir }
    }
}

#[async_trait]
impl SslEngine for LinuxSslEngine {
    async fn install_certificate(&self, payload: SslPayload) -> Result<(), String> {
        
        // 🛡️ 1. Zero-Trust Path Traversal Shield
        // Domain names must only contain alphanumeric characters, hyphens, and dots.
        if payload.domain_name.is_empty() || payload.domain_name.contains("..") || payload.domain_name.contains('/') {
            return Err("SECURITY VIOLATION: Invalid domain name format".into());
        }
        
        let is_valid_domain = payload.domain_name.chars().all(|c| c.is_ascii_alphanumeric() || c == '-' || c == '.');
        if !is_valid_domain {
            return Err("SECURITY VIOLATION: Domain contains illegal characters".into());
        }

        let domain_dir = format!("{}/{}", self.ssl_storage_dir, payload.domain_name);
        let domain_path = Path::new(&domain_dir);

        // 🛡️ 2. Eliminate Directory TOCTOU Race
        tokio_fs::create_dir_all(domain_path)
            .await
            .map_err(|e| format!("Failed to create SSL directory: {}", e))?;
            
        // 🛡️ 3. SLA Reliability: Remove unwrap()
        let mut perms = tokio_fs::metadata(domain_path)
            .await
            .map_err(|e| format!("Failed to read directory metadata: {}", e))?
            .permissions();
        perms.set_mode(0o750);
        tokio_fs::set_permissions(domain_path, perms)
            .await
            .map_err(|e| format!("Failed to secure SSL directory permissions: {}", e))?;

        // 4. Write the Public Certificate (Fullchain)
        let fullchain_path = format!("{}/fullchain.pem", domain_dir);
        tokio_fs::write(&fullchain_path, &payload.fullchain_pem)
            .await
            .map_err(|e| format!("Failed to write fullchain.pem: {}", e))?;
        
        let mut fc_perms = tokio_fs::metadata(&fullchain_path)
            .await
            .map_err(|e| format!("Failed to read fullchain metadata: {}", e))?
            .permissions();
        fc_perms.set_mode(0o644); // Publicly readable
        tokio_fs::set_permissions(&fullchain_path, fc_perms)
            .await
            .map_err(|e| format!("Failed to set fullchain permissions: {}", e))?;

        // 5. Securely Write the Private Key (Zero-Copy + Zero-Race Boundary)
        let privkey_path = format!("{}/privkey.pem", domain_dir);
        
        // 🚨 CRITICAL SECURITY BOUNDARY 🚨
        let write_result = payload.privkey_pem.use_secret(|secret_bytes| {
            // 🛡️ 4. Eliminate the "Briefly World-Readable" vulnerability.
            // We use OpenOptionsExt `.mode(0o600)` to instruct the Linux Kernel 
            // to create the file with strict permissions FROM INCEPTION.
            let mut file = std_fs::OpenOptions::new()
                .write(true)
                .create(true)
                .truncate(true)
                .mode(0o600) // rw-------
                .open(&privkey_path)
                .map_err(|e| format!("Failed to open privkey file securely: {}", e))?;

            file.write_all(secret_bytes)
                .map_err(|e| format!("Failed to write secret bytes: {}", e))?;
            
            // Explicitly sync to ensure data hits disk before we zeroize RAM
            file.sync_all()
                .map_err(|e| format!("Failed to sync privkey to disk: {}", e))?;
                
            Ok::<(), String>(())
        });

        if let Err(e) = write_result {
            // Cleanup on failure
            let _ = std_fs::remove_file(&privkey_path);
            return Err(e);
        }

        Ok(())
    }
}
