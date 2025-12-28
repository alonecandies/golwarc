#!/bin/bash

# Golwarc Development Script
# Loads environment variables and runs main.go

set -e

echo "========================================"
echo "  Golwarc Development Environment"
echo "========================================"

# Check if .env file exists
if [ -f .env ]; then
    echo "Loading environment variables from .env..."
    export $(cat .env | grep -v '^#' | xargs)
    echo "✓ Environment variables loaded"
else
    echo "⚠ No .env file found, using system defaults"
fi

# Check if config.yaml exists, otherwise use example
if [ ! -f config.yaml ]; then
    if [ -f config.example.yaml ]; then
        echo "Creating config.yaml from example..."
        cp config.example.yaml config.yaml
        echo "✓ config.yaml created"
    else
        echo "❌ No config.yaml or config.example.yaml found"
        exit 1
    fi
fi

echo ""
echo "Starting Golwarc application..."
echo "========================================"
echo ""

# Run main.go
go run main.go

echo ""
echo "========================================"
echo "Application finished"
echo "========================================"
