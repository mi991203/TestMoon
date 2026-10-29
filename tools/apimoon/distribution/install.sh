#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
bin_dir="${1:-$HOME/.local/bin}"

for binary in apimoon testmoon-runner; do
  [[ -f "$ROOT/$binary" ]] || { echo "Missing executable: $binary" >&2; exit 1; }
done
mkdir -p "$bin_dir"
install -m 755 "$ROOT/apimoon" "$bin_dir/apimoon"
install -m 755 "$ROOT/testmoon-runner" "$bin_dir/testmoon-runner"
echo "Installed APIMoon and Runner: $bin_dir"
echo 'Add this directory to PATH in your terminal and Agent environment.'
