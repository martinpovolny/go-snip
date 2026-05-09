#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"
overall_result=0

BOLD="\033[1m"
RESET="\033[0m"

assert_status() {
  local title=$1
  shift
  local expected=$1
  shift
  local response status body

  echo -e "${BOLD}${title}${RESET}"

  response=$(curl -s -w "\n%{http_code}" "$@")
  status=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  echo "$body"

  if [ "$status" -eq "$expected" ]; then
    echo "✅ $status"
    echo
  else
    echo "❌ expected $expected, got $status"
    echo
    overall_result=1
    return 1
  fi
}

# HTTP codes
export http_ok=200
export http_created=201
export http_invalid_request=400
export http_not_found=404
export http_method_not_allowed=405
export http_internal_error=500

# actual "integration tests"
assert_status "create ok" $http_created -X POST "${BASE_URL}/save" \
  -H "Content-Type: application/json" \
  -d '{
    "external_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john4@example.com",
    "date_of_birth": "1990-01-01T00:00:00+00:00"
  }'

assert_status "create invalid date" $http_invalid_request -X POST "${BASE_URL}/save" \
  -H "Content-Type: application/json" \
  -d '{
    "external_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com",
    "date_of_birth": "random crap"
  }'

assert_status "create missing external_id" $http_internal_error -X POST "${BASE_URL}/save" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "missing external id",
    "email": "john@example.com",
    "date_of_birth": "1990-01-01T00:00:00+00:00"
  }'

assert_status "invalid method" $http_method_not_allowed "${BASE_URL}/save"

assert_status "get existing" $http_ok "${BASE_URL}/123e4567-e89b-12d3-a456-426614174000"

assert_status "get empty id" $http_not_found "${BASE_URL}"

assert_status "get non-existent" $http_not_found "${BASE_URL}/foobar"

assert_status "post to get endpoint" $http_method_not_allowed -X POST "${BASE_URL}/123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "missing external id",
    "email": "john@example.com",
    "date_of_birth": "1990-01-01T00:00:00+00:00"
  }'

# json data comparison
expected='{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"John Doe","email":"john4@example.com","date_of_birth":"1990-01-01T00:00:00Z"}'
response=$(curl -s "$BASE_URL/123e4567-e89b-12d3-a456-426614174000")
expected_normalized=$(echo "$expected" | jq -c 'to_entries | sort_by(.key) | from_entries')
actual_normalized=$(echo "$response" | jq -c 'to_entries | sort_by(.key) | from_entries')

echo -e "${BOLD}json data comparison${RESET}"
if [ "$expected_normalized" = "$actual_normalized" ]; then
  echo "✅ match"
else
  echo "❌ response does not match"
  echo "  expected: $expected_normalized"
  echo "  got:      $actual_normalized"
  overall_result=1
fi

exit $overall_result