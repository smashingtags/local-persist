#!/usr/bin/env bash

# Modern Docker run script for local-persist volume plugin
set -e

# Default values
DEFAULT_DATA_PATH="/data/local-persist"
DEFAULT_PLUGIN_DATA_PATH="/var/lib/docker/plugin-data"
IMAGE_NAME="${IMAGE_NAME:-local-persist:latest}"

# Check for required environment variables with defaults
DATA_VOLUME="${DATA_VOLUME:-$DEFAULT_DATA_PATH}"
PLUGIN_DATA_PATH="${PLUGIN_DATA_PATH:-$DEFAULT_PLUGIN_DATA_PATH}"

echo "Starting local-persist volume plugin..."
echo "Data volume path: $DATA_VOLUME"
echo "Plugin data path: $PLUGIN_DATA_PATH"

# Ensure directories exist
mkdir -p "$DATA_VOLUME"
mkdir -p "$PLUGIN_DATA_PATH"

# Build the image if it doesn't exist
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
    echo "Building $IMAGE_NAME..."
    docker build -t "$IMAGE_NAME" .
fi

# Run the plugin container
CMD="docker run -d \
    --name local-persist-plugin \
    --restart unless-stopped \
    --privileged \
    --network host \
    -v /run/docker/plugins/:/run/docker/plugins/ \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    -v ${PLUGIN_DATA_PATH}:/var/lib/docker/plugin-data/ \
    -v ${DATA_VOLUME}:/data \
    -e LOG_LEVEL=${LOG_LEVEL:-info} \
    $IMAGE_NAME"

echo "Running: $CMD"
exec $CMD
