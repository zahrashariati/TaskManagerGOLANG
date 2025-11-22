# Kafka Questions & Answers - Detailed Explanations

## 1. What is UUID and Where is it Used?

### What is UUID?
**UUID** = **Universally Unique Identifier**

- A 128-bit identifier that is **guaranteed to be unique** across space and time
- Format: `550e8400-e29b-41d4-a716-446655440000` (32 hex digits, 5 groups)
- Examples:
  - `123e4567-e89b-12d3-a456-426614174000`
  - `f47ac10b-58cc-4372-a567-0e02b2c3d479`

### Why Use UUID?
1. **Uniqueness**: No collisions (extremely unlikely)
2. **No coordination needed**: Can generate without checking database
3. **Distributed systems**: Multiple services can generate IDs independently
4. **Security**: Harder to guess than sequential IDs (1, 2, 3...)

### Where is UUID Used in Your System?

**In Kafka Events:**
```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",  // ← UUID
  "event_type": "task_scheduled",
  "task_id": 123,
  "due_date": "2024-11-09T10:00:00Z"
}
```

**Why `event_id` needs to be UUID:**
- **Deduplication**: If same message is sent twice (network retry), you can detect duplicates
- **Tracing**: Track specific events across services
- **Idempotency**: Consumer can check "Did I process this event_id already?"

**In Go:**
```go
import "github.com/google/uuid"

// Generate UUID
eventID := uuid.New().String()
// Result: "550e8400-e29b-41d4-a716-446655440000"
```

**Alternative:** You could use database auto-increment IDs, but UUIDs are better for:
- Distributed systems
- Event tracking
- Avoiding ID conflicts

---

## 2. Past Notifications: Delete or Store?

### The Question
In consumer logic, it says "check if due_date matches today or is in the past." Should we delete past notifications or store them?

### Answer: **Store Them (with Retention Policy)**

### Why Store Past Notifications?

#### Reason 1: **Kafka Retention Policy**
- Kafka automatically deletes old messages based on **retention time** (default: 7 days)
- You don't manually delete - Kafka does it automatically
- This is **configurable** (can keep for 30 days, 1 year, etc.)

#### Reason 2: **Replay Capability**
- If your notifier service crashes and restarts, it can **replay** past events
- Useful for debugging: "Why didn't user get notification on 11/9?"
- Can reprocess events if there was a bug

#### Reason 3: **Audit Trail**
- Keep history of what notifications were sent
- Compliance/legal requirements
- Analytics: "How many tasks were overdue last month?"

#### Reason 4: **Late Arrivals**
- Sometimes events arrive late (network issues)
- If task was due on 11/9 but event arrives on 11/10, you still want to notify

### How It Works:

```
Timeline:
Day 1: Task created with due_date = 11/9 → Event sent to Kafka
Day 2-8: Event sits in Kafka (waiting)
Day 9: Notifier reads event → Checks date → Sends notification
Day 10-16: Event still in Kafka (for replay/audit)
Day 17: Kafka automatically deletes (7-day retention)
```

