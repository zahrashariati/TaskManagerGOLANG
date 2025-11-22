# Setup Guide - CockroachDB & Redis

This guide will help you set up CockroachDB and Redis for your task manager project.

## 🗄️ CockroachDB Setup

### Option 1: CockroachDB Cloud (Recommended - Free Tier)

1. **Sign Up**
   - Go to https://cockroachlabs.com/cloud/
   - Create a free account
   - No credit card required for free tier

2. **Create Cluster**
   - Click "Create Cluster"
   - Choose "Serverless" (free tier)
   - Select a cloud provider and region
   - Wait for cluster to be created (~2 minutes)

3. **Get Connection String**
   - Click on your cluster
   - Go to "Connection info" tab
   - Copy the connection string
   - It looks like: `postgresql://user:password@host:26257/defaultdb?sslmode=require`

4. **Set Environment Variable** (optional but recommended)
   ```bash
   export DATABASE_URL="postgresql://user:password@host:26257/defaultdb?sslmode=require"
   ```

### Option 2: Local CockroachDB

1. **Install CockroachDB**
   ```bash
   # macOS
   brew install cockroachdb/tap/cockroach
   
   # Linux
   wget -qO- https://binaries.cockroachdb.com/cockroach-v23.1.0.linux-amd64.tgz | tar xvz
   sudo cp cockroach-v23.1.0.linux-amd64/cockroach /usr/local/bin/
   
   # Windows
   # Download from https://www.cockroachlabs.com/docs/stable/install-cockroachdb-windows.html
   ```

2. **Start Local Cluster**
   ```bash
   cockroach start-single-node --insecure --listen-addr=localhost
   ```

3. **Create Database** (in another terminal)
   ```bash
   cockroach sql --insecure --host=localhost:26257
   ```
   Then in SQL:
   ```sql
   CREATE DATABASE taskmanager;
   ```

4. **Connection String**
   ```
   postgresql://root@localhost:26257/taskmanager?sslmode=disable
   ```

---

## 🔴 Redis Setup

### Option 1: Redis Cloud (Recommended - Free Tier)

1. **Sign Up**
   - Go to https://redis.com/try-free/
   - Create a free account
   - 30MB free tier available

2. **Create Database**
   - Click "New Subscription"
   - Choose "Fixed" plan (free)
   - Select cloud provider and region
   - Create database

3. **Get Connection URL**
   - Click on your database
   - Copy the "Public endpoint" URL
   - It looks like: `redis://default:password@host:port`

4. **Set Environment Variable** (optional but recommended)
   ```bash
   export REDIS_URL="redis://default:password@host:port"
   ```

### Option 2: Local Redis

1. **Install Redis**
   ```bash
   # macOS
   brew install redis
   
   # Linux (Ubuntu/Debian)
   sudo apt-get update
   sudo apt-get install redis-server
   
   # Linux (Fedora/RHEL)
   sudo dnf install redis
   
   # Windows
   # Use WSL or download from https://github.com/microsoftarchive/redis/releases
   ```

2. **Start Redis**
   ```bash
   redis-server
   ```

3. **Test Connection**
   ```bash
   redis-cli ping
   # Should return: PONG
   ```

4. **Connection URL**
   ```
   redis://localhost:6379
   ```

---

## 🚀 Quick Start

### 1. Set Environment Variables

```bash
# CockroachDB
export DATABASE_URL="postgresql://user:password@host:26257/defaultdb?sslmode=require"

# Redis
export REDIS_URL="redis://default:password@host:port"
```

### 2. Install Go Dependencies

```bash
cd booking_ticket
go get github.com/lib/pq
go get github.com/redis/go-redis/v9
```

### 3. Test Connections

Create a simple test file `test_connections.go`:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    
    _ "github.com/lib/pq"
    "github.com/redis/go-redis/v9"
)

func main() {
    // Test CockroachDB
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        fmt.Println("DATABASE_URL not set")
        return
    }
    
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        fmt.Printf("DB connection error: %v\n", err)
        return
    }
    defer db.Close()
    
    err = db.Ping()
    if err != nil {
        fmt.Printf("DB ping error: %v\n", err)
        return
    }
    fmt.Println("✅ CockroachDB connected!")
    
    // Test Redis
    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        fmt.Println("REDIS_URL not set")
        return
    }
    
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        fmt.Printf("Redis parse error: %v\n", err)
        return
    }
    
    client := redis.NewClient(opt)
    defer client.Close()
    
    ctx := context.Background()
    err = client.Ping(ctx).Err()
    if err != nil {
        fmt.Printf("Redis ping error: %v\n", err)
        return
    }
    fmt.Println("✅ Redis connected!")
}
```

Run it:
```bash
go run test_connections.go
```

---

## 🔧 Troubleshooting

### CockroachDB Issues

**Connection refused:**
- Check if CockroachDB is running
- Verify port 26257 is accessible
- Check firewall settings

**SSL error:**
- For local: use `sslmode=disable`
- For cloud: ensure `sslmode=require`

**Authentication failed:**
- Verify username/password
- Check connection string format

### Redis Issues

**Connection refused:**
- Check if Redis is running: `redis-cli ping`
- Verify port 6379 is accessible
- Check firewall settings

**Authentication error:**
- For Redis Cloud: use the password from connection URL
- For local: usually no password needed

**URL parse error:**
- Ensure URL format: `redis://[password@]host:port`
- For local: `redis://localhost:6379`

---

## 📝 Next Steps

Once both are set up:
1. Follow `LEARNING_GUIDE.md` Phase 2 for database integration
2. Follow `LEARNING_GUIDE.md` Phase 3 for Redis caching
3. Start building your task manager!

---

## 💡 Tips

- **Use environment variables** - Don't hardcode credentials
- **Test connections first** - Before writing application code
- **Use separate databases** - One for dev, one for tests
- **Keep credentials secure** - Never commit them to git
- **Use `.env` files** - For local development (with `godotenv` package)

---

**Ready to code! 🎉**


