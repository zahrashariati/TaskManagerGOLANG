# Testing Guide: Tasks Endpoints with ATK & RTK 🔐

Complete guide to test protected endpoints using Access Tokens (ATK) and Refresh Tokens (RTK).

---

## 📋 Prerequisites

1. **Server running**: `docker compose up -d`
2. **Postman** (or any HTTP client)
3. **User account** (register first if needed)

---

## 🔑 Step 1: Get Tokens (Login)

### Request:
```
POST http://localhost:8080/auth/login
Content-Type: application/json

Body:
{
  "username": "mohsen",
  "password": "your_password"
}
```

### Response:
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "base64-encoded-token-here",
  "user": {
    "id": 1125314717191503873,
    "username": "mohsen",
    "email": "mohsen@mohsen.com"
  }
}
```

### ✅ Save These:
- **`token`** → This is your **ATK (Access Token)** - use for protected endpoints
- **`refresh_token`** → This is your **RTK (Refresh Token)** - use to get new ATK when expired

---

## 📝 Step 2: Use ATK for Protected Endpoints

All task endpoints require the ATK in the Authorization header.

### Format:
```
Authorization: Bearer <your_access_token>
```

---

## 🎯 Task Endpoints Testing

### 1. **GET /tasks** - Get All Tasks

**Request:**
```
GET http://localhost:8080/tasks
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Query Parameters (Optional):**
- `?showCompleted=true` - Include completed tasks
- `?showCompleted=false` - Exclude completed tasks (default)

**Example:**
```
GET http://localhost:8080/tasks?showCompleted=true
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1125314717191503873,
    "title": "My Task",
    "description": "Task description",
    "completed": false,
    "priority": "high",
    "created_at": "2025-11-18T18:00:00Z"
  }
]
```

---

### 2. **POST /tasks** - Create Task

**Request:**
```
POST http://localhost:8080/tasks
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

Body:
{
  "title": "Complete project",
  "description": "Finish the task manager API",
  "priority": "high"
}
```

**Priority Options:** `"low"`, `"medium"`, `"high"` (default: `"medium"`)

**Response:**
```json
{
  "id": 1,
  "user_id": 1125314717191503873,
  "title": "Complete project",
  "description": "Finish the task manager API",
  "completed": false,
  "priority": "high",
  "created_at": "2025-11-18T18:00:00Z"
}
```

---

### 3. **GET /tasks/:id** - Get Task by ID

**Request:**
```
GET http://localhost:8080/tasks/1
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1125314717191503873,
  "title": "Complete project",
  "description": "Finish the task manager API",
  "completed": false,
  "priority": "high",
  "created_at": "2025-11-18T18:00:00Z"
}
```

---

### 4. **PUT /tasks/:id** - Update Task

**Request:**
```
PUT http://localhost:8080/tasks/1
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

Body:
{
  "title": "Updated title",
  "description": "Updated description",
  "priority": "medium"
}
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1125314717191503873,
  "title": "Updated title",
  "description": "Updated description",
  "completed": false,
  "priority": "medium",
  "created_at": "2025-11-18T18:00:00Z"
}
```

---

### 5. **DELETE /tasks/:id** - Delete Task

**Request:**
```
DELETE http://localhost:8080/tasks/1
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "message": "Task deleted successfully"
}
```

---

### 6. **PATCH /tasks/:id/complete** - Mark Task as Complete

**Request:**
```
PATCH http://localhost:8080/tasks/1/complete
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1125314717191503873,
  "title": "Complete project",
  "description": "Finish the task manager API",
  "completed": true,
  "priority": "high",
  "created_at": "2025-11-18T18:00:00Z"
}
```

---

## 🔄 Step 3: Refresh ATK When Expired

When your ATK expires (after 5 minutes), you'll get a `401 Unauthorized` error. Use RTK to get a new ATK.

### Request:
```
POST http://localhost:8080/auth/refresh
Content-Type: application/json

Body:
{
  "refresh_token": "your_refresh_token_here"
}
```

### Response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "new_refresh_token_here"
}
```

### ⚠️ Important:
- **Old RTK is revoked** - you can't use it again
- **Save the new RTK** - use it for the next refresh
- **Use the new ATK** - for protected endpoints

---

## 🚪 Step 4: Logout (Revoke RTK)

When you want to log out, revoke your refresh token:

### Request:
```
POST http://localhost:8080/auth/logout
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

