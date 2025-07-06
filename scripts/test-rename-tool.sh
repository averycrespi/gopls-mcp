#!/bin/bash

# Test rename_symbol_by_anchor MCP tool with backup/restore functionality
# This script ensures the test environment is restored after the rename test

set -e

TOOL_NAME="rename_symbol_by_anchor"
TESTDATA_DIR="./testdata/example"
BACKUP_DIR="./testdata/.backup-example-$$"

# Function to cleanup on exit
cleanup() {
    local exit_code=$?
    echo ""
    echo "Cleaning up..."

    if [[ -d "$BACKUP_DIR" ]]; then
        echo "Restoring original testdata from backup..."
        rm -rf "$TESTDATA_DIR"
        mv "$BACKUP_DIR" "$TESTDATA_DIR"
        echo "✓ Original testdata restored"
    fi

    if [[ $exit_code -eq 0 ]]; then
        echo "✓ Test completed successfully"
    else
        echo "✗ Test failed with exit code $exit_code"
    fi

    exit $exit_code
}

# Set up cleanup trap
trap cleanup EXIT INT TERM

echo "Testing $TOOL_NAME tool with backup/restore..."
echo "================================================"

# Check if required files exist
if [[ ! -f "./bin/gopls-mcp" ]]; then
    echo "Error: gopls-mcp binary not found. Run 'make build' first."
    exit 1
fi

if [[ ! -f "./testdata/mcp_init.json" ]]; then
    echo "Error: mcp_init.json not found in testdata/"
    exit 1
fi

if [[ ! -f "./testdata/${TOOL_NAME}.input.json" ]]; then
    echo "Error: ${TOOL_NAME}.input.json not found in testdata/"
    exit 1
fi

if [[ ! -d "$TESTDATA_DIR" ]]; then
    echo "Error: testdata/example directory not found"
    exit 1
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Warning: jq not found. JSON output will not be formatted."
    USE_JQ=false
else
    USE_JQ=true
fi

# Create backup of testdata
echo "Creating backup of testdata/example..."
cp -r "$TESTDATA_DIR" "$BACKUP_DIR"
echo "✓ Backup created at $BACKUP_DIR"
echo ""

# Show original state
echo "Original Calculator struct (before rename):"
echo "-------------------------------------------"
grep -n "type Calculator" "$TESTDATA_DIR/calculator.go" || echo "No Calculator struct found"
echo ""

# Display the tool input
echo "Tool Input:"
echo "-----------"
if $USE_JQ; then
    cat "./testdata/${TOOL_NAME}.input.json" | jq -C .
else
    cat "./testdata/${TOOL_NAME}.input.json"
fi
echo ""

# Run the MCP server with timeout and capture output
echo "Running rename operation..."
echo "---------------------------"
OUTPUT=$(timeout 15s bash -c '{
    tr -d "\n" < ./testdata/mcp_init.json; echo "";
    sleep 1;
    tr -d "\n" < ./testdata/'$TOOL_NAME'.input.json; echo "";
} | ./bin/gopls-mcp --workspace-root ./testdata/example 2>/dev/null' || true)

# Process the output and extract the rename result
FOUND_RESULT=false
echo "$OUTPUT" | while IFS= read -r line; do
    # Skip empty lines
    [[ -z "$line" ]] && continue

    # Try to parse as JSON
    if echo "$line" | jq -e . >/dev/null 2>&1; then
        # Check if this is a tool result (has result.content field)
        if echo "$line" | jq -e '.result.content' >/dev/null 2>&1; then
            echo "Rename Result:"
            echo "--------------"

            # Extract the JSON content from the MCP response
            JSON_CONTENT=$(echo "$line" | jq -r '.result.content[0].text' 2>/dev/null || echo "")

            if [[ -n "$JSON_CONTENT" && "$JSON_CONTENT" != "null" ]]; then
                if $USE_JQ; then
                    # Pretty print the extracted JSON with colors
                    echo "$JSON_CONTENT" | jq -C .
                else
                    # Fallback: just print the raw JSON content
                    echo "$JSON_CONTENT"
                fi
                FOUND_RESULT=true
            else
                echo "No JSON content found in response"
            fi
        fi
    fi
done

echo ""

echo ""
echo "Test completed. Backup will be restored automatically."