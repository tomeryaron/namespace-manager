#!/bin/bash

# Build macOS binary for local testing
echo "Building macOS binary for local testing..."
go build -o namespace-manager-mac ./cmd/app

# Build Linux binary for Docker
echo "Building Linux binary for Docker..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o namespace-manager ./cmd/app

# Build Docker image
echo "Building Docker image..."
docker build -t namespace-manager:latest .

echo "Done!"
echo ""
echo "To test locally (macOS): ./namespace-manager-mac"
echo "Docker image: namespace-manager:latest"