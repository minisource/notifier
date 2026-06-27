# Generate Protocol Buffer and gRPC code

## Prerequisites

Install protoc compiler and Go plugins:

```bash
# Install protoc
# Windows: Download from https://github.com/protocolbuffers/protobuf/releases
# macOS: brew install protobuf
# Linux: apt install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Generate Code

Run the following command from the project root:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/notifier/v1/notifier.proto
```

Or use the Makefile:

```bash
make proto
```

This will generate:
- `proto/notifier/v1/notifier.pb.go` - Protocol Buffer definitions
- `proto/notifier/v1/notifier_grpc.pb.go` - gRPC service definitions
