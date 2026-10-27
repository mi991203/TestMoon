#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
export PATH="$PWD/.tools/bin:$PATH"
for tool in protoc protoc-gen-go protoc-gen-go-grpc protoc-gen-grpc-gateway protoc-gen-openapiv2; do
  command -v "$tool" >/dev/null || { echo "Missing $tool; run make tools" >&2; exit 1; }
done
protoc -I . -I third_party/proto \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  --openapiv2_out=. \
  protocols/fixtures/school_management/v1/service.proto
