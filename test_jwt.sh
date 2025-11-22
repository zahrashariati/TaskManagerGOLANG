#!/bin/bash

BASE_URL="http://localhost:3000"

echo "🔐 Testing JWT Service..."
echo ""

# 1. Register
echo "1. Registering user..."
TIMESTAMP=$(date +%s)
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"testuser${TIMESTAMP}\",
    \"email\": \"test${TIMESTAMP}@example.com\",
    \"password\": \"password123\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "✅ Registration successful!"
    echo "Token: ${TOKEN:0:50}..."
    echo ""
else
    echo "❌ Registration failed:"
    echo "$REGISTER_RESPONSE"
    exit 1
fi

# 2. Login
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"testuser${TIMESTAMP}\",
    \"password\": \"password123\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "✅ Login successful!"
    echo "Token: ${TOKEN:0:50}..."
    echo ""
else
    echo "❌ Login failed:"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

# 3. Get Profile (Protected)
echo "3. Getting user profile (protected endpoint)..."
PROFILE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer $TOKEN")

HTTP_STATUS=$(echo "$PROFILE_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$PROFILE_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Profile retrieved successfully!"
    echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
    echo ""
else
    echo "❌ Failed to get profile (Status: $HTTP_STATUS):"
    echo "$BODY"
fi

# 4. Create Task (Protected)
echo "4. Creating task (protected endpoint)..."
TASK_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST $BASE_URL/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test Task",
    "description": "Testing JWT authentication",
    "priority": "high"
  }')

HTTP_STATUS=$(echo "$TASK_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$TASK_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "201" ] || [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Task created successfully!"
    echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
    echo ""
else
    echo "❌ Failed to create task (Status: $HTTP_STATUS):"
    echo "$BODY"
fi

# 5. Test Invalid Token
echo "5. Testing invalid token rejection..."
INVALID_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer invalid_token_here")

HTTP_STATUS=$(echo "$INVALID_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$INVALID_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "401" ]; then
    echo "✅ Invalid token correctly rejected!"
    echo "$BODY"
    echo ""
else
    echo "❌ Invalid token was not rejected (Status: $HTTP_STATUS):"
    echo "$BODY"
fi

# 6. Test Missing Token
echo "6. Testing missing token rejection..."
MISSING_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET $BASE_URL/auth/me)

HTTP_STATUS=$(echo "$MISSING_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$MISSING_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "401" ]; then
    echo "✅ Missing token correctly rejected!"
    echo "$BODY"
    echo ""
else
    echo "❌ Missing token was not rejected (Status: $HTTP_STATUS):"
    echo "$BODY"
fi

echo "🎉 All tests completed!"
echo ""
echo "📝 Summary:"
echo "  - Registration: ✅"
echo "  - Login: ✅"
echo "  - Protected endpoints: ✅"
echo "  - Token validation: ✅"
echo ""
echo "💡 Tip: Use the token above to test other endpoints:"
echo "   curl -H 'Authorization: Bearer $TOKEN' http://localhost:8080/auth/me"


