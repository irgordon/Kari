// agent/src/sys/cleanup.rs

use crate::sys::traits::ReleaseManager;
use async_trait::async_trait;
use std::path::{Path, PathBuf};
use tokio::fs;

pub struct SystemReleaseManager;

#[async_trait]
impl ReleaseManager for SystemReleaseManager {
    async fn prune_old_releases(&self, releases_dir: &str, keep_count: usize) -> Result<usize, String> {
        
        // 🛡️ 1. Zero-Trust Path Traversal Shield
        if releases_dir.contains("..") {
            return Err("SECURITY VIOLATION: Path traversal detected in cleanup".into());
        }

        let releases_path = Path::new(releases_dir);
        if !releases_path.exists() {
            return Ok(0); // Nothing to prune
        }

        // 🛡️ 2. Active Release Protection (Symlink Resolution)
        // We must discover exactly what `/current` points to so we NEVER delete the live code.
        // Assuming base_dir/releases and base_dir/current
        let base_dir = releases_path.parent().unwrap_or(releases_path);
        let current_symlink = base_dir.join("current");
        
        // read_link resolves the symlink to its actual target directory
        let active_release_target = fs::read_link(&current_symlink).await.unwrap_or_default();

        let mut entries = match fs::read_dir(releases_path).await {
            Ok(dir) => dir,
            Err(e) => return Err(format!("Failed to read releases directory: {}", e)),
        };

        let mut paths: Vec<PathBuf> = Vec::new();

        while let Ok(Some(entry)) = entries.next_entry().await {
            let path = entry.path();
            let file_name = entry.file_name();
            let name_str = file_name.to_string_lossy();

            // 🛡️ 3. Strict Timestamp Validation
            // Only consider directories that exactly match our 14-digit GitOps 
            // timestamp format (YYYYMMDDHHMMSS).
            let is_valid_timestamp = name_str.len() == 14 && name_str.chars().all(|c| c.is_ascii_digit());

            if path.is_dir() && is_valid_timestamp {
                paths.push(path);
            }
        }

        // Sort paths chronologically (alphabetical sorting works for YYYYMMDDHHMMSS)
        paths.sort();

        let total_releases = paths.len();
        if total_releases <= keep_count {
            return Ok(0); // Under the limit, nothing to prune
        }

        let prune_count = total_releases - keep_count;
        let paths_to_delete = &paths[0..prune_count];

        let mut deleted = 0;

        for path in paths_to_delete {
            // 🛡️ 4. The Final Safety Check
            // Double check this path isn't the active release before nuking it.
            // This guarantees safety even if the user rolled back to an old release.
            if path == &active_release_target {
                tracing::info!("🛡️ Skipping active release directory from pruning: {:?}", path);
                continue;
            }

            if let Err(e) = fs::remove_dir_all(path).await {
                // 🛡️ 5. SLA Observability
                // Use structured logging so the Go Brain and Action Center can capture this warning.
                tracing::warn!("Failed to delete old release {:?}: {}", path, e);
            } else {
                deleted += 1;
            }
        }

        Ok(deleted)
    }
}
