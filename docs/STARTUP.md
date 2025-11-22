# Quick Startup Guide 🚀

## After Restarting Your PC

Run these commands to start your Task Manager API:

### Option 1: Start Everything (Recommended)

```bash
cd ~/Desktop/task_manager
docker compose up -d
```

This will:
- ✅ Start CockroachDB
- ✅ Start Redis  
- ✅ Create database automatically
- ✅ Start API server

### Option 2: Check Status First

```bash
cd ~/Desktop/task_manager
docker compose ps
```

If containers are stopped, start them:
```bash
docker compose up -d
```

### Option 3: Rebuild After Code Changes

```bash
cd ~/Desktop/task_manager
docker compose build api
docker compose up -d
```

---

## Verify Everything is Running

```bash
# Check container status
docker compose ps

# Check API logs
docker compose logs api --tail 20

# Test API
curl http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"test123"}'
```

---

## Common Commands

### Start services:
```bash
docker compose up -d
```

### Stop services:
```bash
docker compose down
```

### View logs:
```bash
docker compose logs -f api        # API logs
docker compose logs -f            # All logs
```

### Restart API only:
```bash
docker compose restart api
```

### Rebuild and restart:
```bash
docker compose build api && docker compose up -d
```

---

## Quick Checklist ✅

After PC restart:
- [ ] `cd ~/Desktop/task_manager`
- [ ] `docker compose up -d`
- [ ] Wait ~10 seconds for services to start
- [ ] Check logs: `docker compose logs api --tail 5`
- [ ] Test: `curl http://localhost:3000/auth/register ...`

---

**That's it!** Your API will be running on `http://localhost:3000` 🎉