### Best Practice:
1. **Kafka stores events** (with retention: 7-30 days)
2. **Your notifier processes** events when due_date arrives
3. **Optional: Database table** to track "notifications_sent" (for longer-term storage)
4. **Kafka auto-deletes** old messages (you don't manually delete)

### Implementation Strategy:

**Option A: Process Immediately (Simple)**
- Consumer reads event → Check if due_date <= today → Send notification
- Problem: If event arrives early, sends notification early

**Option B: Store and Schedule (Better)**
- Consumer reads event → Store in memory/database
- Scheduler checks every hour: "Which tasks are due now?"
- Send notifications for tasks that match
- **This is what you want (Option 3)**

---

## 3. How to Implement Scheduler (Cron-like) - Option 3

### What You Want:
- Consumer reads events from Kafka
- Stores them somewhere
- Scheduler runs periodically (every hour/day) to check "which tasks are due?"
- Sends notifications for due tasks

### Architecture:

```
┌─────────────┐
│   Kafka     │
│  Consumer   │ → Reads events → Stores in memory/DB
└─────────────┘

┌─────────────┐
│  Scheduler  │ → Runs every hour → Checks stored tasks
│   (Cron)    │ → Finds due tasks → Sends notifications
└─────────────┘
```

### Implementation Steps:

#### Step 1: Store Events (In-Memory or Database)

**Option A: In-Memory (Simple, for learning)**
```go
type ScheduledTask struct {
    EventID  string
    TaskID   int
    UserID   int
    Title    string
    DueDate  time.Time
}

var scheduledTasks []ScheduledTask  // Store in memory
```

**Option B: Database (Production-ready)**
```sql
CREATE TABLE scheduled_notifications (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(255) UNIQUE,
    task_id INT,
    user_id INT,
    title VARCHAR(255),
    due_date TIMESTAMP,
    notified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### Step 2: Consumer Stores Events

```go
// In consumer.go
func (c *Consumer) ProcessMessage(msg kafka.Message) {
    // Parse event
    event := parseEvent(msg.Value)
    
    // Store in memory/database
    scheduledTask := ScheduledTask{
        EventID: event.EventID,
        TaskID:  event.TaskID,
        DueDate: event.DueDate,
    }
    scheduledTasks = append(scheduledTasks, scheduledTask)
    
    // Commit offset (mark as read)
    c.commitOffset(msg)
}
```

#### Step 3: Scheduler Checks Periodically

**Using Go's `time.Ticker`:**

```go
package main

import (
    "time"
    "context"
)

func startScheduler(ctx context.Context, checkInterval time.Duration) {
    ticker := time.NewTicker(checkInterval) // e.g., 1 hour
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Run check
            checkAndNotify()
        case <-ctx.Done():
            return
        }
    }
}

func checkAndNotify() {
    now := time.Now()
    
    // Find tasks due today or in the past
    for _, task := range scheduledTasks {
        if !task.Notified && task.DueDate.Before(now) || task.DueDate.Equal(now) {
            // Send notification
            sendNotification(task)
            task.Notified = true  // Mark as notified
        }
    }
}
```

**Using External Cron (Alternative):**
- Use `github.com/robfig/cron/v3` library
- More flexible: "Run at 9 AM every day"

```go
import "github.com/robfig/cron/v3"

c := cron.New()
c.AddFunc("0 * * * *", checkAndNotify)  // Every hour
c.Start()
```

#### Step 4: Complete Flow

```go
func main() {
    // 1. Start Kafka consumer (stores events)
    consumer := NewConsumer()
    go consumer.Start()
    
    // 2. Start scheduler (checks every hour)
    ctx := context.Background()
    go startScheduler(ctx, 1*time.Hour)
    
    // Keep running
    select {}
}
```

### Recommended Approach:
1. **Consumer**: Reads from Kafka → Stores in **database table**
2. **Scheduler**: Runs every **1 hour** → Queries database for due tasks → Sends notifications
3. **Mark as notified**: Update `notified = true` to avoid duplicate notifications

---

## 4. Dead Letter Queue (DLQ) - What and Where?

### What is Dead Letter Queue?

**Dead Letter Queue (DLQ)** = A special Kafka topic where **failed messages** go.

### Why Do We Need It?

**Problem Scenario:**
```
Consumer reads message → Tries to process → ERROR (invalid JSON, missing field, etc.)
What happens? Message is lost? Retry forever?
```

**Solution: Dead Letter Queue**
```
Consumer reads message → Tries to process → ERROR
→ Retry 3 times → Still fails?
→ Send to DLQ topic → Log error → Continue processing other messages
```

### How It Works:

```
Normal Flow:
Kafka Topic → Consumer → Process Successfully ✅

Error Flow:
Kafka Topic → Consumer → Process Fails ❌
→ Retry (3 times) → Still Fails ❌
→ Send to DLQ Topic → Log Error → Continue ✅
```

### Where to Put DLQ?

**Create a separate Kafka topic:**
```
Topics:
- task-scheduled (normal events)
- task-scheduled-dlq (failed events)  ← Dead Letter Queue
```

### Implementation:

```go
func (c *Consumer) ProcessMessage(msg kafka.Message) {
    // Try to parse
    event, err := parseEvent(msg.Value)
    if err != nil {
        // Send to DLQ
        c.sendToDLQ(msg, err)
        c.commitOffset(msg)  // Don't retry, move on
        return
    }
    
    // Try to process
    err = c.processEvent(event)
    if err != nil {
        retryCount := c.getRetryCount(msg)
        if retryCount < 3 {
            // Retry
            c.retryMessage(msg)
        } else {
            // Max retries reached → DLQ
            c.sendToDLQ(msg, err)
            c.commitOffset(msg)
        }
        return
    }
    
    // Success
    c.commitOffset(msg)
}

