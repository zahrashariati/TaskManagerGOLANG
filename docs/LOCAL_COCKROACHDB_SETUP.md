# Local CockroachDB Setup Guide

## Method 1: Direct Download (Recommended)

### Step 1: Download CockroachDB
```bash
# Download the binary directly
wget https://binaries.cockroachdb.com/cockroach-v23.2.11.linux-amd64.tgz

# Extract it
tar -xzf cockroach-v23.2.11.linux-amd64.tgz

# Move to system path
sudo cp cockroach-v23.2.11.linux-amd64/cockroach /usr/local/bin/

# Make it executable
sudo chmod +x /usr/local/bin/cockroach

# Verify installation
cockroach version
```

### Step 2: Start CockroachDB
```bash
# Start a single-node cluster (insecure mode for local development)
cockroach start-single-node --insecure --listen-addr=localhost
```

**Keep this terminal open!** CockroachDB will run in the foreground.

### Step 3: Create Database (in a NEW terminal)
```bash
# Connect to CockroachDB SQL shell
cockroach sql --insecure --host=localhost:26257
```

Then run:
```sql
CREATE DATABASE taskmanager;
\q
```

### Step 4: Update Your .env File
```bash
# Add to your .env file:
DATABASE_URL=postgresql://root@localhost:26257/taskmanager?sslmode=disable
```

---

## Method 2: Using Package Manager (Easier)

### For Ubuntu/Debian:
```bash
# Add CockroachDB repository
curl https://binaries.cockroachdb.com/cockroach-v23.2.11.linux-amd64.tgz | tar -xz
sudo cp cockroach-v23.2.11.linux-amd64/cockroach /usr/local/bin/
```

### For Arch Linux:
```bash
yay -S cockroachdb
```

---

## Method 3: Using Docker (Easiest!)

### Step 1: Install Docker (if not installed)
```bash
# Check if Docker is installed
docker --version

# If not installed, install Docker:
# Ubuntu/Debian:
sudo apt-get update
sudo apt-get install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

### Step 2: Run CockroachDB in Docker
```bash
# Start CockroachDB container
docker run -d \
  --name=cockroachdb \
  -p 26257:26257 \
  -p 8080:8080 \
  cockroachdb/cockroach:latest-v23.2 \
  start-single-node --insecure
```

### Step 3: Create Database
```bash
# Create database
docker exec -it cockroachdb ./cockroach sql --insecure --execute="CREATE DATABASE taskmanager;"
```

### Step 4: Connection String
```
DATABASE_URL=postgresql://root@localhost:26257/taskmanager?sslmode=disable
```

### Useful Docker Commands:
```bash
# Stop CockroachDB
docker stop cockroachdb

# Start CockroachDB
docker start cockroachdb

# View logs
docker logs cockroachdb

# Remove container
docker rm cockroachdb
```

---

## Troubleshooting

### Issue: "wget: unexpected end of file"
**Solution:** Download directly in browser or use curl:
```bash
curl -L https://binaries.cockroachdb.com/cockroach-v23.2.11.linux-amd64.tgz -o cockroach.tgz
tar -xzf cockroach.tgz
sudo cp cockroach-v23.2.11.linux-amd64/cockroach /usr/local/bin/
```

### Issue: "Permission denied"
**Solution:** Use sudo or add to PATH:
```bash
# Option 1: Use sudo
sudo cp cockroach /usr/local/bin/

# Option 2: Add to local bin
mkdir -p ~/bin
cp cockroach ~/bin/
export PATH=$PATH:~/bin
```

### Issue: "Port 26257 already in use"
**Solution:** Stop existing CockroachDB or use different port:
```bash
# Find what's using the port
sudo lsof -i :26257

# Kill the process
kill <PID>

# Or use different port
cockroach start-single-node --insecure --port=26258
```

### Issue: "Connection refused"
**Solution:** Make sure CockroachDB is running:
```bash
# Check if running
ps aux | grep cockroach

# Start if not running
cockroach start-single-node --insecure
```

---

## Verify Installation

### Test Connection:
```bash
# Connect to SQL shell
cockroach sql --insecure --host=localhost:26257

# Run test query
SHOW DATABASES;
SELECT version();
\q
```

### Test from Go:
```bash
# Run your app
go run cmd/api/main.go

# Should see:
# "Server starting on :8080"
# (No database connection errors)
```

---

## Quick Start Commands

```bash
# 1. Start CockroachDB
cockroach start-single-node --insecure

# 2. In another terminal, create database
cockroach sql --insecure --execute="CREATE DATABASE taskmanager;"

# 3. Update .env
echo "DATABASE_URL=postgresql://root@localhost:26257/taskmanager?sslmode=disable" >> .env

# 4. Run your app
go run cmd/api/main.go
```

---

## Recommended: Docker Method

**Why Docker?**
- ✅ Easy to start/stop
- ✅ No installation conflicts
- ✅ Clean environment
- ✅ Easy to remove

**Quick Docker Setup:**
```bash
# One command to start everything
docker run -d --name=cockroachdb -p 26257:26257 cockroachdb/cockroach:latest-v23.2 start-single-node --insecure

# Create database
docker exec -it cockroachdb ./cockroach sql --insecure --execute="CREATE DATABASE taskmanager;"

# Done! Use: postgresql://root@localhost:26257/taskmanager?sslmode=disable
```


