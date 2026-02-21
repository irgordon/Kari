# 🔌 The Muscle API (gRPC Schema)

The communication layer between the **Go Brain** and the **Rust Muscle** strictly follows this Protocol Buffer schema. Communication occurs over a local Unix Domain Socket (UDS) to eliminate network latency and external attack vectors.

## `proto/kari/v1/agent.proto`

```protobuf
syntax = "proto3";
package kari.v1;
option go_package = "kari/api/internal/grpc/rustagent";

service SystemAgent {
    // 🛡️ SSL & Let's Encrypt Management
    rpc ManageSslChallenge (ChallengeRequest) returns (ChallengeResponse);
    rpc InstallCertificate (SslInstallRequest) returns (SslInstallResponse);

    // 📦 Container & Jail Management
    rpc ProvisionJail (JailRequest) returns (JailResponse);
    rpc StreamLogs (LogRequest) returns (stream LogChunk);
}

// --- Messages ---

enum ChallengeAction {
    PRESENT = 0;
    CLEANUP = 1;
}

message ChallengeRequest {
    ChallengeAction action = 1;
    string domain = 2;
    string token = 3;     // HTTP-01 ACME Token
    string key_auth = 4;  // The expected response string
}

message ChallengeResponse {
    bool success = 1;
    string error_message = 2;
}

message SslInstallRequest {
    string domain_name = 1;
    bytes fullchain_pem = 2;
    bytes privkey_pem = 3; // Agent must zeroize this buffer after writing!
}

message SslInstallResponse {
    bool success = 1;
    string deploy_path = 2;
}
