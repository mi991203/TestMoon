#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
for metadata in VERSION COMMIT REPOSITORY; do
  [[ -f "$ROOT/$metadata" ]] || { echo "Missing release metadata: $metadata" >&2; exit 1; }
done
version="$(cat "$ROOT/VERSION")"
commit="$(cat "$ROOT/COMMIT")"
repository="$(cat "$ROOT/REPOSITORY")"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]] || { echo 'Invalid release version' >&2; exit 1; }
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo 'Invalid source commit' >&2; exit 1; }
[[ "$repository" =~ ^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$ ]] || { echo 'Invalid source repository' >&2; exit 1; }
command -v docker >/dev/null || { echo 'Docker is required.' >&2; exit 1; }
docker info >/dev/null

temporary=""
trap 'if [[ -n "$temporary" ]]; then rm -rf "$temporary"; fi' EXIT
source_dir="${1:-}"
if [[ -z "$source_dir" ]]; then
  command -v curl >/dev/null || { echo 'curl is required to download the source.' >&2; exit 1; }
  temporary="$(mktemp -d "${TMPDIR:-/tmp}/apimoon-worker.XXXXXX")"
  curl --fail --location --retry 3 "https://github.com/$repository/archive/$commit.tar.gz" -o "$temporary/source.tar.gz"
  mkdir "$temporary/source"
  tar -xzf "$temporary/source.tar.gz" --strip-components=1 -C "$temporary/source"
  source_dir="$temporary/source"
fi
[[ -f "$source_dir/deploy/runner/worker.Dockerfile" ]] || { echo 'TestMoon Worker Dockerfile not found.' >&2; exit 1; }
if [[ -z "$temporary" ]]; then
  # An explicit checkout is useful for private repositories and offline source access.
  actual="$(git -C "$source_dir" rev-parse HEAD)"
  [[ "$actual" == "$commit" ]] || { echo "Source checkout must be at release commit $commit" >&2; exit 1; }
  [[ -z "$(git -C "$source_dir" status --porcelain --untracked-files=normal)" ]] || { echo 'Source checkout has local changes; use a clean release checkout.' >&2; exit 1; }
fi
image="${TESTMOON_WORKER_IMAGE:-testmoon/worker:${version#v}}"
docker build -f "$source_dir/deploy/runner/worker.Dockerfile" -t "$image" "$source_dir"
echo "Built local Worker image: $image"
echo "Set TESTMOON_WORKER_IMAGE=$image before starting testmoon-runner."
