// agent/src/sys/secrets.rs

use secrecy::{ExposeSecret, Secret};
use zeroize::Zeroize;

/// ProviderCredential is an ephemeral, memory-safe wrapper for highly sensitive 
/// data like AWS Route53 API Tokens, Stripe Keys, or RSA Private Keys.
/// 
/// 1. It cannot be accidentally logged (`println!("{:?}", cred)` will output `[REDACTED]`).
/// 2. When the struct goes out of scope, the memory is safely zeroized, 
///    preventing extraction via RAM scraping.
pub struct ProviderCredential {
    token: Secret<Vec<u8>>,
}

impl ProviderCredential {
    /// Wraps raw bytes in a zeroizing Secret
    pub fn new(mut raw_token: Vec<u8>) -> Self {
        let secret = Secret::new(raw_token.clone());
        
        // Physically overwrite the original vector passed into the constructor
        // so no dangling plaintext copies exist in memory.
        raw_token.zeroize(); 
        
        Self { token: secret }
    }

    /// Safely exposes the secret for a fleeting moment to be written to disk or passed
    /// to an API. The caller provides a closure, ensuring the exposed slice cannot 
    /// escape the immediate execution context.
    pub fn use_secret<F, R>(&self, action: F) -> R
    where
        F: FnOnce(&[u8]) -> R,
    {
        // expose_secret() explicitly forces the developer to acknowledge they 
        // are handling plaintext data.
        action(self.token.expose_secret())
    }
}

// Security Audit Note:
// The `secrecy::Secret` type implements `Drop` natively. 
// When a `ProviderCredential` instance is dropped by the Rust compiler, 
// the underlying `Vec<u8>` is automatically zeroized. No manual memory management is required.
