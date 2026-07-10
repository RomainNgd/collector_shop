#!/usr/bin/env sh

set -eu

TAG="${TAG:-local}"
REGISTRY="${REGISTRY:-}"
NO_CACHE="${NO_CACHE:-false}"
BUILDX_CACHE_FROM="${BUILDX_CACHE_FROM:-}"
BUILDX_CACHE_TO="${BUILDX_CACHE_TO:-}"

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

build_cmd="docker build"
if [ -n "$BUILDX_CACHE_FROM" ] || [ -n "$BUILDX_CACHE_TO" ]; then
  build_cmd="docker buildx build --load"
  [ -n "$BUILDX_CACHE_FROM" ] && cache_args="$cache_args --cache-from ${BUILDX_CACHE_FROM},scope=go-api"
  [ -n "$BUILDX_CACHE_TO" ] && cache_args="$cache_args --cache-to ${BUILDX_CACHE_TO},scope=go-api"
fi

echo "Building $api_image"
$build_cmd $cache_args -f "$repo_root/build/docker/go-api.Dockerfile" -t "$api_image" "$repo_root"

cache_args=""
if [ "$NO_CACHE" = "true" ]; then
  cache_args="--no-cache"
fi
if [ -n "$BUILDX_CACHE_FROM" ] || [ -n "$BUILDX_CACHE_TO" ]; then
  [ -n "$BUILDX_CACHE_FROM" ] && cache_args="$cache_args --cache-from ${BUILDX_CACHE_FROM},scope=collector-spa"
  [ -n "$BUILDX_CACHE_TO" ] && cache_args="$cache_args --cache-to ${BUILDX_CACHE_TO},scope=collector-spa"
fi

echo "Building $spa_image"
$build_cmd $cache_args -f "$repo_root/build/docker/collector-spa.Dockerfile" -t "$spa_image" "$repo_root"

echo
echo "Images built successfully:"
echo " - $api_image"
echo " - $spa_image"
