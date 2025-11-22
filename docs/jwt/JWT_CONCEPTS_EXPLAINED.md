# JWT Concepts Explained 🔍

Detailed explanation of UUID, PEM format, and parsing.

---

## 1️⃣ What is UUID in Claims?

### UUID = Universally Unique Identifier

**In your JWT token:**
```go
ID: uuid.New().String()  // Generates unique ID for each token
```

**Example UUID:** `591d91fa-eeb7-40ee-9ccc-19ae8656a1a2`

**What it's used for:**
- **JTI (JWT ID)**: Unique identifier for THIS specific token
- **Token tracking**: Can identify/revoke specific tokens
- **Security**: Prevents token reuse attacks

**In your token:**
```json
{
  "user_id": 1,
  "username": "john",
  "jti": "591d91fa-eeb7-40ee-9ccc-19ae8656a1a2",  ← UUID
  "exp": 1763471855,
  ...
}
```

**Why UUID?**
- Guaranteed unique (even across different servers)
- Format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- Generated using `github.com/google/uuid`

**Code:**
```go
import "github.com/google/uuid"

ID: uuid.New().String()  // Creates new UUID like "591d91fa-eeb7-..."
```

---

## 2️⃣ What is PEM Format?

### PEM = Privacy-Enhanced Mail (file format)

**What it is:**
- Text format for storing cryptographic keys/certificates
- Base64 encoded binary data
- Human-readable (can open in text editor)

**Structure:**
```
-----BEGIN TYPE-----
(base64 encoded data)
-----END TYPE-----
```

**Your private key (PEM format):**
```
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8Eu6d4St1bbTv
/NGkF7sKS3qDcUEbZnJ8X7J6JNbPED8+MtYwrYAn4lNxKpww4TdItiLsVQdsxNct
XuK5Tn1pM5qp/kXVkUZxYRhnMZgJy2MuV00JGhrfFFBKNdTklkH5v+1rxOsnmzXV
...
-----END PRIVATE KEY-----
```

**Your public key (PEM format):**
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvBLuneErdW207/zRpBe7
Ckt6g3FBG2ZyfF+yeiTWzxA/PjLWMK2AJ+JTcSqcMOE3SLYi7FUHbMTXLV7iuU59
...
-----END PUBLIC KEY-----
```

**Why PEM?**
- Standard format (works everywhere)
- Easy to share (text file)
- Can be stored in files, environment variables, etc.

**Other formats:**
- **DER**: Binary format (not human-readable)
- **PKCS#12**: Container format (.p12 files)
- **PEM**: Text format (what we use) ✅

---

## 3️⃣ How Do We Parse PEM?

### Step-by-Step Parsing Process:

#### Step 1: Read File
```go
privateKeyBytes, err := os.ReadFile("keys/private.pem")
// Result: []byte containing the PEM file content
```

#### Step 2: Decode PEM Block
```go
block, _ := pem.Decode(privateKeyBytes)
// Result: *pem.Block with:
//   - Type: "PRIVATE KEY" or "PUBLIC KEY"
//   - Bytes: The actual key data (binary, base64 decoded)
```

**What `pem.Decode()` does:**
- Finds `-----BEGIN...-----` and `-----END...-----`
- Extracts the base64 data between them
- Decodes base64 → binary
- Returns `*pem.Block`

#### Step 3: Parse the Key Data
```go
// For private key (PKCS#8 format):
key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
privateKey := key.(*rsa.PrivateKey)

// For public key (PKIX format):
key, err := x509.ParsePKIXPublicKey(block.Bytes)
publicKey := key.(*rsa.PublicKey)
```

**What `x509.Parse...()` does:**
- Takes binary key data
- Parses according to format (PKCS#8, PKIX, etc.)
- Returns Go crypto types (`*rsa.PrivateKey`, `*rsa.PublicKey`)

---

## 📊 Complete Parsing Flow Diagram

```
┌─────────────────┐
│ keys/private.pem│ (Text file)
└────────┬────────┘
         │
         │ os.ReadFile()
         │
         ▼
┌─────────────────┐
│  []byte (PEM)   │ "-----BEGIN PRIVATE KEY-----\nMIIEvQ..."
└────────┬────────┘
         │
         │ pem.Decode()
         │ - Finds BEGIN/END markers
         │ - Extracts base64 data
         │ - Decodes base64 → binary
         │
         ▼
┌─────────────────┐
│  *pem.Block     │
│  Type: "PRIVATE │
│         KEY"    │
│  Bytes: [binary]│
└────────┬────────┘
         │
         │ x509.ParsePKCS8PrivateKey()
         │ - Parses binary data
         │ - Validates format
         │
         ▼
┌─────────────────┐
│ *rsa.PrivateKey │ (Go crypto type - ready to use!)
└─────────────────┘
```

---

## 💻 Code Example: Complete Parsing

### Private Key Parsing:

```go
// 1. Read file
privateKeyBytes, err := os.ReadFile("keys/private.pem")
// privateKeyBytes = []byte("-----BEGIN PRIVATE KEY-----\nMIIEvQ...")

