# JWT Commands Cheat Sheet 📝

Quick reference for all JWT-related commands.

---

## 🔑 Key Generation

```bash
# Create keys directory
mkdir -p keys

# Generate private key (2048 bits)
openssl genrsa -out keys/private.pem 2048

# Extract public key
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# Set permissions
chmod 600 keys/private.pem   # Private: owner only
chmod 644 keys/public.pem   # Public: readable

# OR use script (recommended)
chmod +x scripts/generate-keys.sh
./scripts/generate-keys.sh
```

---

## 🐳 Docker Commands

```bash
# Start all services
docker compose up -d

# Rebuild API (after code changes)
docker compose build api
docker compose restart api

# View API logs
docker compose logs api --tail 20

# View all logs
docker compose logs -f

# Check container status
docker compose ps

# Stop all services
docker compose down

# Restart API only
docker compose restart api
```

---

## 🧪 Testing Commands

### Register User:
```bash
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

### Login:
```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### Get Profile (Protected):
```bash
curl -X GET http://localhost:3000/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Create Task (Protected):
```bash
curl -X POST http://localhost:3000/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"title":"My Task","description":"Task description","priority":"high"}'
```

### Get All Tasks (Protected):
```bash
curl -X GET http://localhost:3000/tasks \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Run Test Script:
```bash
./test_jwt.sh
```

---

## 🔍 Verification Commands

### Check Keys Exist:
```bash
ls -la keys/
```

### View Private Key Info:
```bash
openssl rsa -in keys/private.pem -text -noout | head -20
```

### View Public Key Info:
```bash
openssl rsa -in keys/public.pem -pubin -text -noout | head -20
```

### Decode JWT Token:
Visit: https://jwt.io and paste your token

---

## 📦 Go Commands

### Install Dependencies:
```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

### Sync Dependencies:
```bash
go mod tidy
```

### Build:
```bash
go build ./cmd/api
```

### Build All:
```bash
go build ./...
```

### Run Locally:
```bash
go run cmd/api/main.go
```

### Check for Errors:
```bash
go vet ./...
```

---

## 🚀 Startup Sequence

```bash
# 1. Navigate to project
cd ~/Desktop/task_manager

# 2. Generate keys (first time only)
./scripts/generate-keys.sh

# 3. Start services
docker compose up -d

# 4. Wait ~10 seconds, then verify
docker compose ps

# 5. Test API
curl http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"test123"}'
```

---

## 🔧 Troubleshooting Commands

### Check API Logs:
```bash
docker compose logs api --tail 50
```

### Check Database:
```bash
docker compose exec cockroachdb cockroach sql --insecure --execute="SHOW DATABASES;"
```

### Check Redis:
```bash
docker compose exec redis redis-cli ping
```

### Rebuild Everything:
```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

### Check Keys in Container:
```bash
docker compose exec api ls -la /app/keys/
```

---

## 📋 Environment Variables

### Set in docker-compose.yml:
```yaml
environment:
  - DATABASE_URL=postgresql://root@cockroachdb:26257/taskmanager?sslmode=disable
  - REDIS_URL=redis:6379
  - PORT=8080
  - JWT_PRIVATE_KEY_PATH=/app/keys/private.pem
  - JWT_PUBLIC_KEY_PATH=/app/keys/public.pem
```

---

## 🎯 Quick Test Sequence

```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"test123"}' \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# 2. Use token
curl -X GET http://localhost:3000/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

**Save this for quick reference!** 📌

