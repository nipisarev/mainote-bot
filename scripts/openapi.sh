#!/usr/bin/env bash

# OpenAPI Generation Script for Mainote Server
# This script generates OpenAPI specifications and Go server code

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SERVER_DIR="${PROJECT_ROOT}/mainote_server"

function docker_run() {
  docker run --rm \
    -v "${PROJECT_ROOT}":/app \
    -w /app/mainote_server \
    "${@}"
}

function generate_openapi() {
  local src=''
  local out=''
  local subdir=''
  local jopts='-Xss4M'

  local args=''

  while (($# > 0)); do
    local arg="${1}"; shift

    case "${arg}" in
      -s | --source) src="${1}"; shift ;;
      -o | --output) out="${1}"; shift ;;
      -d | --subdir) subdir="${1}"; shift ;;
      -j | --java-opts) jopts="${1}"; shift ;;
      *) args="${args} ${arg#'-X'}"
    esac
  done

  if [[ -z "${out}" ]]; then
    echo "ERROR: -o | --out is required"
    exit 1
  fi

  local root="${SERVER_DIR}"
  local tmpdir
  tmpdir="$(mktemp -d)"

  echo "🧹 Cleaning output directory: ${root}/${out}/${subdir}"
  # shellcheck disable=SC2115
  rm -rf "${root}/${out}/${subdir}"
  mkdir -p "${root}/${out}"

  echo "🐳 Running OpenAPI generator with Docker..."
  docker_run \
    -v "${tmpdir}:/output" \
    -e JAVA_OPTS="${jopts}" \
    openapitools/openapi-generator-cli:v7.0.1 \
    generate \
      -i "${src}" \
      -o '/output' \
      ${args#' '}

  echo "📋 Copying generated files to output directory..."
  cp -R "${tmpdir}/${subdir}" "${root}/${out}/"

  rm -rf "${tmpdir}"
  echo "✅ OpenAPI generation completed"
}

function generate_api_yaml() {
  echo "📄 Generating OpenAPI YAML specification..."
  
  generate_openapi -g openapi-yaml \
     -s "api/src.yaml" \
     -o "api/generated/" \
     --additional-properties="outputFile=openapi.yaml" \
     --skip-operation-example \
     -d "openapi.yaml"
}

function generate_api_server_go() {
  local only_interfaces=${1:-true}  # Default to true if not provided

  echo "🔧 Generating Go server code (onlyInterfaces=${only_interfaces})..."
  
  # Generate using custom templates
  generate_openapi -g go-server \
     -s "api/src.yaml" \
     -o "pkg/generated" \
     -d "api" \
     -t "/app/extra/openapi-templates" \
     --additional-properties="outputAsLibrary=true,sourceFolder=api,onlyInterfaces=${only_interfaces},packageName=api,structPrefix=Api,enumClassPrefix=true,useCustomStructTags=true"
}

function generate_api() {
  echo "🚀 Starting OpenAPI generation for Mainote Server..."
  echo "📁 Project root: ${PROJECT_ROOT}"
  echo "📁 Server directory: ${SERVER_DIR}"
  
  # Check if source file exists
  if [[ ! -f "${SERVER_DIR}/api/src.yaml" ]]; then
    echo "❌ ERROR: ${SERVER_DIR}/api/src.yaml not found!"
    exit 1
  fi
  
  echo "✅ Source file found"
  
  # Change to server directory for relative paths to work correctly
  cd "${SERVER_DIR}"
  
  # Generate YAML specification
  generate_api_yaml
  
  # Generate Go server interfaces
  generate_api_server_go true
  
  # Run go mod tidy to clean up dependencies
  echo "🧹 Running go mod tidy..."
  go mod tidy
  
  # Format generated code
  echo "🎨 Formatting generated Go code..."
  go fmt ./pkg/generated/...
  
  echo "🎉 OpenAPI generation completed successfully!"
  echo ""
  echo "Generated files:"
  echo "  📄 ${SERVER_DIR}/api/generated/openapi.yaml - OpenAPI specification"
  echo "  🔧 ${SERVER_DIR}/pkg/generated/api/ - Go server interfaces"
}

function clean() {
  echo "🧹 Cleaning generated files..."
  cd "${SERVER_DIR}"
  rm -rf api/generated/*
  rm -rf pkg/generated/api/*
  echo "✅ Generated files cleaned"
}

function help() {
  echo "OpenAPI Generation Script for Mainote Server"
  echo ""
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  generate    Generate OpenAPI YAML and Go server code (default)"
  echo "  clean       Clean all generated files"
  echo "  help        Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0                # Generate all files"
  echo "  $0 generate       # Generate all files"
  echo "  $0 clean          # Clean generated files"
  echo ""
}

# Main script logic
case "${1:-generate}" in
  generate)
    generate_api
    ;;
  clean)
    clean
    ;;
  help|--help|-h)
    help
    ;;
  *)
    echo "❌ Unknown command: $1"
    echo ""
    help
    exit 1
    ;;
esac 