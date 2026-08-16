#!/usr/bin/env sh
set -eu
IMAGE_NAME=${1:-cry001}
PLATFORM=${2:-linux/amd64}
docker buildx build --platform "$PLATFORM" -f benzhi.Dockerfile -t "$IMAGE_NAME" .
