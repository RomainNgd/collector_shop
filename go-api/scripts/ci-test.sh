#!/usr/bin/env bash

set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd)"

cd "$project_dir"

echo "Running Go tests from $project_dir"
# Database-backed packages share the same PostgreSQL service in CI.
# Run packages sequentially to prevent concurrent AutoMigrate operations.
go test -p 1 ./... -coverprofile=coverage.out
