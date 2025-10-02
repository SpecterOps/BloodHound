#!/bin/bash

HOST="http://bloodhound.localhost"
USERNAME="admin"
PASSWORD="SFdzJoW2GT7Fn68aEieKn7S1S2DLdXnw"

# Login
echo "1. Logging in..."
TOKEN=$(curl -s "$HOST/api/v2/login" \
  -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"login_method\": \"secret\", \"username\": \"$USERNAME\", \"secret\": \"$PASSWORD\"}" \
  | jq -r '.data.session_token')
echo "Token: ${TOKEN:0:20}..."
echo ""

# Get current nodes
echo "2. Getting current nodes..."
curl -s -X POST "$HOST/api/v2/graphs/cypher" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "MATCH (n) RETURN n"}' | jq '.data.nodes | length'
echo ""

# Add a node
UNIQUE_ID=$(date +%s)
echo "3. Adding a new node (S-1-5-21-SIMPLE-TEST-$UNIQUE_ID)..."
curl -s -X POST "$HOST/api/v2/graph/nodes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"object_id\": \"S-1-5-21-SIMPLE-TEST-$UNIQUE_ID\",
    \"labels\": [\"User\", \"Base\"],
    \"properties\": {
      \"name\": \"simpletest-$UNIQUE_ID@domain.local\"
    }
  }" | jq '.'
echo ""

# Get nodes again
echo "4. Getting nodes after adding new node..."
curl -s -X POST "$HOST/api/v2/graphs/cypher" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "MATCH (n) RETURN n"}' | jq '.data.nodes | length'
echo ""

# Show the new node
echo "5. Showing the new node..."
curl -s -X POST "$HOST/api/v2/graphs/cypher" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"query\": \"MATCH (n) WHERE n.objectid = \\\"S-1-5-21-SIMPLE-TEST-$UNIQUE_ID\\\" RETURN n\"}" | jq '.data.nodes'

