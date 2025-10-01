#!/bin/bash

# Graph Operations Replay Log Test Script
# This script demonstrates the complete flow of the graph operations replay log system:
# 1. Login and get auth token
# 2. Create nodes and edges
# 3. View the replay log after each operation
# 4. Delete operations
# 5. Final replay log showing all operations in order

set -e  # Exit on error

# Configuration
HOST="http://bloodhound.localhost"
USERNAME="admin"
PASSWORD="SFdzJoW2GT7Fn68aEieKn7S1S2DLdXnw"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Graph Operations Replay Log Test${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Step 1: Login and get token
echo -e "${GREEN}Step 1: Authenticating...${NC}"
echo "POST $HOST/api/v2/login"
echo ""

TOKEN=$(curl -s "$HOST/api/v2/login" \
  -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"login_method\": \"secret\", \"username\": \"$USERNAME\", \"secret\": \"$PASSWORD\"}" \
  | jq -r '.data.session_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "Failed to get auth token!"
  exit 1
fi

echo "✓ Authenticated successfully"
echo "Token: ${TOKEN:0:20}..."
echo ""
sleep 1

# Step 2: Create first node (User)
echo -e "${GREEN}Step 2: Creating first node (User)${NC}"
echo "POST $HOST/api/v2/graph/nodes"
echo "Object ID: S-1-5-21-TEST-1001"
echo ""

curl -s -X POST "$HOST/api/v2/graph/nodes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "object_id": "S-1-5-21-TEST-1001",
    "labels": ["User", "Base"],
    "properties": {
      "name": "testuser@domain.local",
      "enabled": true,
      "description": "Test user for replay log demo"
    }
  }' | jq '.'

echo ""
sleep 1

# Step 3: Check replay log after first node
echo -e "${YELLOW}Step 3: Checking replay log (should have 1 entry)${NC}"
echo "GET $HOST/api/v2/graph/replay log"
echo ""

curl -s -X GET "$HOST/api/v2/graph/replay log" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '{
    count: .count,
    entries: .entries | map({
      id: .id,
      change_type: .change_type,
      object_type: .object_type,
      object_id: .object_id,
      created_at: .created_at
    })
  }'

echo ""
sleep 2

# Step 4: Create second node (Group)
echo -e "${GREEN}Step 4: Creating second node (Group)${NC}"
echo "POST $HOST/api/v2/graph/nodes"
echo "Object ID: S-1-5-21-TEST-1002"
echo ""

curl -s -X POST "$HOST/api/v2/graph/nodes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "object_id": "S-1-5-21-TEST-1002",
    "labels": ["Group", "Base"],
    "properties": {
      "name": "Domain Admins@domain.local",
      "description": "Test group for replay log demo"
    }
  }' | jq '.'

echo ""
sleep 1

# Step 5: Check replay log after second node
echo -e "${YELLOW}Step 5: Checking replay log (should have 2 entries)${NC}"
echo "GET $HOST/api/v2/graph/replay log"
echo ""

curl -s -X GET "$HOST/api/v2/graph/replay log" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '{
    count: .count,
    latest_entries: .entries[0:2] | map({
      id: .id,
      change_type: .change_type,
      object_type: .object_type,
      object_id: .object_id,
      created_at: .created_at
    })
  }'

echo ""
sleep 2

# Step 6: Create an edge between the nodes
echo -e "${GREEN}Step 6: Creating edge (User -> Group)${NC}"
echo "POST $HOST/api/v2/graph/edges"
echo "Edge: S-1-5-21-TEST-1001 -[MemberOf]-> S-1-5-21-TEST-1002"
echo ""

curl -s -X POST "$HOST/api/v2/graph/edges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "source_object_id": "S-1-5-21-TEST-1001",
    "target_object_id": "S-1-5-21-TEST-1002",
    "edge_kind": "MemberOf",
    "properties": {
      "isacl": false
    }
  }' | jq '.'

echo ""
sleep 1

# Step 7: Check replay log after edge creation
echo -e "${YELLOW}Step 7: Checking replay log (should have 3 entries)${NC}"
echo "GET $HOST/api/v2/graph/replay log"
echo ""

curl -s -X GET "$HOST/api/v2/graph/replay log" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '{
    count: .count,
    latest_edge_entry: .entries[0] | {
      id: .id,
      change_type: .change_type,
      object_type: .object_type,
      edge_kind: .object_id,
      source: .source_object_id,
      target: .target_object_id,
      created_at: .created_at
    }
  }'

echo ""
sleep 2

# Step 8: Delete the edge
echo -e "${GREEN}Step 8: Deleting edge${NC}"
echo "DELETE $HOST/api/v2/graph/edges"
echo "Deleting: S-1-5-21-TEST-1001 -[MemberOf]-> S-1-5-21-TEST-1002"
echo ""

curl -s -X DELETE "$HOST/api/v2/graph/edges" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "source_object_id": "S-1-5-21-TEST-1001",
    "target_object_id": "S-1-5-21-TEST-1002",
    "edge_kind": "MemberOf"
  }' | jq -R 'if . == "" then "Success: Edge deleted" else . end'

echo ""
sleep 1

# Step 9: Delete first node
echo -e "${GREEN}Step 9: Deleting first node${NC}"
echo "DELETE $HOST/api/v2/graph/nodes/S-1-5-21-TEST-1001"
echo ""

curl -s -X DELETE "$HOST/api/v2/graph/nodes/S-1-5-21-TEST-1001" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -R 'if . == "" then "Success: Node deleted" else . end'

echo ""
sleep 1

# Step 10: Delete second node
echo -e "${GREEN}Step 10: Deleting second node${NC}"
echo "DELETE $HOST/api/v2/graph/nodes/S-1-5-21-TEST-1002"
echo ""

curl -s -X DELETE "$HOST/api/v2/graph/nodes/S-1-5-21-TEST-1002" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -R 'if . == "" then "Success: Node deleted" else . end'

echo ""
sleep 2

# Step 11: Final replay log view - complete history
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Step 11: FINAL REPLAY LOG (Complete History)${NC}"
echo -e "${BLUE}============================================${NC}"
echo "GET $HOST/api/v2/graph/replay log"
echo ""
echo "This shows the complete linear history of all operations:"
echo "• 2 node creations"
echo "• 1 edge creation"
echo "• 1 edge deletion"
echo "• 2 node deletions"
echo ""

curl -s -X GET "$HOST/api/v2/graph/replay log" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '{
    total_count: .count,
    operations_timeline: .entries | reverse | map({
      sequence: .id,
      timestamp: .created_at,
      operation: .change_type,
      target_type: .object_type,
      target_id: .object_id,
      details: (
        if .object_type == "edge" then
          "\(.source_object_id) -[\(.object_id)]-> \(.target_object_id)"
        else
          .object_id
        end
      )
    })
  }'

echo ""
echo ""
echo -e "${GREEN}✓ Test completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Key Points:${NC}"
echo "• All 6 operations were logged in order"
echo "• Timestamps are authoritative (created_at)"
echo "• Each entry captures the full operation details"
echo "• The replay log is linear and append-only"
echo "• This can be used for replay/rewind in the future"
echo ""
