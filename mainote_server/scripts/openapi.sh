#!/usr/bin/env bash

# OpenAPI Generation Script for Mainote Server
# Adapted from the main project's openapi.sh for single-service architecture

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MAIN_PROJECT_ROOT="$(cd "${PROJECT_ROOT}/.." && pwd)"

function docker_run() {
  docker run --rm \
    -v "${PWD}":/app \
    -w /app \
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

  local root="${PROJECT_ROOT}"
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
    ghcr.io/qase-tms/openapi/oapigen:v1 \
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
     --additional-properties="outputFile=private.yaml" \
     --skip-operation-example \
     -d "private.yaml"
}

function generate_api_server_go() {
  local only_interfaces=${1:-true}  # Default to true if not provided

  echo "🔧 Generating Go server code (onlyInterfaces=${only_interfaces})..."
  
  # Generate using standard templates for now (no custom templates)
  generate_openapi -g go-server \
     -s "api/src.yaml" \
     -o "api/generated" \
     -d "private" \
     --additional-properties="outputAsLibrary=true,sourceFolder=private,onlyInterfaces=${only_interfaces},packageName=private"
}

function generate_api() {
  echo "🚀 Starting OpenAPI generation for Mainote Server..."
  echo "📁 Project root: ${PROJECT_ROOT}"
  echo "📁 Main project root: ${MAIN_PROJECT_ROOT}"
  
  # Change to project root for relative paths to work correctly
  cd "${PROJECT_ROOT}"
  
  # Check if source file exists
  if [[ ! -f "api/src.yaml" ]]; then
    echo "❌ ERROR: api/src.yaml not found!"
    exit 1
  fi
  
  echo "✅ Source file found"
  
  # Generate YAML specification
  generate_api_yaml
  
  # Generate Go server interfaces
  generate_api_server_go true
  
  # Run go mod tidy to clean up dependencies
  echo "🧹 Running go mod tidy..."
  go mod tidy
  
  # Format generated code
  echo "🎨 Formatting generated Go code..."
  go fmt ./api/generated/...
  
  echo "🎉 OpenAPI generation completed successfully!"
  echo ""
  echo "Generated files:"
  echo "  📄 api/generated/private.yaml - OpenAPI specification"
  echo "  🔧 api/generated/private/ - Go server interfaces"
}

function clean() {
  echo "🧹 Cleaning generated files..."
  cd "${PROJECT_ROOT}"
  rm -rf api/generated/*
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
