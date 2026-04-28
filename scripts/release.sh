#!/bin/bash
# scripts/release.sh - Prepare and create a release
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v2.0.0

set -e

VERSION="${1?Usage: $0 <version>}"

cd "$(dirname "$0")/.."

echo "=== Release Preparation for $VERSION ==="

echo ""
echo "1. Running full test suite..."
./scripts/test-all.sh

echo ""
echo "2. Building binary..."
go build -o gone ./cmd/gone
echo "Binary: gone"

echo ""
echo "3. Checking for uncommitted changes..."
if git diff --quiet && git diff --staged --quiet; then
    echo "   Working tree is clean"
else
    echo "   WARNING: You have uncommitted changes"
    git status
fi

echo ""
echo "4. Next steps:"
echo "   - Review CHANGELOG.md"
echo "   - Update version in code (search for current version)"
echo "   - Create git tag: git tag -a $VERSION -m 'Release $VERSION'"
echo "   - Push tag: git push origin $VERSION"
echo "   - Create GitHub release"
echo ""
echo "Release preparation complete!"
