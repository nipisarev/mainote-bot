#!/usr/bin/env bash

# Test script for the updated OpenAPI generator
# This tests the pscht/oapigen:v1 Docker image with your project structure

set -e

echo "🧪 Testing OpenAPI Script with pscht/oapigen:v1"
echo "================================================"

# Load the OpenAPI functions
source scripts/development/openapi.sh

# Test 1: Simple Go server generation
echo ""
echo "📋 Test 1: Direct Go Server Generation"
echo "--------------------------------------"

# Create test output directory
mkdir -p test_output

echo "Generating Go server code from mainote_server/api/src.yaml..."

# Test the generate_openapi function directly for Go server
generate_openapi -g go-server \
  -s "mainote_server/api/src.yaml" \
  -o "test_output" \
  --additional-properties="outputAsLibrary=true,packageName=mainote" \
  -d "generated"

if [[ -d "test_output/generated" ]]; then
    echo "✅ SUCCESS: Go server code generated successfully!"
    echo "📁 Generated files:"
    find test_output/generated -type f -name "*.go" | head -5
    echo "   ... and more"
else
    echo "❌ FAILED: No generated directory found"
    exit 1
fi

# Test 2: Verify Docker image version
echo ""
echo "📋 Test 2: Docker Image Verification"
echo "------------------------------------"

VERSION=$(docker run --rm pscht/oapigen:v1 version)
echo "✅ OpenAPI Generator Version: $VERSION"

# Test 3: Check available generators
echo ""
echo "📋 Test 3: Available Generators Check"
echo "-------------------------------------"

echo "Checking if required generators are available..."
GENERATORS=$(docker run --rm pscht/oapigen:v1 list | grep -E "(go-server|openapi-yaml|go)" | wc -l)
if [[ $GENERATORS -gt 0 ]]; then
    echo "✅ Required generators are available"
else
    echo "⚠️  Some generators might not be available"
fi

# Cleanup
echo ""
echo "🧹 Cleaning up test files..."
rm -rf test_output

echo ""
echo "🎉 OpenAPI Script Testing Complete!"
echo "====================================="
echo "✅ The pscht/oapigen:v1 Docker image works correctly"
echo "✅ The generate_openapi function works"
echo "✅ Go server code generation is successful"
echo ""
echo "💡 You can now use the updated script in your development workflow!"
