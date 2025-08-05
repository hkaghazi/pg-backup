#!/bin/bash

# Script to manually trigger a backup via HTTP API
# Usage: ./trigger-backup.sh [port]

PORT=${1:-8080}
URL="http://localhost:${PORT}/trigger"

echo "Triggering manual backup via ${URL}..."

response=$(curl -s -X POST "${URL}" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_CODE:%{http_code}")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
json_response=$(echo "$response" | sed '/HTTP_CODE:/d')

case $http_code in
  202)
    echo "✅ Backup triggered successfully!"
    echo "$json_response" | jq . 2>/dev/null || echo "$json_response"
    ;;
  405)
    echo "❌ Method not allowed. Use POST method."
    echo "$json_response" | jq . 2>/dev/null || echo "$json_response"
    ;;
  503)
    echo "❌ Service unavailable. Backup service not initialized."
    echo "$json_response" | jq . 2>/dev/null || echo "$json_response"
    ;;
  *)
    echo "❌ Unexpected response (HTTP $http_code):"
    echo "$json_response"
    ;;
esac
