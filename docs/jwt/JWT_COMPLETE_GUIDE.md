# Complete JWT Implementation Guide 📚

This guide explains everything about JWT authentication in your Task Manager API - from key generation to token usage.

---

## 🔐 Part 1: What is JWT?

### JWT = JSON Web Token

**What it is:**
- A secure way to send user information between client and server
- Self-contained (includes user info + expiration)
- Signed with cryptographic keys (can't be tampered with)

**Structure:**
```
JWT Token = Header.Payload.Signature

Example:
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImpvaG4ifQ.signature_here
```

**Parts:**
1. **Header**: Algorithm (RS256) + token type (JWT)
2. **Payload**: User data (user_id, username, expiration, etc.)
3. **Signature**: Cryptographic signature (proves token is authentic)

---

## 🔑 Part 2: RSA Key Generation

### What are RSA Keys?

**RSA** = Rivest-Shamir-Adleman (cryptographic algorithm)

**Two keys:**
- **Private Key**: Secret key (only server knows) → Used to **sign** (create) tokens
- **Public Key**: Can be shared → Used to **verify** (check) tokens

**Why RSA?**
- **Asymmetric**: Different keys for signing vs verifying
- **Secure**: Can't forge tokens without private key
- **Better than HS256**: More secure than symmetric keys

### Commands Used:

#### 1. Generate Private Key:
```bash
openssl genrsa -out keys/private.pem 2048
```

**What this does:**
- Creates RSA private key (2048 bits = secure)
- Saves to `keys/private.pem`
- Format: PKCS#8 (modern OpenSSL default)

#### 2. Extract Public Key:
```bash
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

**What this does:**
- Extracts public key from private key
- Saves to `keys/public.pem`
- Format: PKIX/X.509 (standard format)

#### 3. Set Permissions (Security):
```bash
chmod 600 keys/private.pem   # Only owner can read/write
chmod 644 keys/public.pem    # Readable by all
```

**Why:**
- Private key must be secret (600 = owner only)
- Public key can be shared (644 = readable)

#### 4. Using the Script:
```bash
chmod +x scripts/generate-keys.sh
./scripts/generate-keys.sh
```

**What the script does:**
- Creates `keys/` directory if needed
- Generates private key (2048 bits)
- Extracts public key
- Sets proper permissions
- Shows success message

---

## 📁 Part 3: File Structure

### Key Files:
```
keys/
├── private.pem    # Private key (KEEP SECRET!)
└── public.pem     # Public key (can be shared)

internal/auth/
├── jwt.go         # JWT service (generate/validate tokens)
└── middleware.go  # JWT middleware (protect routes)
```

---

## 💻 Part 4: Code Implementation

### 4.1 Loading Keys (`internal/auth/jwt.go`)

**Function:** `NewJWTService(privatekeyPath, publickeyPath string)`

**What it does:**
1. Reads private key file
2. Decodes PEM format
3. Parses RSA private key (supports PKCS#1 and PKCS#8)
4. Reads public key file
5. Decodes PEM format
6. Parses RSA public key (PKIX/X.509 format)
7. Returns JWTService with both keys

**Code:**
```go
func NewJWTService(privatekeyPath, publickeyPath string) (*JWTService, error) {
    // Read private key
    privateKeyBytes, err := os.ReadFile(privatekeyPath)
    block, _ := pem.Decode(privateKeyBytes)
    
    // Parse private key (try PKCS#1, then PKCS#8)
    privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
        privateKey, _ = key.(*rsa.PrivateKey)
    }
    
    // Read public key
    publicKeyBytes, err := os.ReadFile(publickeyPath)
    block, _ = pem.Decode(publicKeyBytes)
    
    // Parse public key (PKIX format)
    key, err := x509.ParsePKIXPublicKey(block.Bytes)
    publicKey, _ := key.(*rsa.PublicKey)
    
    return &JWTService{
        privateKey: privateKey,
        publicKey: publicKey,
    }, nil
}
```

---

### 4.2 Generating Tokens (`internal/auth/jwt.go`)

**Function:** `GenerateToken(userId int, username string)`

**What it does:**
1. Creates Claims (user data + expiration)
2. Creates JWT token with RS256 algorithm
3. Signs token with private key
4. Returns token string

**Code:**
```go
func (s *JWTService) GenerateToken(userId int, username string) (string, error) {
    claims := &Claims{
        UserID: userId,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            Issuer: "task_manager",
            Subject: fmt.Sprintf("%d", userId),
            Audience: jwt.ClaimStrings{"task_manager"},
            NotBefore: jwt.NewNumericDate(time.Now()),
            ID: uuid.New().String(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    tokenString, err := token.SignedString(s.privateKey)
    return tokenString, nil
}
```

**What's in the token:**
- `user_id`: User's database ID
- `username`: User's username
- `exp`: Expiration time (24 hours from now)
- `iat`: Issued at time
- `iss`: Issuer ("task_manager")
- `sub`: Subject (user ID as string)
- `aud`: Audience (["task_manager"])
- `nbf`: Not before time
- `jti`: Unique token ID (UUID)

---

### 4.3 Validating Tokens (`internal/auth/jwt.go`)

**Function:** `ValidateToken(tokenString string)`

**What it does:**
1. Parses token string
2. Verifies signature with public key
3. Checks expiration
4. Returns Claims (user data)

**Code:**
```go
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify algorithm is RS256
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("invalid signing method")
        }
        return s.publicKey, nil  // Use public key to verify
    })
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}
```

---

### 4.4 Middleware (`internal/auth/middleware.go`)

**Function:** `JWTMiddleware(jwtService *JWTService)`

**What it does:**
1. Extracts token from `Authorization` header
2. Validates token
3. Stores user info in context (`c.Locals()`)
4. Allows request to continue OR returns 401

**Code:**
```go
func JWTMiddleware(jwtService *JWTService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get token from header
        tokenString, err := jwtService.GetTokenFromHeader(c)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization header"})
        }
        
        // Validate token
        claims, err := jwtService.ValidateToken(tokenString)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
        }
        
        // Store user info in context
        c.Locals("userID", claims.UserID)
        c.Locals("username", claims.Username)
        
        return c.Next()  // Continue to handler
    }
}
```

---

## 🔄 Part 5: Authentication Flow

### Registration Flow:

```
1. Client → POST /auth/register
   Body: {username, email, password}

