#!/usr/bin/env zsh

echo "🔧 Testing OpenAPI generation manually..."

cd "/Users/npisarev/Projects/My Drafts/mainote-bot"

# Test if Docker is working
echo "1. Testing Docker..."
if docker --version > /dev/null 2>&1; then
    echo "✅ Docker is available"
else
    echo "❌ Docker is not available"
    exit 1
fi

# Test if the OpenAPI source file exists
echo "2. Checking source file..."
if [[ -f "mainote_server/api/src.yaml" ]]; then
    echo "✅ Source file exists: mainote_server/api/src.yaml"
else
    echo "❌ Source file not found"
    exit 1
fi

# Create directories
echo "3. Creating directories..."
mkdir -p "mainote_server/api/generated"
mkdir -p "mainote_server/pkg/generated"

# Test YAML generation manually
echo "4. Testing YAML generation..."
docker run --rm \
    -v "${PWD}":/app \
    -w /app \
    pscht/oapigen:v1 \
    generate \
    -i "mainote_server/api/src.yaml" \
    -g openapi-yaml \
    -o "/app/mainote_server/api/generated/" \
    --additional-properties="outputFile=openapi.yaml"

echo "YAML generation result:"
ls -la mainote_server/api/generated/

# Test Go generation manually
echo "5. Testing Go generation..."
TEMP_DIR="/tmp/openapi_test_$(date +%s)"
mkdir -p "$TEMP_DIR"

docker run --rm \
    -v "${PWD}":/app \
    -v "$TEMP_DIR":/output \
    -w /app \
    pscht/oapigen:v1 \
    generate \
    -i "mainote_server/api/src.yaml" \
    -g go-server \
    -o "/output" \
    -t "./extra/openapi-templates/go-server" \
    --additional-properties="outputAsLibrary=true,sourceFolder=api,onlyInterfaces=true,packageName=api"

echo "Go generation temp result:"
ls -la "$TEMP_DIR"

if [[ -d "$TEMP_DIR/api" ]]; then
    echo "API subdirectory contents:"
    ls -la "$TEMP_DIR/api"
    
    echo "Copying Go files..."
    cp "$TEMP_DIR/api"/*.go "mainote_server/pkg/generated/" 2>/dev/null || echo "No Go files to copy"
fi

echo "Final Go result:"
ls -la mainote_server/pkg/generated/

# Cleanup
rm -rf "$TEMP_DIR"

echo "6. Test completed!"
