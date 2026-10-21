#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$PWD/.tools/bin:$PATH"
for tool in protoc protoc-gen-go protoc-gen-go-grpc protoc-gen-grpc-gateway protoc-gen-openapiv2; do command -v "$tool" >/dev/null || { echo "Missing $tool; run make tools" >&2; exit 1; }; done
mapfile_compat=()
while IFS= read -r file; do mapfile_compat+=("$file"); done < <(find protocols -name '*.proto' | sort)
go build -o .tools/bin/protoc-gen-typescript ./tools/protoc-gen-typescript
protoc -I . -I third_party/proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative --openapiv2_out=. --typescript_out=. "${mapfile_compat[@]}"
