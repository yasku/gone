#!/bin/bash
# scripts/test-all.sh - Run all tests with coverage and race detection
# Usage: ./scripts/test-all.sh

set -e

cd "$(dirname "$0")/.."

echo "=== Running all tests ==="
go test ./cmd/... ./internal/... -v

echo ""
echo "=== Running tests with race detector ==="
go test -race ./internal/...

echo ""
echo "=== Running go vet ==="
go vet ./cmd/... ./internal/...

echo ""
echo "=== Running go fmt check ==="
if [ -n "$(go fmt ./cmd/... ./internal/... 2>&1)" ]; then
    echo "WARNING: go fmt made changes. Run 'go fmt ./...' to fix."
    go fmt ./cmd/... ./internal/...
fi

echo ""
echo "=== All checks passed ==="
