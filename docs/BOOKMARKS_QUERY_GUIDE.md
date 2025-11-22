# Bookmarks Query Guide - VOD Platform

## What is a "Bookmarks Query"?

A **bookmarks query** is a database query that retrieves a user's bookmarked programs/videos from your VOD platform. It's similar to how this task manager queries tasks, but instead of tasks, you're querying bookmarked video content.

---

## Understanding the Concept

### In Simple Terms:
- **Bookmark**: A saved reference to a program/video that a user wants to watch later
- **Bookmarks Query**: A database query that fetches all (or filtered) bookmarks for a specific user
- **Bookmarks List**: The collection of bookmarked programs displayed to the user

### Real-World Example:
```
User clicks "My Bookmarks" → Frontend calls API → Backend runs bookmarks query → Returns list of bookmarked programs
```

---

## Database Structure (Typical VOD Platform)

### Tables You'd Likely Have:

```sql
-- Programs/Videos table
CREATE TABLE programs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    description TEXT,
    duration INTEGER,
    genre VARCHAR(100),
    release_date DATE,
    -- other fields...
);

-- Bookmarks table (junction table)
CREATE TABLE bookmarks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    program_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, program_id) -- Prevent duplicate bookmarks
);
```

---

## Common Bookmarks Query Patterns

### 1. **Get All Bookmarks for a User**
```sql
SELECT 
    b.id AS bookmark_id,
    b.created_at AS bookmarked_at,
    p.id AS program_id,
    p.title,
    p.description,
    p.duration,
    p.genre
FROM bookmarks b
INNER JOIN programs p ON b.program_id = p.id
WHERE b.user_id = $1
ORDER BY b.created_at DESC;
```

**What this does:**
- Gets all bookmarked programs for a specific user
- Joins with programs table to get program details
- Orders by most recently bookmarked first

**Similar to:** `GetAllTasks()` in your task manager

---

### 2. **Check if Program is Bookmarked**
```sql
SELECT EXISTS(
    SELECT 1 FROM bookmarks 
    WHERE user_id = $1 AND program_id = $2
);
```

**What this does:**
- Quickly checks if a specific program is already bookmarked
- Returns true/false
- Used before showing "Bookmark" vs "Unbookmark" button

---

### 3. **Get Bookmarks with Pagination**
```sql
SELECT 
    b.id AS bookmark_id,
    p.id AS program_id,
    p.title,
    p.description
FROM bookmarks b
INNER JOIN programs p ON b.program_id = p.id
WHERE b.user_id = $1
ORDER BY b.created_at DESC
LIMIT $2 OFFSET $3;
```

**What this does:**
- Limits results (e.g., 20 per page)
- Uses OFFSET for pagination
- Prevents loading thousands of bookmarks at once

**Parameters:**
- `$1`: user_id
- `$2`: limit (e.g., 20)
- `$3`: offset (e.g., 0 for page 1, 20 for page 2)

---

### 4. **Search Bookmarks by Program Title**
```sql
SELECT 
    b.id AS bookmark_id,
    p.id AS program_id,
    p.title,
    p.description
FROM bookmarks b
INNER JOIN programs p ON b.program_id = p.id
WHERE b.user_id = $1 
    AND p.title ILIKE '%' || $2 || '%'
ORDER BY b.created_at DESC;
```

**What this does:**
- Searches within bookmarked programs
- Case-insensitive search (ILIKE)
- Finds programs matching search term

---

### 5. **Filter Bookmarks by Genre**
```sql
SELECT 
    b.id AS bookmark_id,
    p.id AS program_id,
    p.title,
    p.genre
FROM bookmarks b
INNER JOIN programs p ON b.program_id = p.id
WHERE b.user_id = $1 
    AND p.genre = $2
ORDER BY b.created_at DESC;
```

**What this does:**
- Filters bookmarks by program genre
- Shows only bookmarked programs in specific category

---

## Go Implementation Pattern (Similar to Your Task Manager)

