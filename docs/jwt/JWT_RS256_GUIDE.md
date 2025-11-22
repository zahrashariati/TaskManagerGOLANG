# JWT Authentication with RS256 - Complete Guide

## 🔐 What is RS256?

**RS256** = RSA Signature with SHA-256

**How it works:**
- **Private Key** = Signs (creates) JWT tokens (server keeps secret)
- **Public Key** = Verifies JWT tokens (can be shared)

**Why RS256?**
- ✅ More secure than HS256 (symmetric)
- ✅ Public key can be shared (for microservices)
- ✅ Private key stays secret on server

---

## 🔑 Where Keys Are Used

### Private Key (Server - Keep Secret!)
**Location:** `keys/private.pem` (never commit to git!)

**Used for:**
- **Signing tokens** (when user logs in)
- Creating JWT tokens
- Only server has this key

**Example:**
```go
// When user logs in
token := jwt.Sign(privateKey, userData)
```

### Public Key (Can Be Shared)
**Location:** `keys/public.pem` (can be public)

**Used for:**
- **Verifying tokens** (on every request)
- Validating JWT tokens
- Can be shared with other services

**Example:**
```go
// On every protected route
userData := jwt.Verify(publicKey, token)
```

---

## 📁 File Structure

```
booking_ticket/
├── keys/
│   ├── private.pem    # Private key (KEEP SECRET!)
│   └── public.pem      # Public key (can share)
├── internal/
│   ├── auth/
│   │   ├── jwt.go          # JWT signing/verification
│   │   ├── middleware.go   # JWT middleware
│   │   └── service.go      # Auth service
│   ├── handlers/
│   │   └── auth_handler.go # Login/Register handlers
│   └── models/
│       └── user.go          # User model
```

---

## 🔄 Flow Explanation

### 1. User Registration/Login
```
User → POST /auth/login
  → Auth Service validates credentials
  → Auth Service signs JWT with PRIVATE KEY
  → Returns JWT token to user
```

### 2. Protected Request
```
User → GET /tasks (with JWT in header)
  → JWT Middleware intercepts
  → Middleware verifies JWT with PUBLIC KEY
  → If valid → Continue to handler
  → If invalid → Return 401 Unauthorized
```

---

## 🛠️ Implementation Steps

1. **Generate RSA keys** (one-time setup)
2. **Create JWT service** (sign/verify tokens)
3. **Create auth middleware** (protect routes)
4. **Create auth handlers** (login/register)
5. **Protect routes** (add middleware)

---

## 📝 Key Generation Commands

```bash
# Generate private key
openssl genrsa -out keys/private.pem 2048

# Extract public key from private key
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# Verify keys
cat keys/private.pem
cat keys/public.pem
```

---

## 🔒 Security Best Practices

1. **Never commit private key to git**
   - Add `keys/private.pem` to `.gitignore`
   - Keep it secret!

2. **Use environment variables for key paths**
   - `JWT_PRIVATE_KEY_PATH=keys/private.pem`
   - `JWT_PUBLIC_KEY_PATH=keys/public.pem`

3. **Set proper file permissions**
   ```bash
   chmod 600 keys/private.pem  # Only owner can read/write
   chmod 644 keys/public.pem   # Readable by all
   ```

4. **In production:**
   - Store keys in secure vault (AWS Secrets Manager, etc.)
   - Use environment variables
   - Never hardcode keys

---

## 📚 Next Steps

Follow the implementation files to add JWT authentication to your API!


