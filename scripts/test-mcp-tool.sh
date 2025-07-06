#!/bin/bash

# Test MCP tool with pretty-printed JSON output
# Usage: ./scripts/test-mcp-tool.sh <tool_name>

set -e

TOOL_NAME="$1"
if [[ -z "$TOOL_NAME" ]]; then
    echo "Usage: $0 <tool_name>"
    echo "Available tools: find_symbol_definitions_by_name, find_symbol_references_by_anchor, list_symbols_in_file"
    exit 1
fi

# Validate tool name
case "$TOOL_NAME" in
    "find_symbol_definitions_by_name"|"find_symbol_references_by_anchor"|"list_symbols_in_file")
        ;;
    *)
        echo "Error: Unknown tool '$TOOL_NAME'"
        echo "Available tools: find_symbol_definitions_by_name, find_symbol_references_by_anchor, list_symbols_in_file"
        exit 1
        ;;
esac

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

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Warning: jq not found. JSON output will not be formatted."
    USE_JQ=false
else
    USE_JQ=true
fi

echo "Testing $TOOL_NAME tool..."
echo "=================================="

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
OUTPUT=$(timeout 10s bash -c '{
    tr -d "\n" < ./testdata/mcp_init.json; echo "";
    sleep 1;
    tr -d "\n" < ./testdata/'$TOOL_NAME'.input.json; echo "";
    sleep 2;
} | ./bin/gopls-mcp --workspace-root ./testdata/example 2>/dev/null' || true)

# Process the output - save to temp file to avoid subshell issues
TEMP_OUTPUT="/tmp/mcp_output_$$"
echo "$OUTPUT" > "$TEMP_OUTPUT"

while IFS= read -r line; do
    # Skip empty lines
    [[ -z "$line" ]] && continue
    
    # Try to parse as JSON
    if echo "$line" | jq -e . >/dev/null 2>&1; then
        # Check if this is a tool result (has result.content field)
        if echo "$line" | jq -e '.result.content' >/dev/null 2>&1; then
            echo "Tool Result:"
            echo "------------"
            
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
            else
                echo "No JSON content found in response"
            fi
            echo ""
        fi
        # Skip initialization responses (don't print them)
    else
        # Not JSON, probably an error or log message
        echo "Output: $line"
        echo ""
    fi
done < "$TEMP_OUTPUT"

# Clean up temp file
rm -f "$TEMP_OUTPUT"

echo "Test completed."