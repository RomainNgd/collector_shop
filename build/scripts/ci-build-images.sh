#!/usr/bin/env bash

set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

TAG="${TAG:-${IMAGE_TAG:-ci}}"
REGISTRY="${REGISTRY:-${IMAGE_REGISTRY:-docker.io}}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-collector-shop}"

TAG="$TAG" \
REGISTRY="$REGISTRY/$IMAGE_NAMESPACE" \
bash "$script_dir/build-images.sh"
