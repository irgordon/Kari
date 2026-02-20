# Agent API

Versioned contract: `/contracts/agent/v1/agent.proto`

## Transport

- gRPC
- mTLS TCP or UNIX socket

## Service

- `AgentService.RunSystemCheck`
- `AgentService.ActivateSite`

## RPC: RunSystemCheck

Request:

- `server_id` string

Response:

- `distro` string
- `version` string
- `services` map<string,string>
- `firewall_type` string
- `firewall_status` string

## RPC: ActivateSite

Request:

- `site_id` string
- `domain` string
- `ipv4` string
- `ipv6` string
- `owner_uid` int32

Response:

- `ok` bool

## Brain Mapping

- Brain API `POST /v1/sites/activate` maps to Agent `ActivateSite`.
- Brain API `POST /v1/servers/onboard` maps to Agent `RunSystemCheck`.
- Brain handles policy and validation. Agent executes host operations.
