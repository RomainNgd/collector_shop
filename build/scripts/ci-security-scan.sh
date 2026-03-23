#!/usr/bin/env bash

set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../.." && pwd)"

TRIVY_SEVERITY="${TRIVY_SEVERITY:-HIGH,CRITICAL}"
TRIVY_EXIT_CODE="${TRIVY_EXIT_CODE:-1}"
IMAGE_TAG="${IMAGE_TAG:-ci}"
IMAGE_REGISTRY="${IMAGE_REGISTRY:-docker.io}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-collector-shop}"
SCAN_MODE="${SCAN_MODE:-fs}"

scan_filesystem() {
  echo "Running Trivy filesystem scan from $repo_root"
  trivy fs \
    --scanners vuln,secret,misconfig \
    --severity "$TRIVY_SEVERITY" \
    --exit-code "$TRIVY_EXIT_CODE" \
    "$repo_root"
}

scan_image() {
  local image_name="$1"

  echo "Running Trivy image scan for $image_name"
  trivy image \
    --severity "$TRIVY_SEVERITY" \
    --exit-code "$TRIVY_EXIT_CODE" \
    "$image_name"
}

case "$SCAN_MODE" in
  fs)
    scan_filesystem
    ;;
  images)
    scan_image "$IMAGE_REGISTRY/$IMAGE_NAMESPACE/go-api:$IMAGE_TAG"
    scan_image "$IMAGE_REGISTRY/$IMAGE_NAMESPACE/collector-spa:$IMAGE_TAG"
    ;;
  *)
    echo "Unsupported SCAN_MODE: $SCAN_MODE"
    exit 1
    ;;
esac