func (c *Consumer) sendToDLQ(msg kafka.Message, err error) {
    dlqMessage := DLQMessage{
        OriginalMessage: msg.Value,
        Error: err.Error(),
        Timestamp: time.Now(),
        Topic: msg.Topic,
        Partition: msg.Partition,
        Offset: msg.Offset,
    }
    
    // Send to DLQ topic
    c.producer.Send("task-scheduled-dlq", dlqMessage)
    
    // Log for monitoring
    log.Printf("Message sent to DLQ: %v", err)
}
```

### What to Do with DLQ Messages?

1. **Monitor**: Alert when DLQ has messages
2. **Investigate**: Check why messages failed
3. **Fix**: Correct the issue
4. **Replay**: Manually reprocess DLQ messages after fix

### Best Practice:
- **Retry 3 times** before sending to DLQ
- **Log everything** (for debugging)
- **Monitor DLQ size** (alert if growing)
- **Manual review** of DLQ messages

---

## 5. Event Versioning - What and How?

### What is Event Versioning?

**Event Versioning** = Adding a version number to events so you can **evolve** your event schema over time.

### Why Do We Need It?

**Problem:**
```
Version 1: {
  "task_id": 123,
  "due_date": "2024-11-09"
}

Later, you want to add fields:
Version 2: {
  "task_id": 123,
  "due_date": "2024-11-09",
  "priority": "high",        // NEW FIELD
  "reminder_time": "09:00"   // NEW FIELD
}
```

**What if consumer expects old format?** → Breaks!

### Solution: Version Field

```json
{
  "event_version": "1.0",  // ← Version number
  "event_id": "uuid",
  "event_type": "task_scheduled",
  "task_id": 123,
  "due_date": "2024-11-09"
}
```

### How to Implement:

#### Step 1: Add Version to Event Struct

```go
type TaskScheduledEvent struct {
    EventVersion string    `json:"event_version"`  // "1.0", "2.0", etc.
    EventID      string    `json:"event_id"`
    EventType    string    `json:"event_type"`
    TaskID       int       `json:"task_id"`
    DueDate      time.Time `json:"due_date"`
    // Version 1.0 fields above
    
    // Version 2.0 fields (optional, for backward compatibility)
    Priority     *string   `json:"priority,omitempty"`
    ReminderTime *string   `json:"reminder_time,omitempty"`
}
```

#### Step 2: Consumer Handles Different Versions

```go
func (c *Consumer) ProcessEvent(eventData []byte) error {
    // Parse base event (to get version)
    var baseEvent struct {
        EventVersion string `json:"event_version"`
    }
    json.Unmarshal(eventData, &baseEvent)
    
    // Route to version-specific handler
    switch baseEvent.EventVersion {
    case "1.0":
        return c.processV1Event(eventData)
    case "2.0":
        return c.processV2Event(eventData)
    default:
        return fmt.Errorf("unknown event version: %s", baseEvent.EventVersion)
    }
}

func (c *Consumer) processV1Event(data []byte) error {
    var event TaskScheduledEventV1
    json.Unmarshal(data, &event)
    // Process V1 format
}

