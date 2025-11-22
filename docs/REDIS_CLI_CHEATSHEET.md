# Redis CLI Cheatsheet 🔴

Complete guide to using `redis-cli` commands for your Task Manager API.

---

## 🔌 Connecting to Redis

### From Host Machine (Your PC):
```bash
# Connect to Redis in Docker container
docker compose exec redis redis-cli

# Or using port mapping
redis-cli -h localhost -p 6380
```

### From Inside Container:
```bash
# Already connected if you exec into container
docker compose exec redis sh
redis-cli
```

---

## 📋 Basic Commands

### Connection & Info
```bash
# Check if Redis is running
PING
# Response: PONG

# Get server info
INFO

# Get specific info section
INFO memory
INFO stats
INFO keyspace

# Get all configuration
CONFIG GET *

# Get specific config
CONFIG GET maxmemory
```

### Keys Management
```bash
# List all keys
KEYS *

# List keys matching pattern
KEYS tasks:*
KEYS task:*

# Count total keys
DBSIZE

# Check if key exists
EXISTS tasks:all
# Returns: 1 (exists) or 0 (doesn't exist)

# Get key type
TYPE tasks:all
# Returns: string, list, set, zset, hash

# Get time to live (TTL)
TTL tasks:all
# Returns: seconds until expiration, -1 (no expiration), -2 (key doesn't exist)

# Delete a key
DEL tasks:all

# Delete multiple keys
DEL tasks:all tasks:incomplete task:1 task:2

# Delete all keys (DANGER!)
FLUSHDB    # Current database only
FLUSHALL   # All databases (DANGER!)

# Rename a key
RENAME tasks:all tasks:all_backup

# Copy a key
COPY tasks:all tasks:all_copy
```

---

## 📦 String Operations (Your Tasks Cache)

### Get/Set Values
```bash
# Get value
GET tasks:all

# Set value
SET tasks:all "your json data here"

# Set with expiration (TTL in seconds)
SET tasks:all "data" EX 300
SETEX tasks:all 300 "data"

# Set only if key doesn't exist
SETNX tasks:all "data"

# Get and set atomically
GETSET tasks:all "new data"

# Append to value
APPEND tasks:all "more data"

# Get substring
GETRANGE tasks:all 0 10

# Get string length
STRLEN tasks:all

# Increment number
INCR counter
INCRBY counter 5

# Decrement number
DECR counter
DECRBY counter 5
```

### Your Application Keys
```bash
# View cached tasks (all)
GET tasks:all

# View cached tasks (incomplete only)
GET tasks:incomplete

# View specific task
GET task:1
GET task:123

# Check what's cached
KEYS tasks:*
KEYS task:*

# See all cached task IDs
KEYS task:*

# Check TTL (time until expiration)
TTL tasks:all
TTL task:1
```

---

## 🔍 Searching & Filtering

### Pattern Matching
```bash
# Find all task keys
KEYS task:*

# Find all tasks cache keys
KEYS tasks:*

# Find keys starting with "task"
KEYS task*

# Find keys with specific pattern
KEYS *:all
KEYS *:incomplete

# Count keys matching pattern
KEYS task:* | wc -l
```

### Scan (Safer than KEYS for production)
```bash
# Scan keys (cursor-based, non-blocking)
SCAN 0 MATCH task:* COUNT 100

# Scan all keys
SCAN 0

# Scan with pattern
SCAN 0 MATCH tasks:* COUNT 10
```

---

## ⏰ Expiration & TTL

### Set Expiration
```bash
# Set expiration in seconds
EXPIRE tasks:all 300

# Set expiration in milliseconds
PEXPIRE tasks:all 300000

# Set expiration at specific timestamp
EXPIREAT tasks:all 1704067200

# Remove expiration (make permanent)
PERSIST tasks:all

# Get remaining TTL
TTL tasks:all
# Returns: seconds, -1 (no expiration), -2 (key doesn't exist)

# Get remaining TTL in milliseconds
PTTL tasks:all
```

### Your App's Cache Expiration
```bash
# Check when cache expires
TTL tasks:all
# If returns 300, cache expires in 5 minutes

# Manually expire cache (force refresh)
EXPIRE tasks:all 0
# or
DEL tasks:all
```

---

## 📊 Monitoring & Debugging

### Monitor Commands in Real-Time
```bash
# Watch all commands (real-time)
MONITOR

# Exit monitor: Ctrl+C
```

### Statistics
```bash
# Get all stats
INFO stats

# Get memory stats
INFO memory

# Get keyspace info
INFO keyspace

# Get client connections
CLIENT LIST

# Get slow queries
SLOWLOG GET 10
```

### Debug Your Cache
```bash
# Check if cache exists
EXISTS tasks:all
EXISTS tasks:incomplete
EXISTS task:1

# See what's in cache
GET tasks:all | python3 -m json.tool  # Pretty print JSON
GET task:1 | python3 -m json.tool

# Check cache age
TTL tasks:all

# See all cached keys
KEYS *

# Count cached items
DBSIZE
```

