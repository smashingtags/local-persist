#!/usr/bin/env bash

# Build script for local-persist using modern Docker approach
set -e

echo "Building local-persist using modernized Dockerfile-build..."

# Build the development image
LOCAL_PERSIST=$(docker build -q -f Dockerfile-build --target dev-builder .)

# Create bin directory if it doesn't exist
mkdir -p bin

# Run the build container with proper volume mounts
echo "Running build container..."
docker run --rm -v "$(pwd)/bin:/src/bin" $LOCAL_PERSIST

echo "Build completed successfully. Binaries are in ./bin/"
ls -la bin/
