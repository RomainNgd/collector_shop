#!/usr/bin/env sh

set -eu

TAG="${TAG:-local}"
REGISTRY="${REGISTRY:-}"
NO_CACHE="${NO_CACHE:-false}"

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../.." && pwd)"

image_name() {
  name="$1"

  if [ -z "$REGISTRY" ]; then
    printf '%s\n' "collector-shop/$name:$TAG"
    return
  fi

  printf '%s\n' "$REGISTRY/$name:$TAG"
}

api_image="$(image_name go-api)"
spa_image="$(image_name collector-spa)"

cache_args=""
if [ "$NO_CACHE" = "true" ]; then
  cache_args="--no-cache"
fi

echo "Building $api_image"
docker build $cache_args -f "$repo_root/build/docker/go-api.Dockerfile" -t "$api_image" "$repo_root"

echo "Building $spa_image"
docker build $cache_args -f "$repo_root/build/docker/collector-spa.Dockerfile" -t "$spa_image" "$repo_root"

echo
echo "Images built successfully:"
echo " - $api_image"
echo " - $spa_image"
