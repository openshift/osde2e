#!/bin/bash
# Build verification script for osde2e dashboard

set -e

echo "=== osde2e Dashboard Build Verification ==="
echo ""

# Check Go version
echo "1. Checking Go version..."
if command -v go &> /dev/null; then
    go version
else
    echo "ERROR: Go not found in PATH"
    exit 1
fi
echo ""

# Check if we're in the right directory
echo "2. Checking directory..."
if [ ! -f "go.mod" ]; then
    echo "ERROR: Not in osde2e root directory"
    exit 1
fi
echo "✓ In osde2e root directory"
echo ""

# Verify dashboard files exist
echo "3. Verifying dashboard files..."
FILES=(
    "pkg/dashboard/models/types.go"
    "pkg/dashboard/config/config.go"
    "pkg/dashboard/collectors/reserves.go"
    "pkg/dashboard/collectors/usage.go"
    "pkg/dashboard/collectors/s3tests.go"
    "pkg/dashboard/server/server.go"
    "pkg/dashboard/server/templates.go"
    "pkg/dashboard/handlers/utils.go"
    "cmd/osde2e/dashboard/cmd.go"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✓ $file"
    else
        echo "✗ MISSING: $file"
        exit 1
    fi
done
echo ""

# Verify templates exist
echo "4. Verifying HTML templates..."
TEMPLATES=(
    "pkg/dashboard/server/templates/base.html"
    "pkg/dashboard/server/templates/dashboard.html"
    "pkg/dashboard/server/templates/reserves.html"
    "pkg/dashboard/server/templates/usage.html"
    "pkg/dashboard/server/templates/tests.html"
)

for template in "${TEMPLATES[@]}"; do
    if [ -f "$template" ]; then
        echo "✓ $template"
    else
        echo "✗ MISSING: $template"
        exit 1
    fi
done
echo ""

# Check for syntax errors (gofmt)
echo "5. Checking Go syntax..."
DASHBOARD_FILES=$(find pkg/dashboard cmd/osde2e/dashboard -name "*.go" 2>/dev/null)
if [ -n "$DASHBOARD_FILES" ]; then
    gofmt -l $DASHBOARD_FILES > /tmp/dashboard-fmt-check.txt
    if [ -s /tmp/dashboard-fmt-check.txt ]; then
        echo "⚠ Files need formatting:"
        cat /tmp/dashboard-fmt-check.txt
    else
        echo "✓ All files properly formatted"
    fi
else
    echo "⚠ No Go files found"
fi
echo ""

# Try to build dashboard package
echo "6. Building dashboard package..."
if go build -v ./pkg/dashboard/... 2>&1 | tee /tmp/dashboard-build.log; then
    echo "✓ Dashboard package builds successfully"
else
    echo "✗ Build failed. See /tmp/dashboard-build.log for details"
    exit 1
fi
echo ""

# Try to build main osde2e with dashboard
echo "7. Building osde2e with dashboard command..."
if go build -o /tmp/osde2e ./cmd/osde2e 2>&1 | tee /tmp/osde2e-build.log; then
    echo "✓ osde2e builds successfully with dashboard command"
else
    echo "✗ Build failed. See /tmp/osde2e-build.log for details"
    exit 1
fi
echo ""

# Verify dashboard command is registered
echo "8. Verifying dashboard command..."
if grep -q "dashboard.Cmd" cmd/osde2e/main.go; then
    echo "✓ Dashboard command registered in main.go"
else
    echo "✗ Dashboard command NOT registered in main.go"
    exit 1
fi
echo ""

# Test help command
echo "9. Testing dashboard help..."
if /tmp/osde2e dashboard --help > /tmp/dashboard-help.txt 2>&1; then
    echo "✓ Dashboard help command works"
    echo ""
    echo "=== Dashboard Help Output ==="
    cat /tmp/dashboard-help.txt
else
    echo "✗ Dashboard help command failed"
    exit 1
fi
echo ""

echo "==================================="
echo "✅ All verification checks passed!"
echo "==================================="
echo ""
echo "Dashboard is ready to use. Start with:"
echo "  ./osde2e dashboard --port 8080"
echo ""
