#!/bin/bash
# scripts/check-tools.sh - Check if required external tools are installed
# Usage: ./scripts/check-tools.sh

set -e

cd "$(dirname "$0")/.."

echo "=== Checking Required Tools ==="
echo ""

REQUIRED_TOOLS=("fd" "osqueryi" "glances")
OPTIONAL_TOOLS=("bmon" "mtr" "nmap")

missing=0

echo "Required tools:"
for tool in "${REQUIRED_TOOLS[@]}"; do
    if command -v "$tool" &> /dev/null; then
        version=$("$tool" --version 2>&1 | head -1 || echo "installed")
        echo "  ✓ $tool: $version"
    else
        echo "  ✗ $tool: NOT FOUND (optional enhancement)"
    fi
done

echo ""
echo "Optional tools (for enhanced functionality):"
for tool in "${OPTIONAL_TOOLS[@]}"; do
    if command -v "$tool" &> /dev/null; then
        version=$("$tool" --version 2>&1 | head -1 || echo "installed")
        echo "  ✓ $tool: $version"
    else
        echo "  ○ $tool: not installed"
    fi
done

echo ""
echo "To install missing tools:"
echo "  brew install fd osqueryi glances bmon mtr nmap"
echo ""

# Check if we have at least the basic tools
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is required but not installed"
    missing=1
else
    go version | sed 's/^/  /'
fi

echo ""
if [ $missing -eq 1 ]; then
    exit 1
fi

echo "Core tools check passed!"
