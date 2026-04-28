#!/bin/bash
# scripts/coverage.sh - Run tests with coverage report
# Usage: ./scripts/coverage.sh [package]

set -e

cd "$(dirname "$0")/.."

PACKAGE="${1:-./internal/...}"

echo "=== Running tests with coverage ==="
go test -coverprofile=coverage.out "$PACKAGE"
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "Coverage report generated: coverage.html"
echo ""
go tool cover -func=coverage.out | tail -10