func (c *Consumer) processV2Event(data []byte) error {
    var event TaskScheduledEventV2
    json.Unmarshal(data, &event)
    // Process V2 format (has priority, reminder_time)
}
```

#### Step 3: Migration Strategy

**Option A: Support Multiple Versions**
- Consumer supports V1 and V2
- Gradually migrate producers to V2
- Eventually deprecate V1

**Option B: Schema Registry** (Advanced)
- Use Confluent Schema Registry
- Validates schemas
- Handles compatibility automatically

### Best Practice:
- **Start with version "1.0"**
- **Always include version** in events
- **Support 2-3 versions** during migration
- **Document changes** between versions

---

## 6. Outbox Pattern vs Not Blocking API - Clarification

### The Confusion:

**Earlier I said:**
> "If Kafka fails, API should work (don't block API)"

**Then I said:**
> "Outbox pattern: Save to DB and Kafka in same transaction (both succeed or both fail)"

**These seem contradictory!** Let me clarify:

### Two Different Approaches:

#### Approach 1: **Fire-and-Forget** (What I Recommended First)

```
API Request → Save to DB ✅ → Send to Kafka → Return Response ✅
                      ↓
                  Kafka Fails ❌ → Log Error → Continue (don't fail request)
```

**Pros:**
- Fast (doesn't wait for Kafka)
- API never fails due to Kafka

**Cons:**
- **Inconsistency**: Task saved in DB but event not in Kafka
- Notifier might miss the task

#### Approach 2: **Outbox Pattern** (More Reliable)

```
API Request → Save to DB ✅ → Save event to "outbox" table ✅ → Return Response ✅
                      ↓
                  (Same Transaction - both succeed or both fail)
                      ↓
              Background Process → Reads outbox → Publishes to Kafka
```

**How It Works:**

1. **API saves task AND event in same transaction:**
```sql
BEGIN TRANSACTION;
  INSERT INTO tasks (...) VALUES (...);
  INSERT INTO outbox (event_data) VALUES ('{"task_id": 123, ...}');
COMMIT;  -- Both succeed or both fail
```

2. **Separate background process:**
```go
// Runs every few seconds
func outboxProcessor() {
    // Read unprocessed events from outbox
    events := db.GetUnprocessedOutboxEvents()
    
    for _, event := range events {
        // Publish to Kafka
        err := kafkaProducer.Send(event.Data)
        if err == nil {
            // Mark as processed
            db.MarkOutboxProcessed(event.ID)
        }
    }
}
```

**Pros:**
- **Consistency**: Task and event saved together
- **Reliability**: If Kafka is down, events queue in outbox
- **Guaranteed delivery**: Eventually all events reach Kafka

**Cons:**
- Slightly more complex
- Small delay (background process runs every few seconds)

### Which Should You Use?

**For Learning (Start Simple):**
- Use **Approach 1** (Fire-and-Forget)
- Log errors if Kafka fails
- Accept small risk of inconsistency

**For Production:**
- Use **Approach 2** (Outbox Pattern)
- Guarantees consistency
- More reliable

### Recommendation for Your Project:

**Start with Approach 1**, then **migrate to Approach 2** when you understand Kafka better.

---

## 7. Kafka Docker Setup - Step by Step

### Prerequisites:
- Docker and Docker Compose installed
- Your existing `docker-compose.yml` file

### Step 1: Add Kafka to docker-compose.yml

Add these services to your existing `docker-compose.yml`:

```yaml
services:
  # ... your existing services (cockroachdb, redis, api) ...

  # Zookeeper (required by Kafka)
  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    container_name: taskmanager-zookeeper
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "2181"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Kafka Broker
  kafka:
    image: confluentinc/cp-kafka:7.5.0
    container_name: taskmanager-kafka
    depends_on:
      zookeeper:
        condition: service_healthy
    ports:
      - "9092:9092"
      - "9093:9093"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092,PLAINTEXT_INTERNAL://kafka:29092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT_INTERNAL
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    healthcheck:
      test: ["CMD", "kafka-broker-api-versions", "--bootstrap-server", "localhost:9092"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### Step 2: Update Your API Service

Add Kafka dependency to your API service:

```yaml
  api:
    # ... existing config ...
    depends_on:
      # ... existing dependencies ...
      kafka:
        condition: service_healthy
    environment:
      # ... existing env vars ...
      KAFKA_BROKER_URL: kafka:29092  # Internal Docker network
      KAFKA_BROKER_URL_EXTERNAL: localhost:9092  # For local testing
```

### Step 3: Start Services

```bash
# Start all services (including Kafka)
docker-compose up -d

# Check if Kafka is running
docker-compose ps

# View Kafka logs
docker-compose logs kafka
```

### Step 4: Create Topic (Optional - Auto-create is enabled)

```bash
# Enter Kafka container
docker exec -it taskmanager-kafka bash

# Create topic manually (optional, auto-create is enabled)
kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic task-scheduled \
  --partitions 3 \
  --replication-factor 1

# List topics
kafka-topics --list --bootstrap-server localhost:9092
```

### Step 5: Test Kafka (Command Line)

**Terminal 1 - Producer:**
```bash
docker exec -it taskmanager-kafka kafka-console-producer \
  --bootstrap-server localhost:9092 \
  --topic task-scheduled
# Type messages and press Enter
```

**Terminal 2 - Consumer:**
```bash
docker exec -it taskmanager-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic task-scheduled \
  --from-beginning
# Should see messages from Terminal 1
```

### Step 6: Update Environment Variables

Add to your `.env` file:
```env
KAFKA_BROKER_URL=localhost:9092
```

### Step 7: Install Go Kafka Library

```bash
go get github.com/segmentio/kafka-go
# OR
go get github.com/confluentinc/confluent-kafka-go/kafka
# OR
go get github.com/IBM/sarama
```

### Common Issues:

1. **Port already in use**: Change port `9092` to something else
2. **Connection refused**: Make sure Kafka is healthy before starting API
3. **Topic not found**: Enable auto-create or create manually

### Verification:

```bash
# Check if Kafka is accessible
docker exec taskmanager-kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092

# Should see broker information
```

---

## 8. Architecture Review

### Current Architecture (Before Kafka):

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────┐
│  API Server │
│  (Fiber)    │
└──────┬──────┘
       │
       ├──→ CockroachDB (Tasks, Users)
       └──→ Redis (Cache)
```

### Proposed Architecture (With Kafka):

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────┐      ┌─────────────┐
│  API Server │─────→│   Kafka     │
│  (Fiber)    │      │   Broker    │
└──────┬──────┘      └──────┬──────┘
       │                    │
       ├──→ CockroachDB     │
       └──→ Redis           │
                            ▼
                    ┌─────────────┐
                    │  Notifier   │
                    │   Service   │
                    └─────────────┘
```

### Architecture Assessment:

#### ✅ **Good:**
1. **Separation of concerns**: API and Notifier are separate
2. **Scalable**: Can add more notifier instances
3. **Decoupled**: API doesn't know about notifications

#### ⚠️ **Considerations:**

1. **Database Schema:**
   - Need to add `due_date` field to `Task` model
   - Consider adding `scheduled_notifications` table (for scheduler)

2. **Configuration:**
   - Add Kafka config to `config.go`
   - Add Kafka URLs to environment variables

3. **Error Handling:**
   - Producer should not fail API requests
   - Consumer should handle parsing errors gracefully

4. **Monitoring:**
   - Log Kafka operations
   - Monitor consumer lag
   - Track failed messages

5. **Testing:**
   - Test with Kafka down (producer should handle)
   - Test consumer reconnection
   - Test message ordering

### Recommended Structure:

```
task_manager/
├── cmd/
│   ├── api/
│   │   └── main.go          # Existing API
│   └── notifier/
│       └── main.go          # New notifier service
├── internal/
│   ├── producer/            # New: Kafka producer
│   │   └── producer.go
│   ├── consumer/            # New: Kafka consumer
│   │   └── consumer.go
│   ├── notifier/            # New: Notification logic
│   │   └── notifier.go
│   ├── scheduler/           # New: Cron scheduler
│   │   └── scheduler.go
│   └── models/
│       └── events.go        # New: Event models
└── docker-compose.yml       # Updated with Kafka
```

### Next Steps:

1. ✅ Add Kafka to docker-compose.yml
2. ✅ Add `due_date` to Task model
3. ✅ Create producer in API service
4. ✅ Create notifier service with consumer
5. ✅ Implement scheduler
6. ✅ Add error handling and DLQ
7. ✅ Test end-to-end

---

## Summary

1. **UUID**: Unique identifier for events (prevents duplicates, enables tracing)
2. **Past Notifications**: Store in Kafka (auto-deleted after retention period), useful for replay/audit
3. **Scheduler**: Store events in DB, run cron job every hour to check due tasks
4. **DLQ**: Separate Kafka topic for failed messages (after retries)
5. **Event Versioning**: Add `event_version` field to support schema evolution
6. **Outbox Pattern**: More reliable than fire-and-forget (saves to DB + outbox in transaction)
7. **Kafka Setup**: Add Zookeeper + Kafka to docker-compose.yml
8. **Architecture**: Good separation, needs `due_date` field and event models

Ready to implement! 🚀

