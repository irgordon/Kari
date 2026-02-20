# System Check

System check is run during server onboarding.

## Brain Flow

1. API receives `POST /v1/servers/onboard`.
2. Brain validates request.
3. Brain calls `SystemChecker.RunSystemCheck`.
4. Brain returns normalized report.

## Current Scaffold Output

```json
{
  "distro": "ubuntu",
  "version": "22.04",
  "services": {
    "nginx": "running",
    "php-fpm": "running"
  },
  "firewall_type": "ufw",
  "firewall_status": "active"
}
```

## Next Implementation Steps

- Replace in-memory checker with gRPC client implementation.
- Add package presence and open-port checks.
- Persist onboarding reports for audit and troubleshooting.
