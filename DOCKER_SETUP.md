# Docker Setup Guide

This guide will help you run the entire project (API + CockroachDB + Redis) using Docker.

## 🚀 Quick Start

### Step 1: Make init script executable
```bash
chmod +x init-db.sh
```

### Step 2: Start all services
```bash
docker-compose up -d
```

### Step 3: Initialize database
```bash
./init-db.sh
```

### Step 4: Check logs
```bash
docker-compose logs -f api
```

### Step 5: Test API
```bash
curl http://localhost:3000/tasks
```

**That's it!** Your API is running at `http://localhost:3000`

---

## 📋 What Gets Started

1. **CockroachDB** - Database (port 26257)
   - Admin UI: http://localhost:8080
   - Connection: `postgresql://root@cockroachdb:26257/taskmanager?sslmode=disable`

2. **Redis** - Cache (port 6379)
   - Connection: `redis:6379`

3. **API** - Your Go application (port 3000)
   - API: http://localhost:3000
   - Health: http://localhost:3000/tasks

---

## 🛠️ Common Commands

### Start everything
```bash
docker-compose up -d
```

### Stop everything
```bash
docker-compose down
```

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f cockroachdb
docker-compose logs -f redis
```

### Restart a service
```bash
docker-compose restart api
```

### Rebuild after code changes
```bash
docker-compose up -d --build api
```

### Access CockroachDB SQL shell
```bash
docker exec -it taskmanager-cockroachdb ./cockroach sql --insecure
```

### Access Redis CLI
```bash
docker exec -it taskmanager-redis redis-cli
```

### Remove everything (including data)
```bash
docker-compose down -v
```

---

## 🔧 Development Workflow

### 1. Make code changes
Edit your Go files as usual.

### 2. Rebuild and restart
```bash
docker-compose up -d --build api
```

### 3. Check logs
```bash
docker-compose logs -f api
```

---

## 🐛 Troubleshooting

### Issue: Port already in use
**Solution:** Change ports in `docker-compose.yml`:
```yaml
ports:
  - "3001:8080"  # Change 3000 to 3001
```

### Issue: Database connection failed
**Solution:** Wait for CockroachDB to be ready:
```bash
# Check if CockroachDB is healthy
docker-compose ps

# Wait a bit, then restart API
docker-compose restart api
```

### Issue: Can't connect to Redis
**Solution:** Check Redis is running:
```bash
docker-compose ps redis
docker-compose logs redis
```

### Issue: Database doesn't exist
**Solution:** Run init script:
```bash
./init-db.sh
```

### Issue: Changes not reflected
**Solution:** Rebuild the image:
```bash
docker-compose up -d --build api
```

---

## 📁 File Structure

```
.
├── Dockerfile              # Go app container
├── docker-compose.yml      # Orchestrates all services
├── .dockerignore          # Files to exclude from build
├── init-db.sh             # Database initialization script
└── .env.example           # Environment variables template
```

---

## 🌐 Accessing Services

### API Endpoints
- **Base URL:** http://localhost:3000
- **Get all tasks:** GET http://localhost:3000/tasks
- **Create task:** POST http://localhost:3000/tasks
- **Get task:** GET http://localhost:3000/tasks/:id
- **Update task:** PUT http://localhost:3000/tasks/:id
- **Delete task:** DELETE http://localhost:3000/tasks/:id
- **Complete task:** PATCH http://localhost:3000/tasks/:id/complete

### CockroachDB Admin UI
- **URL:** http://localhost:8081
- View database, run queries, monitor performance

### Redis
- **Port:** 6379 (internal)
- Access via: `docker exec -it taskmanager-redis redis-cli`

---

## 💾 Data Persistence

Data is stored in Docker volumes:
- `cockroachdb-data` - Database data
- `redis-data` - Cache data

**To remove all data:**
```bash
docker-compose down -v
```

**To backup data:**
```bash
# Backup CockroachDB
docker exec taskmanager-cockroachdb ./cockroach dump taskmanager --insecure > backup.sql

# Backup Redis
docker exec taskmanager-redis redis-cli SAVE
```

---

## 🎯 Next Steps

1. **Start services:** `docker-compose up -d`
2. **Initialize DB:** `./init-db.sh`
3. **Test API:** `curl http://localhost:3000/tasks`
4. **Start developing!**

---

## 📝 Notes

- **Docker uses:** `.env.docker` file (or env vars in docker-compose.yml)
- **Local development uses:** `.env` file (for `go run cmd/api/main.go`)
- **Service names in Docker:** `cockroachdb` and `redis` (internal Docker networking)
- **Service names locally:** `localhost` (for local services)
- All services restart automatically if they crash
- Data persists between container restarts

## 🔄 Environment Variables Explained

### When Running with Docker:
- Uses `.env.docker` file
- Service names: `cockroachdb`, `redis` (Docker internal)
- Connection: `postgresql://root@cockroachdb:26257/...`

### When Running Locally (without Docker):
- Uses `.env` file
- Service names: `localhost`
- Connection: `postgresql://root@localhost:26257/...`

### When Using Cloud:
- Update `.env` with cloud connection strings
- Use cloud URLs: `postgresql://user:pass@cloud-host:26257/...`