### Model Structure:
```go
type Bookmark struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    ProgramID int       `json:"program_id" db:"program_id"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type BookmarkedProgram struct {
    BookmarkID int       `json:"bookmark_id"`
    ProgramID  int       `json:"program_id"`
    Title      string    `json:"title"`
    Description string   `json:"description"`
    Duration   int       `json:"duration"`
    Genre      string    `json:"genre"`
    BookmarkedAt time.Time `json:"bookmarked_at"`
}
```

### Repository Pattern (Like Your TaskRepository):
```go
func (r *BookmarkRepository) GetAllBookmarks(userID int) ([]BookmarkedProgram, error) {
    query := `
        SELECT 
            b.id AS bookmark_id,
            b.created_at AS bookmarked_at,
            p.id AS program_id,
            p.title,
            p.description,
            p.duration,
            p.genre
        FROM bookmarks b
        INNER JOIN programs p ON b.program_id = p.id
        WHERE b.user_id = $1
        ORDER BY b.created_at DESC
    `
    
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var bookmarks []BookmarkedProgram
    for rows.Next() {
        var bookmark BookmarkedProgram
        err := rows.Scan(
            &bookmark.BookmarkID,
            &bookmark.BookmarkedAt,
            &bookmark.ProgramID,
            &bookmark.Title,
            &bookmark.Description,
            &bookmark.Duration,
            &bookmark.Genre,
        )
        if err != nil {
            return nil, err
        }
        bookmarks = append(bookmarks, bookmark)
    }
    
    return bookmarks, nil
}
```

---

## Common Issues That Need Revision

### 1. **Performance Issues**
**Problem:** Loading all bookmarks at once without pagination
```sql
-- BAD: Loads everything
SELECT * FROM bookmarks WHERE user_id = $1;
```

**Solution:** Add pagination
```sql
-- GOOD: Limits results
SELECT * FROM bookmarks WHERE user_id = $1 LIMIT 20 OFFSET 0;
```

---

### 2. **Missing JOINs**
**Problem:** Only getting bookmark IDs, not program details
```sql
-- BAD: Only bookmark data
SELECT * FROM bookmarks WHERE user_id = $1;
```

**Solution:** JOIN with programs table
```sql
-- GOOD: Includes program details
SELECT b.*, p.title, p.description 
FROM bookmarks b
JOIN programs p ON b.program_id = p.id
WHERE b.user_id = $1;
```

---

### 3. **No Filtering/Sorting**
**Problem:** No way to filter or sort bookmarks
```sql
-- BAD: No sorting
SELECT * FROM bookmarks WHERE user_id = $1;
```

**Solution:** Add ORDER BY and WHERE filters
```sql
-- GOOD: Sorted by date, filterable
SELECT * FROM bookmarks 
WHERE user_id = $1 
    AND genre = $2  -- optional filter
ORDER BY created_at DESC;
```

---

### 4. **N+1 Query Problem**
**Problem:** Making separate query for each program
```go
// BAD: One query per program
bookmarks := getBookmarks(userID)
for _, bookmark := range bookmarks {
    program := getProgram(bookmark.ProgramID) // N queries!
}
```

**Solution:** Use JOIN to get everything in one query
```go
// GOOD: One query with JOIN
bookmarkedPrograms := getBookmarkedPrograms(userID) // Includes program data
```

---

### 5. **Missing Indexes**
**Problem:** Slow queries on large datasets
```sql
-- Missing index on user_id
SELECT * FROM bookmarks WHERE user_id = $1; -- Slow!
```

**Solution:** Add database indexes
```sql
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_program_id ON bookmarks(program_id);
```

---

## API Endpoint Examples

### GET /api/bookmarks
```go
func (h *BookmarkHandler) GetAllBookmarks(c *fiber.Ctx) error {
    userID := c.Locals("userID").(int)
    
    // Get query parameters
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 20)
    genre := c.Query("genre") // optional filter
    
    bookmarks, err := h.service.GetAllBookmarks(userID, page, limit, genre)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(bookmarks)
}
```

**Query Examples:**
- `GET /api/bookmarks` - Get all bookmarks (first 20)
- `GET /api/bookmarks?page=2` - Get page 2
- `GET /api/bookmarks?genre=action` - Filter by genre
- `GET /api/bookmarks?page=2&limit=50&genre=comedy` - Combined filters

---

## What Your Senior Might Want Revised

Based on common issues, here's what might need fixing:

### 1. **Add Pagination**
- Currently might be loading all bookmarks at once
- Need to add `LIMIT` and `OFFSET` to queries
- Add page/limit query parameters to API

### 2. **Optimize JOINs**
- Ensure you're joining with programs table properly
- Avoid N+1 queries
- Use proper JOIN types (INNER vs LEFT)

### 3. **Add Filtering**
- Filter by genre, date, program type
- Search within bookmarks
- Sort options (date, title, etc.)

### 4. **Add Caching**
- Similar to your task manager's cache
- Cache frequently accessed bookmarks
- Invalidate cache when bookmark is added/removed

### 5. **Error Handling**
- Handle cases where program was deleted but bookmark exists
- Handle database connection errors
- Validate user permissions

### 6. **Performance**
- Add database indexes
- Optimize query execution plans
- Consider using materialized views for complex queries

---

## Key Takeaways

1. **Bookmarks Query** = Database query to fetch user's bookmarked programs
2. **Similar Pattern** = Like `GetAllTasks()` but for bookmarks
3. **Common Issues** = Performance, missing JOINs, no pagination, N+1 queries
4. **Revision Needed** = Likely performance optimization, pagination, or query structure

---

## Questions to Ask Your Senior

1. **What specific issue** needs revision? (Performance? Missing features?)
2. **What's the current implementation** doing wrong?
3. **What's the expected behavior** after revision?
4. **Are there specific requirements** I should know about? (Pagination? Filtering? Sorting?)

---

## Next Steps

1. **Find the current bookmarks query** in your VOD platform codebase
2. **Identify the issue** (ask your senior or check performance logs)
3. **Apply the appropriate fix** from this guide
4. **Test thoroughly** with different scenarios
5. **Review with your senior** before deploying

---

## Related Concepts in Your Task Manager

Your task manager already implements similar patterns:
- ✅ **Query with user filtering**: `GetAll(userID, showCompleted)`
- ✅ **Pagination ready**: Can be added easily
- ✅ **Caching**: Already implemented in `TaskService`
- ✅ **Error handling**: Proper error types and handling

You can use these patterns as reference when implementing bookmarks queries!


