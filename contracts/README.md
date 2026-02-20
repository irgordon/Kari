# Contracts

Versioned protobuf contracts for Brain and Agent.

## Generate Code

```bash
cd contracts
buf generate
```

Generated output targets:

- `gen/go`
- `gen/rust`

These generated files should be consumed by:

- Brain gRPC client/server boundary
- Agent gRPC server/client boundary
