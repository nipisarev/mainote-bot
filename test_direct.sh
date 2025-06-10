#!/usr/bin/env bash

cd "/Users/npisarev/Projects/My Drafts/mainote-bot"

# Test just the generate_openapi function directly
echo "Testing generate_openapi function for YAML..."

source scripts/development/openapi.sh

# Generate YAML directly using the core function
generate_openapi -g openapi-yaml \
  -s "mainote_server/api/src.yaml" \
  -o "mainote_server/api/generated/" \
  --additional-properties="outputFile=test.yaml" \
  --skip-operation-example \
  -d "test.yaml"

echo "Checking generated files:"
ls -la mainote_server/api/generated/

echo ""
echo "Testing generate_openapi function for Go..."

# Generate Go files directly using the core function  
generate_openapi -g go-server \
  -s "mainote_server/api/src.yaml" \
  -o "mainote_server/pkg/generated" \
  -t "./extra/openapi-templates/go-server" \
  -d "api" \
  --additional-properties="outputAsLibrary=true,sourceFolder=api,onlyInterfaces=true,packageName=api"

echo "Checking generated Go directory:"
ls -la mainote_server/pkg/generated/

if [[ -d "mainote_server/pkg/generated/api" ]]; then
  echo "Go files in api subdirectory:"
  find mainote_server/pkg/generated/api -name "*.go" | head -10
fi