2. Handler → Parse request
   → Call authService.Register()
   → Hash password (bcrypt)
   → Save to database
   → Get user ID

3. Handler → Generate JWT token
   → jwtService.GenerateToken(userID, username)
   → Sign with private key
   → Return token + user info

4. Client ← Receives token
   → Saves token (Postman environment)
   → Uses token for protected requests
```

### Login Flow:

```
1. Client → POST /auth/login
   Body: {username, password}

2. Handler → Parse request
   → Call authService.Login()
   → Get user from database
   → Compare password (bcrypt)
   → Return user

3. Handler → Generate NEW JWT token
   → jwtService.GenerateToken(userID, username)
   → Sign with private key
   → Return token + user info

4. Client ← Receives NEW token
   → Updates saved token
   → Uses new token for requests
```

### Protected Request Flow:

```
1. Client → GET /tasks
   Header: Authorization: Bearer <token>

2. Middleware → Intercepts request
   → Extracts token from header
   → Validates token (checks signature + expiration)
   → Extracts userID from token
   → Stores in context: c.Locals("userID", userID)

3. Handler → Gets userID from context
   → userID := c.Locals("userID").(int)
   → Calls service with userID
   → Returns only user's tasks

4. Client ← Receives tasks
```

---

## 🛠️ Part 6: All Commands Used

### Key Generation:
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

# OR use script
chmod +x scripts/generate-keys.sh
./scripts/generate-keys.sh
```

### Docker Commands:
```bash
# Start all services
docker compose up -d

# Rebuild API after code changes
docker compose build api
docker compose restart api

# View logs
docker compose logs api --tail 20

# Check status
docker compose ps

# Stop everything
docker compose down
```

### Go Commands:
```bash
# Install dependencies
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt

# Sync dependencies
go mod tidy

# Build
go build ./cmd/api

# Run locally
go run cmd/api/main.go
```

### Testing Commands:
```bash
# Test registration
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"pass123"}'

# Test login
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"pass123"}'

# Test protected endpoint
curl -X GET http://localhost:3000/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Run test script
./test_jwt.sh
```

---

## 📊 Part 7: Key Formats Explained

### Private Key Formats:

**PKCS#1** (Traditional):
```
-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----
```

**PKCS#8** (Modern - what OpenSSL generates):
```
-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
```

**Our code supports both!**

### Public Key Formats:

**PKCS#1** (RSA-specific):
```
-----BEGIN RSA PUBLIC KEY-----
...
-----END RSA PUBLIC KEY-----
```

**PKIX/X.509** (Standard - what `openssl rsa -pubout` generates):
```
-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----
```

**Our code uses PKIX (standard format)**

---

## 🔒 Part 8: Security Details

### How Signing Works:

1. **Server creates token** with user data
2. **Server signs** with private key → Creates signature
3. **Token sent** to client: `Header.Payload.Signature`

### How Verification Works:

1. **Client sends** token in Authorization header
2. **Server extracts** token
3. **Server verifies** signature using public key
4. **If signature valid** → Token is authentic
5. **If signature invalid** → Token is fake/rejected

### Why It's Secure:

- **Can't forge**: Need private key to create valid tokens
- **Can't modify**: Changing payload breaks signature
- **Expires**: Tokens expire after 24 hours
- **Unique**: Each token has unique ID (`jti`)

---

## 📝 Part 9: Configuration

### Environment Variables:

```bash
# In docker-compose.yml or .env
JWT_PRIVATE_KEY_PATH=/app/keys/private.pem
JWT_PUBLIC_KEY_PATH=/app/keys/public.pem
```

### Config Loading (`internal/config/config.go`):

```go
JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
```

**Default paths:** `keys/private.pem` and `keys/public.pem`

---

