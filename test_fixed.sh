#!/usr/bin/env bash

cd "/Users/npisarev/Projects/My Drafts/mainote-bot"

echo "🧪 Testing OpenAPI generation with correct parameters..."

source scripts/development/openapi.sh

# Test YAML generation with manual parameters to debug
echo "1. Testing YAML generation..."
mkdir -p "mainote_server/api/generated"

# Direct call to generate_openapi for YAML
generate_openapi -g openapi-yaml \
  -s "mainote_server/api/src.yaml" \
  -o "mainote_server/api/generated/" \
  -d "."

echo "Checking YAML results:"
ls -la mainote_server/api/generated/

echo ""
echo "2. Testing Go generation..."
mkdir -p "mainote_server/pkg/generated"

# Test Go generation to a temp directory first
TEMP_DIR=$(mktemp -d)
echo "Using temp directory: $TEMP_DIR"

generate_openapi -g go-server \
  -s "mainote_server/api/src.yaml" \
  -o "$TEMP_DIR" \
  -t "./extra/openapi-templates/go-server" \
  -d "api" \
  --additional-properties="outputAsLibrary=true,sourceFolder=api,onlyInterfaces=true,packageName=api"

echo "Temp directory contents:"
ls -la "$TEMP_DIR"

if [[ -d "$TEMP_DIR/api" ]]; then
  echo "API subdirectory contents:"
  ls -la "$TEMP_DIR/api"
  
  echo "Moving Go files to target directory..."
  find "$TEMP_DIR/api" -name "*.go" -exec cp {} "mainote_server/pkg/generated/" \;
  
  echo "Final result:"
  ls -la mainote_server/pkg/generated/
fi

# Cleanup
rm -rf "$TEMP_DIR"
