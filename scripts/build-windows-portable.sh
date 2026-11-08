#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
project_dir="$(cd -- "$script_dir/.." && pwd)"

cd "$project_dir"

go tool mygo build \
  -platform windows/amd64 \
  -skip-installer \
  -o build/portable \
  .

printf '\nPortable output: %s\n' \
  "$project_dir/build/portable/windows-amd64"
