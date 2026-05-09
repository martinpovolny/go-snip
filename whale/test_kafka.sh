#!/usr/bin/env bash
set -euo pipefail

PROXY_URL="${PROXY_URL:-http://localhost:18082}"
WHALE_URL="${WHALE_URL:-http://localhost:8080}"
TOPIC="${TOPIC:-people}"
EXTERNAL_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

echo "Posting person to Kafka topic '$TOPIC' via Pandaproxy..."
echo "  external_id: $EXTERNAL_ID"

curl -sf -X POST "$PROXY_URL/topics/$TOPIC" \
  -H "Content-Type: application/vnd.kafka.json.v2+json" \
  -d "{
    \"records\": [
      {
        \"value\": {
          \"external_id\": \"$EXTERNAL_ID\",
          \"name\": \"Test Person\",
          \"email\": \"test@example.com\",
          \"date_of_birth\": \"1990-01-01T00:00:00+00:00\"
        }
      }
    ]
  }" | jq .

echo ""
echo "Waiting for whale to consume and store the record..."
sleep 2

echo "Fetching person from whale API..."
curl -sf "$WHALE_URL/$EXTERNAL_ID" | jq .