## 🎯 Part 10: Complete Flow Diagram

```
┌─────────┐
│ Client  │
└────┬────┘
     │ POST /auth/register
     ├─────────────────────┐
     │                     │
     ▼                     ▼
┌─────────────┐    ┌──────────────┐
│   Handler   │───▶│ Auth Service │
└─────┬───────┘    └──────┬───────┘
      │                   │
      │                   ▼
      │            ┌──────────────┐
      │            │   Database   │
      │            │  (Save User) │
      │            └──────┬───────┘
      │                   │
      │                   │ Returns userID
      │                   │
      ▼                   │
┌─────────────┐          │
│ JWT Service │◀─────────┘
└─────┬───────┘
      │ GenerateToken(userID, username)
      │ Sign with private key
      │
      ▼
┌─────────────┐
│   Token     │
│ (JWT String)│
└─────┬───────┘
      │
      │ Return to client
      │
      ▼
┌─────────┐
│ Client  │ (Saves token)
└─────────┘

─────────────────────────────────────

┌─────────┐
│ Client  │
└────┬────┘
     │ GET /tasks
     │ Header: Authorization: Bearer <token>
     │
     ▼
┌─────────────┐
│ Middleware │
└─────┬──────┘
      │ Extract token
      │
      ▼
┌─────────────┐
│ JWT Service │
└─────┬───────┘
      │ ValidateToken(token)
      │ Verify signature with public key
      │
      ▼
┌─────────────┐
│   Claims   │ (userID, username)
└─────┬───────┘
      │
      │ Store in context
      │ c.Locals("userID", userID)
      │
      ▼
┌─────────────┐
│   Handler   │
└─────┬───────┘
      │ Get userID from context
      │ Call service with userID
      │
      ▼
┌─────────────┐
│   Service   │
└─────┬───────┘
      │ Get tasks WHERE user_id = userID
      │
      ▼
┌─────────────┐
│  Database   │
└─────┬───────┘
      │
      │ Return tasks
      │
      ▼
┌─────────┐
│ Client  │ (Receives tasks)
└─────────┘
```

---

## 📋 Part 11: Summary Checklist

### Key Generation:
- [x] Generate RSA private key (2048 bits)
- [x] Extract public key from private key
- [x] Set proper file permissions
- [x] Keys stored in `keys/` directory

### Code Implementation:
- [x] Load keys in `NewJWTService()`
- [x] Support PKCS#8 private keys
- [x] Support PKIX public keys
- [x] Generate tokens with RS256
- [x] Validate tokens with public key
- [x] Middleware extracts and validates tokens
- [x] Store userID in context

### Configuration:
- [x] Config loads key paths
- [x] Environment variables supported
- [x] Default paths set
- [x] Docker mounts keys directory

### Routes:
- [x] Public routes: `/auth/login`, `/auth/register`
- [x] Protected routes: `/tasks/*`, `/auth/me`
- [x] Middleware applied to protected routes

---

## 🎓 Part 12: Key Concepts

### RS256 Algorithm:
- **RSA** + **SHA-256**
- Asymmetric encryption
- Private key signs, public key verifies

### Token Expiration:
- **24 hours** from creation
- Stored in `exp` claim
- Checked automatically during validation

### Context Storage:
- `c.Locals("userID")` - Available in handlers
- `c.Locals("username")` - Available in handlers
- Set by middleware, used by handlers

### Stateless Authentication:
- No database lookup needed
- Token contains all user info
- Fast and scalable

---

## 🔍 Part 13: Debugging

### Check Token Contents:
1. Visit: https://jwt.io
2. Paste your token
3. See header, payload, signature

### Verify Keys:
```bash
# Check private key
openssl rsa -in keys/private.pem -text -noout

# Check public key
openssl rsa -in keys/public.pem -pubin -text -noout
```

### Test Token Generation:
```go
// In your code, add logging:
token, err := jwtService.GenerateToken(1, "testuser")
fmt.Printf("Generated token: %s\n", token)
```

### Test Token Validation:
```go
claims, err := jwtService.ValidateToken(tokenString)
fmt.Printf("UserID: %d, Username: %s\n", claims.UserID, claims.Username)
```

---

## 📚 Part 14: References

### Libraries Used:
- `github.com/golang-jwt/jwt/v5` - JWT library
- `crypto/rsa` - RSA key handling
- `crypto/x509` - Key parsing
- `encoding/pem` - PEM format decoding

### Standards:
- **JWT**: RFC 7519
- **RS256**: RFC 7518
- **PKCS#8**: RFC 5208
- **PKIX**: RFC 5280

---

## ✅ Part 15: Quick Reference

### Generate Keys:
```bash
./scripts/generate-keys.sh
```

### Start Services:
```bash
docker compose up -d
```

### Test Registration:
```bash
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","email":"user@example.com","password":"pass123"}'
```

### Test Protected Endpoint:
```bash
curl -X GET http://localhost:3000/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

**That's everything!** 🎉

Your JWT implementation is complete and working. The keys are generated once, tokens are created on login/register, and middleware protects all routes automatically.