---

## 🧹 Cache Management (Your App)

### Clear Task Cache
```bash
# Clear all tasks cache
DEL tasks:all tasks:incomplete

# Clear specific task cache
DEL task:1 task:2 task:3

# Clear all task-related cache
KEYS task* | xargs redis-cli DEL
# Or safer:
redis-cli --scan --pattern "task*" | xargs redis-cli DEL
```

### Invalidate Cache (Like Your App Does)
```bash
# Your app does this when tasks change:
DEL tasks:all tasks:incomplete

# Or clear everything
FLUSHDB
```

### Check Cache Status
```bash
# See all cached keys
KEYS *

# Check if cache is working
EXISTS tasks:all
# 1 = cached, 0 = not cached (cache miss)

# See cache size
MEMORY USAGE tasks:all
```

---

## 🔐 Advanced Operations

### Batch Operations
```bash
# Execute multiple commands
MULTI
SET key1 "value1"
SET key2 "value2"
DEL key3
EXEC

# Cancel transaction
MULTI
SET key1 "value1"
DISCARD
```

### Pub/Sub (Not used in your app, but useful)
```bash
# Subscribe to channel
SUBSCRIBE task_updates

# Publish message
PUBLISH task_updates "Task 1 completed"
```

---

## 🎯 Common Tasks for Your App

### 1. Check Cache Status
```bash
# Connect
docker compose exec redis redis-cli

# Check what's cached
KEYS *

# Check specific cache
GET tasks:all
GET tasks:incomplete
GET task:1
```

### 2. Clear Cache (Force Refresh)
```bash
# Clear all task cache
DEL tasks:all tasks:incomplete

# Clear specific task
DEL task:123

# Clear everything
FLUSHDB
```

### 3. Debug Cache Issues
```bash
# Check if key exists
EXISTS tasks:all

# See cache content
GET tasks:all

# Check expiration
TTL tasks:all

# See all keys
KEYS *
```

### 4. Monitor Cache Activity
```bash
# Watch commands in real-time
MONITOR

# Check stats
INFO stats
INFO memory
```

### 5. Test Cache Manually
```bash
# Set test data
SET tasks:all '[{"id":1,"title":"Test"}]'

# Get it back
GET tasks:all

# Set expiration
EXPIRE tasks:all 60

# Check TTL
TTL tasks:all
```

---

## 📝 Your App's Cache Keys

Based on your code, these are the keys used:

```
tasks:all          → Cached list of ALL tasks (completed + incomplete)
tasks:incomplete   → Cached list of ONLY incomplete tasks
task:1             → Cached single task with ID 1
task:2             → Cached single task with ID 2
...
```

### Cache Flow:
```
1. GET /tasks?showCompleted=true
   → Checks: tasks:all
   → If miss: Queries DB, stores in tasks:all

2. GET /tasks?showCompleted=false
   → Checks: tasks:incomplete
   → If miss: Queries DB, stores in tasks:incomplete

3. GET /tasks/:id
   → Checks: task:{id}
   → If miss: Queries DB, stores in task:{id}

4. POST/PUT/DELETE /tasks
   → Invalidates: DEL tasks:all tasks:incomplete
   → Next GET will refresh cache
```

---

## 🚀 Quick Reference

### Most Used Commands
```bash
# Connect
docker compose exec redis redis-cli

# List keys
KEYS *

# Get value
GET tasks:all

# Delete key
DEL tasks:all

# Check existence
EXISTS tasks:all

# Check TTL
TTL tasks:all

# Clear all
FLUSHDB

# Monitor
MONITOR

# Info
INFO
```

### One-Liners
```bash
# Count all keys
redis-cli DBSIZE

# Get all keys
redis-cli KEYS "*"

# Delete all task cache
redis-cli DEL tasks:all tasks:incomplete

# Check cache
redis-cli EXISTS tasks:all

# See cache content
redis-cli GET tasks:all | python3 -m json.tool
```

---

## 💡 Tips

1. **Use SCAN instead of KEYS** in production (non-blocking)
2. **Check TTL** to see when cache expires
3. **Use MONITOR** to debug cache operations
4. **FLUSHDB** clears current database (safe)
5. **FLUSHALL** clears ALL databases (dangerous!)
6. **JSON pretty print**: `GET key | python3 -m json.tool`

---

## 🎓 Practice Commands

Try these to understand your cache:

```bash
# 1. Connect
docker compose exec redis redis-cli

# 2. See what's cached
KEYS *

# 3. Get a cached value
GET tasks:all

# 4. Check expiration
TTL tasks:all

# 5. Clear cache
DEL tasks:all tasks:incomplete

# 6. Make API request (will populate cache)
# Then check again:
KEYS *
GET tasks:all
```

---

## 📚 Resources

- **Redis Docs**: https://redis.io/commands
- **Your Redis Port**: `6380` (host) → `6379` (container)
- **Container Name**: `taskmanager-redis`

Happy caching! 🚀

