#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
version="${1:-}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]] || { echo 'Usage: package-apimoon.sh vX.Y.Z [output-directory]' >&2; exit 2; }
output="${2:-$ROOT/.local/apimoon-release/$version}"
mkdir -p "$output"
output="$(cd "$output" && pwd)"
[[ ! -e "$output/SHA256SUMS" ]] || { echo 'Output directory already contains a release; choose a new directory.' >&2; exit 1; }
commit="$(git rev-parse HEAD)"
repository="${GITHUB_REPOSITORY:-mi991203/TestMoon}"
[[ "$repository" =~ ^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$ ]] || { echo 'Invalid source repository' >&2; exit 1; }
date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
metadata=github.com/mi991203/TestMoon/libs/go/buildinfo
flags="-s -w -X $metadata.Version=$version -X $metadata.Commit=$commit -X $metadata.Date=$date"
temporary="$(mktemp -d "${TMPDIR:-/tmp}/apimoon-release.XXXXXX")"
trap 'rm -rf "$temporary"' EXIT
targets="${APIMOON_RELEASE_TARGETS:-darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64}"
for target in $targets; do
  case "$target" in darwin/amd64|darwin/arm64|linux/amd64|linux/arm64|windows/amd64|windows/arm64);; *) echo "Unsupported target: $target" >&2; exit 2;; esac
done
assets=()
for target in $targets; do
  goos="${target%/*}"
  goarch="${target#*/}"
  name="apimoon_${version}_${goos}_${goarch}"
  if [[ "$goos" == windows ]]; then
    runner_name="testmoon-runner_${version}_${goos}_${goarch}.exe"
    [[ ! -e "$output/$name.exe" && ! -e "$output/$runner_name" ]] || { echo "Windows executable already exists for $target; choose a new directory." >&2; exit 1; }
    echo "Building APIMoon and Runner for $target"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" "${GO:-go}" build -trimpath -ldflags "$flags" -o "$temporary/$name.exe" ./tools/apimoon/cmd
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" "${GO:-go}" build -trimpath -ldflags "$flags" -o "$temporary/$runner_name" ./backend/testing/runner/cmd/runner
    mv "$temporary/$name.exe" "$temporary/$runner_name" "$output/"
    assets+=("$name.exe" "$runner_name")
    continue
  fi
  archive="$output/$name.tar.gz"
  [[ ! -e "$archive" ]] || { echo "Archive already exists: $archive" >&2; exit 1; }
  package="$temporary/$name"
  mkdir -p "$package"
  echo "Building APIMoon and Runner for $target"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" "${GO:-go}" build -trimpath -ldflags "$flags" -o "$package/apimoon" ./tools/apimoon/cmd
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" "${GO:-go}" build -trimpath -ldflags "$flags" -o "$package/testmoon-runner" ./backend/testing/runner/cmd/runner
  cp tools/apimoon/distribution/install.sh tools/apimoon/distribution/build-worker.sh tools/apimoon/distribution/README.md "$package/"
  printf '%s\n' "$version" > "$package/VERSION"
  printf '%s\n' "$commit" > "$package/COMMIT"
  printf '%s\n' "$repository" > "$package/REPOSITORY"
  chmod 755 "$package/apimoon" "$package/testmoon-runner" "$package/install.sh" "$package/build-worker.sh"
  COPYFILE_DISABLE=1 tar -czf "$archive" -C "$temporary" "$name"
  assets+=("$name.tar.gz")
done
(
  cd "$output"
  if command -v sha256sum >/dev/null; then
    sha256sum "${assets[@]}" > SHA256SUMS
  else
    shasum -a 256 "${assets[@]}" > SHA256SUMS
  fi
)
echo "Release assets and checksums: $output"
