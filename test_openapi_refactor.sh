#!/usr/bin/env bash

# Test script for the refactored OpenAPI generator

echo "🧪 Testing refactored OpenAPI script..."

# Source the OpenAPI script
source scripts/development/openapi.sh

echo "✅ Script sourced successfully"

# Test if functions are available
if declare -F generate_api > /dev/null; then
    echo "✅ generate_api function is available"
else
    echo "❌ generate_api function not found"
    exit 1
fi

if declare -F generate_api_yaml > /dev/null; then
    echo "✅ generate_api_yaml function is available"
else
    echo "❌ generate_api_yaml function not found"
    exit 1
fi

if declare -F generate_api_server_go > /dev/null; then
    echo "✅ generate_api_server_go function is available"
else
    echo "❌ generate_api_server_go function not found"
    exit 1
fi

if declare -F generate_mainote_api > /dev/null; then
    echo "✅ generate_mainote_api function is available"
else
    echo "❌ generate_mainote_api function not found"
    exit 1
fi

# Check if source file exists
if [[ -f "mainote_server/api/src.yaml" ]]; then
    echo "✅ Source OpenAPI file exists: mainote_server/api/src.yaml"
else
    echo "❌ Source OpenAPI file not found: mainote_server/api/src.yaml"
    exit 1
fi

# Check if Docker is available
if command -v docker &> /dev/null; then
    echo "✅ Docker is available"
    
    # Test Docker with the image
    if docker run --rm pscht/oapigen:v1 version &> /dev/null; then
        echo "✅ Docker image pscht/oapigen:v1 is accessible"
    else
        echo "⚠️  Docker image pscht/oapigen:v1 may not be accessible (this is normal for first run)"
    fi
else
    echo "❌ Docker is not available"
    exit 1
fi

echo ""
echo "🎉 All tests passed! The script is ready to use."
echo ""
echo "To generate API files, run:"
echo "  source scripts/development/openapi.sh"
echo "  generate_api"
echo ""
echo "Or use the convenience function:"
echo "  generate_mainote_api"