// 2. Decode PEM
block, _ := pem.Decode(privateKeyBytes)
// block.Type = "PRIVATE KEY"
// block.Bytes = []byte{0x30, 0x82, 0x01, ...} (binary key data)

// 3. Parse key
privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
// privateKey = *rsa.PrivateKey (ready to use for signing!)
```

### Public Key Parsing:

```go
// 1. Read file
publicKeyBytes, err := os.ReadFile("keys/public.pem")

// 2. Decode PEM
block, _ := pem.Decode(publicKeyBytes)
// block.Type = "PUBLIC KEY"
// block.Bytes = []byte{0x30, 0x59, 0x30, ...} (binary key data)

// 3. Parse key
key, err := x509.ParsePKIXPublicKey(block.Bytes)
publicKey := key.(*rsa.PublicKey)
// publicKey = *rsa.PublicKey (ready to use for verification!)
```

---

## 🔍 Understanding PEM Structure

### What's Inside a PEM File:

```
-----BEGIN PRIVATE KEY-----          ← Header (marks start)
MIIEvQIBADANBgkqhkiG9w0BAQEFA...    ← Base64 encoded binary data
Ckt6g3FBG2ZyfF+yeiTWzxA/PjLWMK...    ← (multiple lines)
XuK5Tn1pM5qp/kXVkUZxYRhnMZgJy...    ←
-----END PRIVATE KEY-----            ← Footer (marks end)
```

**Breaking it down:**
1. **Header**: `-----BEGIN TYPE-----` (tells you what's inside)
2. **Body**: Base64 encoded binary data (the actual key)
3. **Footer**: `-----END TYPE-----` (marks end)

**Base64 encoding:**
- Binary data → Text representation
- Uses characters: A-Z, a-z, 0-9, +, /
- Safe to store in text files

---

## 🛠️ Parsing Functions Explained

### `pem.Decode()`:
```go
block, rest := pem.Decode(data)
```

**Parameters:**
- `data`: `[]byte` containing PEM data

**Returns:**
- `block`: `*pem.Block` with decoded data
- `rest`: Remaining data (if multiple PEM blocks)

**What it does:**
- Finds `-----BEGIN...-----` marker
- Finds `-----END...-----` marker
- Extracts base64 data between markers
- Decodes base64 → binary
- Returns `*pem.Block` with `Type` and `Bytes`

### `x509.ParsePKCS8PrivateKey()`:
```go
key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
```

**Parameters:**
- `block.Bytes`: Binary key data (from PEM decode)

**Returns:**
- `key`: `interface{}` (can be RSA, ECDSA, etc.)
- `err`: Error if parsing fails

**What it does:**
- Parses PKCS#8 format (generic private key format)
- Validates structure
- Returns Go crypto key type

### `x509.ParsePKIXPublicKey()`:
```go
key, err := x509.ParsePKIXPublicKey(block.Bytes)
```

**Parameters:**
- `block.Bytes`: Binary key data

**Returns:**
- `key`: `interface{}` (public key)
- `err`: Error if parsing fails

**What it does:**
- Parses PKIX/X.509 format (standard public key format)
- Validates structure
- Returns Go crypto key type

---

## 📝 Why We Support Multiple Formats

### Private Key Formats:

**PKCS#1** (Old format):
```
-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----
```
- RSA-specific
- Older OpenSSL versions

**PKCS#8** (Modern format):
```
-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
```
- Generic format (works with RSA, ECDSA, etc.)
- Modern OpenSSL default ✅

**Our code tries both:**
```go
// Try PKCS#1 first
privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
if err != nil {
    // Fall back to PKCS#8
    key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
    privateKey = key.(*rsa.PrivateKey)
}
```

### Public Key Formats:

**PKCS#1** (RSA-specific):
```
-----BEGIN RSA PUBLIC KEY-----
...
-----END RSA PUBLIC KEY-----
```

**PKIX/X.509** (Standard):
```
-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----
```
- Standard format ✅
- What `openssl rsa -pubout` generates

**Our code uses PKIX:**
```go
key, err := x509.ParsePKIXPublicKey(block.Bytes)
publicKey := key.(*rsa.PublicKey)
```

---

## 🎯 Summary

### UUID:
- **What**: Unique identifier for each token
- **Format**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- **Used for**: JTI claim (JWT ID)
- **Why**: Token tracking, security

### PEM Format:
- **What**: Text format for storing keys
- **Structure**: `-----BEGIN TYPE-----\nbase64_data\n-----END TYPE-----`
- **Why**: Human-readable, standard format
- **Contains**: Base64 encoded binary key data

### Parsing:
1. **Read file** → `[]byte`
2. **Decode PEM** → `*pem.Block` (extracts binary data)
3. **Parse key** → `*rsa.PrivateKey` or `*rsa.PublicKey` (Go crypto type)

**Result**: Keys ready to use for signing/verifying tokens! 🔐