Body:
{
  "refresh_token": "your_refresh_token_here"
}
```

### Response:
```json
{
  "message": "Logged out successfully"
}
```

After logout:
- ✅ RTK is revoked (can't be used anymore)
- ⚠️ ATK still valid until expiration (5 minutes)
- 🔒 User needs to login again to get new tokens

---

## 📱 Postman Setup (Recommended)

### Option 1: Manual Setup

1. **Login Request:**
   - Method: `POST`
   - URL: `http://localhost:8080/auth/login`
   - Headers: `Content-Type: application/json`
   - Body: `{ "username": "...", "password": "..." }`
   - Copy `token` and `refresh_token` from response

2. **Task Requests:**
   - Method: `GET/POST/PUT/DELETE/PATCH`
   - URL: `http://localhost:8080/tasks` (or `/tasks/:id`)
   - Headers:
     - `Content-Type: application/json` (for POST/PUT)
     - `Authorization: Bearer <paste_token_here>`

3. **Refresh Request:**
   - Method: `POST`
   - URL: `http://localhost:8080/auth/refresh`
   - Headers: `Content-Type: application/json`
   - Body: `{ "refresh_token": "<paste_refresh_token>" }`
   - Update your saved token with the new `access_token`

---

### Option 2: Use Postman Environment Variables

1. **Create Environment:**
   - Click "Environments" → "Create Environment"
   - Name: "Task Manager Local"
   - Add variables:
     - `access_token` (leave empty)
     - `refresh_token` (leave empty)
     - `base_url` = `http://localhost:8080`

2. **Login Request:**
   - Add Test Script:
   ```javascript
   if (pm.response.code === 200) {
       const jsonData = pm.response.json();
       pm.environment.set("access_token", jsonData.token);
       pm.environment.set("refresh_token", jsonData.refresh_token);
   }
   ```

3. **Task Requests:**
   - URL: `{{base_url}}/tasks`
   - Authorization Header: `Bearer {{access_token}}`

4. **Refresh Request:**
   - Add Test Script to save new tokens:
   ```javascript
   if (pm.response.code === 200) {
       const jsonData = pm.response.json();
       pm.environment.set("access_token", jsonData.access_token);
       pm.environment.set("refresh_token", jsonData.refresh_token);
   }
   ```

---

## 🧪 Complete Testing Flow

### 1. Register/Login
```
POST /auth/login
→ Save: access_token, refresh_token
```

### 2. Create Task
```
POST /tasks
Headers: Authorization: Bearer {access_token}
Body: { "title": "...", "description": "...", "priority": "high" }
```

### 3. Get All Tasks
```
GET /tasks
Headers: Authorization: Bearer {access_token}
```

### 4. Get Task by ID
```
GET /tasks/1
Headers: Authorization: Bearer {access_token}
```

### 5. Update Task
```
PUT /tasks/1
Headers: Authorization: Bearer {access_token}
Body: { "title": "...", "description": "...", "priority": "medium" }
```

### 6. Complete Task
```
PATCH /tasks/1/complete
Headers: Authorization: Bearer {access_token}
```

### 7. Delete Task
```
DELETE /tasks/1
Headers: Authorization: Bearer {access_token}
```

### 8. When ATK Expires (after 5 min)
```
POST /auth/refresh
Body: { "refresh_token": "{refresh_token}" }
→ Save: new access_token, new refresh_token
```

### 9. Logout
```
POST /auth/logout
Headers: Authorization: Bearer {access_token}
Body: { "refresh_token": "{refresh_token}" }
```

---

## ❌ Common Errors

### 401 Unauthorized
- **Cause**: ATK expired or invalid
- **Solution**: Use `/auth/refresh` to get new ATK

### 403 Forbidden
- **Cause**: No token provided or invalid token
- **Solution**: Check Authorization header format: `Bearer <token>`

### 400 Bad Request
- **Cause**: Invalid JSON body or missing fields
- **Solution**: Check request body format

### 500 Internal Server Error
- **Cause**: Server error
- **Solution**: Check server logs: `docker compose logs api`

---

## ✅ Quick Test Checklist

- [ ] Login → Get ATK and RTK
- [ ] Create task with ATK
- [ ] Get all tasks with ATK
- [ ] Get task by ID with ATK
- [ ] Update task with ATK
- [ ] Complete task with ATK
- [ ] Delete task with ATK
- [ ] Refresh ATK when expired
- [ ] Logout (revoke RTK)

---

## 🎯 Tips

1. **ATK expires in 5 minutes** - refresh before it expires
2. **RTK expires in 7 days** - but gets revoked on login/logout
3. **Always use Bearer token format**: `Authorization: Bearer <token>`
4. **Save new RTK after refresh** - old one is revoked
5. **Test with different users** - tasks are user-specific

---

Happy Testing! 🚀



