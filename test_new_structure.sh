#!/usr/bin/env bash

# Test script for the updated OpenAPI generator with new directory structure
# Testing generation into mainote_server/pkg/generated/

set -e

echo "🧪 Testing Updated OpenAPI Script"
echo "================================="
echo "Target directory: mainote_server/pkg/generated/"
echo ""

# Load the updated OpenAPI functions
source scripts/development/openapi.sh

# Test the new mainote-specific function
echo "📋 Test: Generate Go files for mainote_server"
echo "---------------------------------------------"

# Generate files using the new function
generate_mainote_server_api "public" true

# Verify the files were created in the correct location
echo ""
echo "📋 Verification: Check generated files"
echo "-------------------------------------"

if [[ -d "mainote_server/pkg/generated/public" ]]; then
    GO_FILES=$(find mainote_server/pkg/generated/public -name "*.go" | wc -l | tr -d ' ')
    echo "✅ Generated $GO_FILES Go files successfully"
    echo "📁 Location: mainote_server/pkg/generated/public/"
    
    # List some key files
    echo ""
    echo "Key generated files:"
    find mainote_server/pkg/generated/public -name "*.go" | head -8 | while read file; do
        echo "   📄 $(basename "$file")"
    done
    
    # Check if there are more files
    TOTAL_FILES=$(find mainote_server/pkg/generated/public -name "*.go" | wc -l | tr -d ' ')
    if [[ $TOTAL_FILES -gt 8 ]]; then
        echo "   ... and $((TOTAL_FILES - 8)) more files"
    fi
    
    echo ""
    echo "🎉 SUCCESS: Go files generated in correct location!"
    echo "📍 Files are now in: mainote_server/pkg/generated/public/"
else
    echo "❌ FAILED: No files generated in expected location"
    exit 1
fi

echo ""
echo "🧹 Cleanup:"
echo "To remove test files: rm -rf mainote_server/pkg/generated/public"
