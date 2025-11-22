# JWT Service Testing Guide 🔐

This guide will help you test your JWT authentication service step by step.

## Prerequisites

1. **Generate RSA Keys** (if not already done):
   ```bash
   chmod +x scripts/generate-keys.sh
   ./scripts/generate-keys.sh
   ```
   
   Or manually:
   ```bash
   mkdir -p keys
   openssl genrsa -out keys/private.pem 2048
   openssl rsa -in keys/private.pem -pubout -out keys/public.pem
   chmod 600 keys/private.pem
   chmod 644 keys/public.pem
   ```

2. **Start your services** (Database, Redis, API):
   ```bash
   docker compose up -d
   ```
   
   Or run locally:
   ```bash
   # Make sure DATABASE_URL and REDIS_URL are set in .env
   go run cmd/api/main.go
   ```

---

## Step 1: Test User Registration ✅

**Endpoint:** `POST /auth/register`

**Request:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Expected Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

**What to check:**
- ✅ Status code is `201`
- ✅ Response contains a `token` field (long JWT string)
- ✅ Response contains `user` object with id, username, email
- ✅ Password is NOT in the response (security!)

---

## Step 2: Test User Login ✅

**Endpoint:** `POST /auth/login`

**Request:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**Expected Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

**What to check:**
- ✅ Status code is `200`
- ✅ Token is different from registration token (new token generated)
- ✅ User information matches

**Test Invalid Credentials:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "wrongpassword"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "Invalid credentials"
}
```

---

## Step 3: Test Protected Endpoints with JWT Token 🔒

**Save your token from login:**
```bash
# Save token to variable (replace YOUR_TOKEN with actual token)
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Test Get User Profile (Protected)

**Endpoint:** `GET /auth/me`

**Request:**
```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "id": 1,
  "username": "testuser",
  "email": "test@example.com"
}
```

**What to check:**
- ✅ Status code is `200`
- ✅ Returns your user information
- ✅ Works only with valid token

### Test Without Token (Should Fail)

```bash
curl -X GET http://localhost:8080/auth/me
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "Invalid authorization header"
}
```

### Test With Invalid Token (Should Fail)

```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer invalid_token_here"
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "Invalid token"
}
```

---

## Step 4: Test Task Operations with JWT 🔐

### Create a Task (Protected)

**Endpoint:** `POST /tasks`

**Request:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My First Task",
    "description": "Test task description",
    "priority": "high"
  }'
```

**Expected Response (201 Created):**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "My First Task",
  "description": "Test task description",
  "completed": false,
  "priority": "high",
  "created_at": "2024-11-17T..."
}
```

### Get All Tasks (Protected)

**Endpoint:** `GET /tasks`

**Request:**
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "title": "My First Task",
    ...
  }
]
```

**What to check:**
- ✅ Only returns tasks belonging to the authenticated user
- ✅ Tasks from other users are not visible

---

## Step 5: Test Token Expiration ⏰

JWT tokens expire after 24 hours. To test expiration, you can:

1. **Decode token** (to see expiration time):
   - Visit https://jwt.io
   - Paste your token
   - Check the `exp` field (Unix timestamp)

2. **Manually expire a token** (for testing):
   - Wait 24 hours, OR
   - Modify the token expiration in `internal/auth/jwt.go` temporarily:
     ```go
     ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Second)), // 1 second for testing
     ```

3. **Test expired token:**
   ```bash
   # Use an expired token
   curl -X GET http://localhost:8080/auth/me \
     -H "Authorization: Bearer EXPIRED_TOKEN"
   ```
   
   **Expected Response (401 Unauthorized):**
   ```json
   {
     "error": "Invalid token"
   }
   ```

---

## Step 6: Test Token Structure 🔍

You can decode your JWT token to verify its structure:

1. **Visit:** https://jwt.io
2. **Paste your token** in the "Encoded" section
3. **Verify:**
   - Algorithm: `RS256`
   - Payload contains:
     - `user_id`: Your user ID
     - `username`: Your username
     - `exp`: Expiration timestamp
     - `iat`: Issued at timestamp
     - `iss`: "task_manager"
     - `sub`: Your user ID as string
     - `aud`: ["task_manager"]

---

## Step 7: Test Multiple Users 👥

### Register Second User

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user2",
    "email": "user2@example.com",
    "password": "password123"
  }'
```

