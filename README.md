# Go Learning Project - Task Manager 🎓

Welcome! This is your **self-guided learning project** to master Go programming.

## 🎯 What You'll Build

A complete REST API task manager backend with:
- ✅ **HTTP Server** - REST API with routes and handlers
- ✅ **CockroachDB** - Production-grade distributed database (PostgreSQL compatible)
- ✅ **Redis Caching** - High-performance caching layer
- ✅ **JSON API** - Request/response handling
- ✅ **Concurrency** - Goroutines and channels
- ✅ **Testing** - Unit tests

## 📚 How to Use This Project

**Important:** This is a **learning project** - you'll write the code yourself following the guide!

1. **Read `LEARNING_GUIDE.md`** - Step-by-step instructions
2. **Reference `QUICK_REFERENCE.md`** - Syntax cheat sheet
3. **Write code yourself** - Don't copy-paste, type it out!
4. **Ask questions** - I'm here to help when you're stuck

## 🚀 Quick Start

1. **Set up CockroachDB and Redis:**
   - Read `SETUP.md` for detailed instructions
   - Sign up for free tiers or install locally
   - Get your connection strings

2. **Set environment variables:**
   ```bash
   export DATABASE_URL="postgresql://user:password@host:26257/database?sslmode=require"
   export REDIS_URL="redis://default:password@host:port"
   ```

3. **Start coding:**
   ```bash
   # Create your first file
   touch main.go
   ```

4. **Open `LEARNING_GUIDE.md`** and follow Phase 1, Step 1.1

5. **Write your first HTTP server:**
   ```go
   package main
   
   import (
       "net/http"
       "log"
   )
   
   func main() {
       http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
           w.Write([]byte("Hello, API!"))
       })
       
       log.Println("Server starting on :8080")
       log.Fatal(http.ListenAndServe(":8080", nil))
   }
   ```

6. **Run it:**
   ```bash
   go run main.go
   # Test: curl http://localhost:8080
   ```

## 📖 Learning Path

Follow the phases in `LEARNING_GUIDE.md`:

- **Day 1**: Basics & Setup
- **Day 2-3**: Database Integration (CockroachDB)
- **Day 4**: Caching Layer (Redis)
- **Day 5**: REST API (HTTP Server, Routes, Handlers)
- **Day 6**: Testing
- **Day 7**: Advanced Features (Goroutines, Concurrency)

## 🛠️ Prerequisites

- Go 1.21+ installed
- Text editor (VS Code, Vim, etc.)
- Terminal/Command prompt
- Willingness to learn! 😊

## 📝 Project Structure

```
booking_ticket/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # HTTP handlers (API layer)
│   │   └── task_handler.go
│   ├── service/                 # Business logic layer
│   │   └── task_service.go
│   ├── repository/              # Data access layer
│   │   └── task_repository.go
│   ├── models/                  # Data models
│   │   └── task.go
│   ├── config/                  # Configuration
│   │   └── config.go
│   └── cache/                   # Cache layer
│       └── cache.go
├── LEARNING_GUIDE.md            # Step-by-step instructions
├── ARCHITECTURE.md              # Architecture explanation
├── ERROR_HANDLING.md            # Error handling guide (IMPORTANT!)
├── QUICK_REFERENCE.md           # Syntax cheat sheet
├── SETUP.md                     # CockroachDB & Redis setup
├── README.md                    # This file
└── go.mod                       # Go module file
```

**Read these guides:**
- `ARCHITECTURE.md` - Understand layered structure
- `ERROR_HANDLING.md` - **CRITICAL** - Learn error handling patterns

## 💡 Learning Philosophy

1. **Write code yourself** - Muscle memory matters
2. **Read errors carefully** - They're trying to help!
3. **Experiment** - Break things, fix them, learn
4. **Ask questions** - No question is too small
5. **Take breaks** - Learning takes time

## 🆘 Need Help?

When you're stuck:
1. Check `QUICK_REFERENCE.md` for syntax
2. Read the error message carefully
3. Search Go documentation: `go doc package.Function`
4. Ask me specific questions!

## 🎉 You've Got This!

Remember: The goal is to **learn**, not to finish quickly. Take your time, understand each concept, and build something you're proud of!

**Happy coding! 🚀**
