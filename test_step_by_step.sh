#!/usr/bin/env bash

echo "🧪 Testing OpenAPI generation step by step..."

cd "/Users/npisarev/Projects/My Drafts/mainote-bot"

# Source the script
echo "1. Sourcing OpenAPI script..."
source scripts/development/openapi.sh

# Test YAML generation
echo "2. Testing YAML generation..."
generate_api_yaml

# Check if YAML was created
if [[ -f "mainote_server/api/generated/openapi.yaml" ]]; then
    echo "✅ YAML file generated successfully"
    ls -la mainote_server/api/generated/
else
    echo "❌ YAML file not found"
fi

# Test Go generation
echo "3. Testing Go generation..."
generate_api_server_go

# Check if Go files were created
echo "4. Checking for Go files..."
if [[ -n "$(find mainote_server/pkg/generated -name '*.go' 2>/dev/null)" ]]; then
    echo "✅ Go files generated successfully"
    ls -la mainote_server/pkg/generated/
    echo "Go files found:"
    find mainote_server/pkg/generated -name "*.go"
else
    echo "❌ No Go files found"
    echo "Checking if directory exists:"
    ls -la mainote_server/pkg/
fi

echo "5. Test completed!"
