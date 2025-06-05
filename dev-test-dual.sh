#!/bin/bash

# Development testinecho -e "${BLUE}📦 Building Go backend...${NC}"
cd mainote_server
go build -o ../mainote-backend cmd/server/main.go
cd ..

echo -e "${GREEN}✅ Go backend built successfully${NC}"

echo -e "${BLUE}🔧 Setting up environment variables...${NC}"
export GO_PORT=8081
export GO_BACKEND_URL=http://localhost:8081
export ENVIRONMENT=development
export SENTRY_DSN=${SENTRY_DSN:-""}"

echo -e "${BLUE}🚀 Starting Go backend on port 8081...${NC}"
./mainote-backend &
GO_PID=$!l Python bot + Go backend
# This script helps you test both services locally before deployment

set -e

echo "🚀 Setting up dual service testing environment..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to cleanup background processes
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up processes...${NC}"
    if [ ! -z "$GO_PID" ] && ps -p $GO_PID > /dev/null; then
        kill $GO_PID
        echo "Stopped Go backend (PID: $GO_PID)"
    fi
    if [ ! -z "$PYTHON_PID" ] && ps -p $PYTHON_PID > /dev/null; then
        kill $PYTHON_PID
        echo "Stopped Python bot (PID: $PYTHON_PID)"
    fi
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM EXIT

echo -e "${BLUE}📦 Building Go backend...${NC}"
cd mainote_server
go build -o ../mainote-backend cmd/server/main.go
cd ..

echo -e "${GREEN}✅ Go backend built successfully${NC}"

echo -e "${BLUE}🔧 Setting up environment variables...${NC}"
export GO_PORT=8081
export GO_BACKEND_URL=http://localhost:8081
export ENVIRONMENT=development
export SENTRY_DSN=${SENTRY_DSN:-""}

echo -e "${BLUE}🚀 Starting Go backend on port 8081...${NC}"
./mainote-backend &
GO_PID=$!

# Wait a moment for Go backend to start
sleep 2

echo -e "${BLUE}🚀 Starting Python bot on port 8080...${NC}"
python main.py &
PYTHON_PID=$!

# Wait a moment for Python bot to start
sleep 3

echo -e "${GREEN}✅ Both services are running!${NC}"
echo -e "${BLUE}📋 Service Status:${NC}"
echo -e "  🐍 Python Bot:  http://localhost:8080"
echo -e "  🐹 Go Backend:  http://localhost:8081"
echo -e "  🏥 Health Check: http://localhost:8081/health"

echo -e "\n${YELLOW}🧪 Testing services...${NC}"

# Test Go backend health
echo -e "${BLUE}Testing Go backend health check...${NC}"
if curl -s http://localhost:8081/health > /dev/null; then
    echo -e "${GREEN}✅ Go backend health check passed${NC}"
else
    echo -e "${RED}❌ Go backend health check failed${NC}"
fi

# Test Python bot (basic connectivity)
echo -e "${BLUE}Testing Python bot connectivity...${NC}"
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Python bot is responding${NC}"
else
    echo -e "${YELLOW}⚠️  Python bot may not be fully ready (this is normal for webhook bots)${NC}"
fi

echo -e "\n${GREEN}🎉 Dual service setup complete!${NC}"
echo -e "${YELLOW}💡 Press Ctrl+C to stop both services${NC}"

# Keep script running and monitor processes
while true do
    sleep 1
    
    # Check if processes are still running
    if ! ps -p $GO_PID > /dev/null; then
        echo -e "${RED}❌ Go backend stopped unexpectedly${NC}"
        break
    fi
    
    if ! ps -p $PYTHON_PID > /dev/null; then
        echo -e "${RED}❌ Python bot stopped unexpectedly${NC}"
        break
    fi
done 