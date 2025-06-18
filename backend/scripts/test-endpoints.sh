#!/bin/bash

# Test script for Relay Voice Server endpoints
set -e

SERVER_URL="http://localhost:8080"
echo "🚀 Testing Relay Voice Server at $SERVER_URL"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    
    echo -e "\n${YELLOW}Testing:${NC} $description"
    echo "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$SERVER_URL$endpoint")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$SERVER_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo -e "${GREEN}✅ Success${NC} (HTTP $http_code)"
        echo "Response: $body" | jq '.' 2>/dev/null || echo "Response: $body"
    else
        echo -e "${RED}❌ Failed${NC} (HTTP $http_code)"
        echo "Response: $body"
    fi
}

# Check if server is running
echo "Checking if server is running..."
if ! curl -s "$SERVER_URL/health" > /dev/null; then
    echo -e "${RED}❌ Server is not running at $SERVER_URL${NC}"
    echo "Please start the server first with: npm run dev"
    exit 1
fi

echo -e "${GREEN}✅ Server is running${NC}"

# Test endpoints
test_endpoint "GET" "/health" "Health check endpoint"

test_endpoint "GET" "/api/projects" "List projects endpoint"

# Test project selection (will likely fail without valid project)
test_endpoint "POST" "/api/projects/test-project/select" "Select project endpoint" '{}'

# Test project status (will likely fail without valid project)
test_endpoint "GET" "/api/projects/test-project/status" "Project status endpoint"

# Test non-existent endpoint
test_endpoint "GET" "/api/nonexistent" "Non-existent endpoint (should return 404)"

echo -e "\n${GREEN}🎉 Endpoint testing completed!${NC}"
echo ""
echo "Next steps:"
echo "1. Test WebSocket connection with: node test-client.js"
echo "2. Test with actual OpenAI API key for voice functionality"
echo "3. Set up projects and test voice commands"