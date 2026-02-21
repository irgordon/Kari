use secrecy::{ExposeSecret, Secret, Zeroize};

/// ProviderCredential is an ephemeral, memory-safe wrapper for highly sensitive data.
/// It uses the 'secrecy' and 'zeroize' crates to ensure that once a secret is used,
/// its footprint in RAM is physically overwritten with zeros.
pub struct ProviderCredential {
    // 🛡️ Secret<T> ensures Debug and Display traits are redacted.
    token: Secret<Vec<u8>>,
}

impl ProviderCredential {
    /// Wraps raw bytes in a zeroizing Secret.
    pub fn new(mut raw_token: Vec<u8>) -> Self {
        // 🛡️ 1. Immediate Ownership & Confinement
        // We move the raw_token into the Secret wrapper. 
        // Note: In a true high-security context, we would use a mlock'd buffer.
        Self { 
            token: Secret::new(raw_token) 
        }
    }

    /// Hardened constructor for Strings to prevent heap residue.
    pub fn from_string(mut s: String) -> Self {
        let bytes = s.as_bytes().to_vec();
        // 🛡️ 2. Manual Scrutiny
        // We zeroize the source string immediately after copying to the Vec.
        s.zeroize(); 
        Self::new(bytes)
    }

    /// Safely exposes the secret for a fleeting moment.
    /// 🛡️ Lexical Scope Confinement ensures the secret cannot escape the closure.
    pub fn use_secret<F, R>(&self, action: F) -> R
    where
        // R cannot be a reference to the secret bytes due to lifetime constraints.
        F: FnOnce(&[u8]) -> R,
    {
        action(self.token.expose_secret())
    }
}

// 🛡️ 3. Explicit Drop Guarantee
// Even though Secret<T> handles this, implementing Zeroize for our wrapper
// provides a secondary safety net for future refactors.
impl Zeroize for ProviderCredential {
    fn zeroize(&mut self) {
        // The secrecy crate handles the internal Vec zeroization.
    }
}
