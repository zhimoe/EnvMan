#!/usr/bin/env bash
# Package committed sources and reproducible vendored dependencies for offline builds.
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
project_dir="$(cd -- "$script_dir/.." && pwd)"
cd "$project_dir"

commit="$(git rev-parse HEAD)"
short_commit="$(git rev-parse --short "$commit")"
output_dir="$project_dir/build/source"
archive_name="EnvMan-source-$short_commit"
staging_dir="$(mktemp -d "${TMPDIR:-/tmp}/envman-source.XXXXXX")"
trap 'rm -rf -- "$staging_dir"' EXIT
mkdir -p "$staging_dir/$archive_name" "$output_dir"

git archive "$commit" | tar -xf - -C "$staging_dir/$archive_name"
(
  cd "$staging_dir/$archive_name"
  GOWORK=off go mod vendor
)
tar -czf "$staging_dir/$archive_name.tar.gz" -C "$staging_dir" "$archive_name"
mv -- "$staging_dir/$archive_name.tar.gz" "$output_dir/$archive_name.tar.gz"
printf 'Source archive: %s\n' "$output_dir/$archive_name.tar.gz"
