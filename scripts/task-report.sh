#!/bin/bash
# scripts/task-report.sh - Generate task status report
# Usage: ./scripts/task-report.sh

set -e

cd "$(dirname "$0")/.."

echo "=== Gone Task Report ==="
echo ""

TOTAL=$(grep -c "^| T-" TASKS.md)
DONE=$(grep -c "done" TASKS.md | head -1)
PENDING=$(grep -c "pending" TASKS.md | head -1)
IN_PROGRESS=$(grep -c "in_progress" TASKS.md | head -1)

echo "Total Tasks: $TOTAL"
echo "Completed:   $DONE"
echo "In Progress: $IN_PROGRESS"
echo "Pending:     $PENDING"
echo ""

echo "=== By Phase ==="
grep -E "^## Phase" TASKS.md | while read -r line; do
    echo "  $line"
done

echo ""
echo "=== Task Dependencies ==="
grep -A2 "Task Dependencies" TASKS.md | tail -20
