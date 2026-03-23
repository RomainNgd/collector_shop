#!/usr/bin/env bash

set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd)"

cd "$project_dir"

echo "Installing frontend dependencies"
npm ci

echo "Running Svelte checks"
npm run check