**Save token:**
```bash
TOKEN2="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Create Task for User 2

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN2" \
  -d '{
    "title": "User 2 Task",
    "description": "This belongs to user 2",
    "priority": "medium"
  }'
```

### Verify Data Isolation

**Get tasks with User 1 token:**
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** Only User 1's tasks appear

**Get tasks with User 2 token:**
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN2"
```

**Expected:** Only User 2's tasks appear

**What to check:**
- ✅ Users can only see their own tasks
- ✅ Data isolation works correctly

---

## Step 8: Test Edge Cases 🧪

### 1. Missing Authorization Header
```bash
curl -X GET http://localhost:8080/auth/me
```
**Expected:** `401 Unauthorized` with "Invalid authorization header"

### 2. Malformed Authorization Header
```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: InvalidFormat token"
```
**Expected:** `401 Unauthorized` with "Invalid authorization header"

### 3. Empty Token
```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer "
```
**Expected:** `401 Unauthorized`

### 4. Tampered Token (Modified Payload)
```bash
# Get a valid token, modify it, then use it
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer TAMPERED_TOKEN"
```
**Expected:** `401 Unauthorized` with "Invalid token"

---

## Step 9: Test Update and Delete User 🔄

### Update User (Protected)

**Endpoint:** `PUT /auth/me`

```bash
curl -X PUT http://localhost:8080/auth/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "username": "updateduser",
    "email": "updated@example.com",
    "password": "newpassword123"
  }'
```

**Expected Response (200 OK):**
```json
{
  "message": "User updated successfully"
}
```

### Delete User (Protected)

**Endpoint:** `DELETE /auth/me`

```bash
curl -X DELETE http://localhost:8080/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

---

## Quick Test Script 🚀

Save this as `test_jwt.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🔐 Testing JWT Service..."
echo ""

# 1. Register
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }')

TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "✅ Registration successful!"
echo "Token: ${TOKEN:0:50}..."
echo ""

# 2. Login
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "✅ Login successful!"
echo "Token: ${TOKEN:0:50}..."
echo ""

# 3. Get Profile (Protected)
echo "3. Getting user profile..."
PROFILE_RESPONSE=$(curl -s -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer $TOKEN")
echo "✅ Profile retrieved:"
echo "$PROFILE_RESPONSE" | jq .
echo ""

# 4. Create Task (Protected)
echo "4. Creating task..."
TASK_RESPONSE=$(curl -s -X POST $BASE_URL/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test Task",
    "description": "Testing JWT",
    "priority": "high"
  }')
echo "✅ Task created:"
echo "$TASK_RESPONSE" | jq .
echo ""

# 5. Test Invalid Token
echo "5. Testing invalid token..."
INVALID_RESPONSE=$(curl -s -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer invalid_token")
echo "✅ Invalid token rejected:"
echo "$INVALID_RESPONSE"
echo ""

echo "🎉 All tests completed!"
```

**Run it:**
```bash
chmod +x test_jwt.sh
./test_jwt.sh
```

---

## Troubleshooting 🔧

### Issue: "JWT keys not configured"
**Solution:** Make sure `keys/private.pem` and `keys/public.pem` exist

### Issue: "Failed to initialize JWT service"
**Solution:** Check file permissions and paths in `.env` or config

### Issue: "Invalid token" even with valid token
**Solution:** 
- Verify token hasn't expired
- Check that public/private keys match
- Ensure token is sent as `Bearer TOKEN` (with space)

### Issue: "Invalid authorization header"
**Solution:** Make sure header format is exactly: `Authorization: Bearer YOUR_TOKEN`

---

## Summary Checklist ✅

- [ ] RSA keys generated (`keys/private.pem`, `keys/public.pem`)
- [ ] User registration works and returns JWT token
- [ ] User login works and returns JWT token
- [ ] Protected endpoints require valid JWT token
- [ ] Invalid tokens are rejected (401)
- [ ] Missing tokens are rejected (401)
- [ ] Users can only access their own data
- [ ] Token contains correct user information
- [ ] Token expiration works (after 24 hours)

---

**Happy Testing! 🎉**

