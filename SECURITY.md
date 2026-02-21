# Security Policy

## 🛡️ Our Commitment
Karı is built on a **Zero-Trust** and **SLA-First** architecture. We prioritize memory safety, cryptographic integrity, and process isolation. However, no software is perfect. We value the work of security researchers and the community in helping us maintain a hardened ecosystem.

## 🟢 Supported Versions
We currently provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| v0.0.3  | ✅ Supported        |
| < v0.0.3| ❌ Discontinued     |

## 📨 Reporting a Vulnerability
**Do not open a GitHub Issue for security vulnerabilities.**

If you discover a security hole, please report it via email to:
**[security@kariapp.dev](mailto:security@kariapp.dev)**

### What we want to see:
* **A clear description** of the vulnerability.
* **Steps to reproduce** (PoC scripts or curl commands are highly appreciated).
* **The potential impact** (e.g., privilege escalation, data leak, RCE).
* **Your contact information** if you wish to be credited.

## 🤝 Our Response Process
1.  **Acknowledgment**: We will acknowledge your report within **24 hours**.
2.  **Validation**: Our team will perform a manual code review and attempt to reproduce the issue.
3.  **Fixing**: We aim to have a patch or mitigation ready within **7 days** for critical issues.
4.  **Disclosure**: We follow a "Coordinated Disclosure" model. We ask that you do not share the vulnerability publicly until a patch is released and we have agreed on a disclosure date.

## 🚫 Out of Scope
* DDoS attacks against Karı instances.
* Social engineering attacks against our maintainers.
* Vulnerabilities in third-party dependencies (though we still want to know so we can update our lockfiles).

## 🎖️ Attribution
Unless you choose to remain anonymous, we will credit you in our `CHANGELOG.md` and project announcements for your contribution to the safety of the engine.
