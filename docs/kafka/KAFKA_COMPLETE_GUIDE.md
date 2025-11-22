# Complete Kafka Learning Guide for Task Scheduling System

## Table of Contents
1. [What is Kafka?](#what-is-kafka)
2. [Core Concepts](#core-concepts)
3. [Why Kafka for Your Use Case?](#why-kafka-for-your-use-case)
4. [Architecture Overview](#architecture-overview)
5. [Key Components Explained](#key-components-explained)
6. [Event-Driven Architecture](#event-driven-architecture)
7. [Implementation Guide](#implementation-guide)
8. [Best Practices](#best-practices)
9. [Common Patterns](#common-patterns)
10. [Troubleshooting](#troubleshooting)

---

## What is Kafka?

Apache Kafka is a **distributed event streaming platform** that allows you to:
- **Publish** (write) and **subscribe** (read) to streams of events
- **Store** streams of events in a fault-tolerant way
- **Process** streams of events as they occur

Think of it as a **message queue** or **event bus** that multiple services can use to communicate with each other asynchronously.

### Real-World Analogy
Imagine Kafka as a **post office system**:
- **Producers** = People sending letters (your API service)
- **Topics** = Different mailboxes (e.g., "task-notifications", "user-events")
- **Brokers** = Post offices (Kafka servers)
- **Consumers** = People receiving letters (your notifier service)
- **Messages** = The actual letters (events)

---

## Core Concepts

### 1. **Producer**
- **What**: A service/application that **sends** messages to Kafka
- **In your case**: Your task manager API that creates scheduled tasks
- **Role**: Publishes events like "Task scheduled for 11/9/2024"

### 2. **Consumer**
- **What**: A service/application that **reads** messages from Kafka
- **In your case**: Your notifier service that reads scheduled tasks
- **Role**: Consumes events and performs actions (like sending notifications)

### 3. **Topic**
- **What**: A category or feed name where messages are stored
- **Example**: `task-scheduled`, `task-reminders`, `user-notifications`
- **Think of it**: Like a database table, but for events
- **Properties**:
  - Messages are stored in **partitions** (for scalability)
  - Messages are **ordered** within a partition
  - Messages are **immutable** (can't be changed once written)

### 4. **Partition**
- **What**: Topics are divided into partitions for parallel processing
- **Why**: Allows multiple consumers to read from the same topic simultaneously
- **Example**: Topic `task-scheduled` might have 3 partitions
- **Key Point**: Messages with the same key go to the same partition (ensures order)

### 5. **Broker**
- **What**: A Kafka server that stores topics and handles read/write requests
- **In production**: Usually multiple brokers (cluster) for fault tolerance
- **For learning**: One broker is fine

### 6. **Consumer Group**
- **What**: A group of consumers working together to consume messages
- **How it works**: Each partition is consumed by only ONE consumer in a group
- **Benefit**: Allows parallel processing and load balancing
- **Example**: If you have 3 partitions and 3 consumers in a group, each consumer handles one partition

### 7. **Offset**
- **What**: A unique identifier for each message in a partition
- **Purpose**: Tracks which messages a consumer has already read
- **Important**: Consumers remember their position using offsets

### 8. **Event/Message**
- **What**: The actual data being sent
- **Structure**: Usually JSON with metadata (timestamp, key, value)
- **In your case**: 
  ```json
  {
    "task_id": 123,
    "user_id": 456,
    "title": "Review project proposal",
    "due_date": "2024-11-09T10:00:00Z",
    "event_type": "task_scheduled"
  }
  ```

---

## Why Kafka for Your Use Case?

### Problem You're Solving
You want to schedule tasks (like "Task X should be done on 11/9") and have a separate service notify users when that date arrives.

### Why Kafka is Perfect Here

1. **Decoupling**: Your API service doesn't need to know about notification logic
   - API creates task → publishes to Kafka → done!
   - Notifier service reads from Kafka → sends notification → done!

2. **Scalability**: 
   - If you have millions of scheduled tasks, Kafka handles it
   - You can add more notifier services without changing your API

3. **Reliability**:
   - Messages are persisted (not lost if service crashes)
   - Consumers can replay messages if needed

4. **Asynchronous Processing**:
   - API doesn't wait for notifications to be sent
   - Fast response times for users

5. **Event-Driven Architecture**:
   - Perfect for "something happened, react to it" scenarios
   - Easy to add more consumers later (e.g., analytics, logging)

---

## Architecture Overview

### Your System Architecture

```
┌─────────────────┐
│   Task Manager  │
│      API        │
│  (Producer)     │
└────────┬────────┘
         │
         │ Publishes event: "Task scheduled for 11/9"
         │
         ▼
┌─────────────────┐
│   Kafka Topic   │
│ "task-scheduled"│
│                 │
│  [Partition 0]  │
│  [Partition 1]  │
│  [Partition 2]  │
└────────┬────────┘
         │
         │ Consumes events
         │
         ▼
┌─────────────────┐
│  Notifier       │
│  Service        │
│  (Consumer)     │
└─────────────────┘
```

### Flow Example

1. **User creates task** via API:
   ```
   POST /tasks
   {
     "title": "Review proposal",
     "due_date": "2024-11-09"
   }
   ```

2. **API Service** (Producer):
   - Saves task to database
   - Publishes event to Kafka topic `task-scheduled`
   - Returns success to user

3. **Kafka** stores the event:
   ```
   Topic: task-scheduled
   Message: {
     "task_id": 123,
     "due_date": "2024-11-09T00:00:00Z",
     "user_id": 456,
     "title": "Review proposal"
   }
   ```

4. **Notifier Service** (Consumer):
   - Reads events from Kafka
   - Checks if due_date matches today
   - Prints notification: "Task 'Review proposal' is due today!"

---

## Key Components Explained

### Producer Details

**What it does:**
- Connects to Kafka broker
- Sends messages to a topic
- Can specify a key (for partitioning)
- Can handle errors and retries

**Key Concepts:**
- **Acknowledgment (acks)**: 
  - `acks=0`: Fire and forget (fastest, least reliable)
  - `acks=1`: Wait for leader acknowledgment (default)
  - `acks=all`: Wait for all replicas (most reliable)

- **Batching**: Groups multiple messages together for efficiency

- **Idempotency**: Ensures messages aren't duplicated

### Consumer Details

**What it does:**
- Connects to Kafka broker
- Subscribes to one or more topics
- Reads messages in order (within a partition)
- Commits offsets to track progress

**Key Concepts:**
- **Consumer Groups**: 
  - Multiple consumers can be in the same group
  - Each partition consumed by only one consumer in group
  - Enables parallel processing

- **Offset Management**:
  - **Auto-commit**: Kafka automatically commits offsets
  - **Manual commit**: You control when to commit (more reliable)

- **Rebalancing**: When consumers join/leave, partitions are redistributed

### Topic Configuration

**Important Settings:**
- **Replication Factor**: How many copies of data (for fault tolerance)
- **Partition Count**: How many partitions (affects parallelism)
- **Retention**: How long to keep messages (default: 7 days)
- **Compaction**: Keep only latest value for each key (useful for state)

---

## Event-Driven Architecture

### What is Event-Driven Architecture?

Instead of services calling each other directly, services communicate through **events**.

### Traditional (Synchronous) Approach:
```
API Service → Notifier Service (HTTP call)
API waits for response
```

**Problems:**
- Tight coupling
- If notifier is down, API fails
- Hard to scale
- Slow (waits for notification)

### Event-Driven (Asynchronous) Approach:
```
API Service → Kafka → Notifier Service
API doesn't wait
```

**Benefits:**
- Loose coupling
- If notifier is down, events queue up
- Easy to scale (add more consumers)
- Fast (API responds immediately)

### Event Types in Your System

1. **Task Scheduled Event**:
   - When: User creates/updates task with due_date
   - Payload: Task details + due_date
   - Consumer: Notifier service

2. **Task Completed Event** (future):
   - When: User marks task as complete
   - Payload: Task ID + completion time
   - Consumer: Analytics service (future)

3. **Task Deleted Event** (future):
   - When: User deletes task
   - Payload: Task ID
   - Consumer: Notifier service (to cancel scheduled notifications)

---

## Implementation Guide

### Phase 1: Setup Kafka

#### Option A: Docker (Recommended for Learning)
```bash
# Use docker-compose to run Kafka + Zookeeper
# Zookeeper: Kafka's metadata manager (required)
```

#### Option B: Local Installation
- Download Kafka from Apache website
- Requires Java
- More complex setup

### Phase 2: Design Your Events

**Event Schema:**
```json
{
  "event_id": "uuid",
  "event_type": "task_scheduled",
  "timestamp": "2024-11-09T10:00:00Z",
  "task_id": 123,
  "user_id": 456,
  "title": "Review proposal",
  "description": "Review the Q4 proposal",
  "due_date": "2024-11-09T17:00:00Z",
  "priority": "high"
}
```

**Key Decisions:**
- Use JSON for simplicity (or Avro/Protobuf for production)
- Include all necessary data (notifier shouldn't query database)
- Add event_id for deduplication
- Add timestamp for event ordering

### Phase 3: Modify Task Model

**Add to Task struct:**
```go
type Task struct {
    // ... existing fields ...
    DueDate *time.Time `json:"due_date" db:"due_date"` // Optional field
}
```

**Why optional?** Not all tasks have due dates.

### Phase 4: Create Producer in API Service

**Where:** In your task handler/service

**When to produce:**
- When creating a task with due_date
- When updating a task's due_date

**Steps:**
1. Initialize Kafka producer
2. When task created/updated with due_date:
   - Save to database (existing logic)
   - Create event message
   - Send to Kafka topic `task-scheduled`
   - Handle errors (log, don't fail the request)

**Important:** Don't fail the HTTP request if Kafka is down! Log the error and continue.

### Phase 5: Create Notifier Service

**New Service Structure:**
```
cmd/notifier/
  main.go

internal/
  consumer/
    consumer.go
  notifier/
    notifier.go
```

**Consumer Logic:**
1. Connect to Kafka
2. Subscribe to `task-scheduled` topic
3. Read messages in a loop
4. For each message:
   - Parse JSON
   - Check if due_date matches today (or is in the past)
   - Print notification
   - Commit offset

**Scheduling Logic:**
- Option 1: Check every message immediately (simple)
- Option 2: Store in memory, check periodically (more efficient)
- Option 3: Use a scheduler (cron-like) to check stored tasks

### Phase 6: Error Handling

**Producer Errors:**
- Network issues → Retry with exponential backoff
- Kafka down → Log error, continue (don't block API)
- Invalid message → Log and skip

**Consumer Errors:**
- Parse error → Log and skip message
- Processing error → Log, commit offset anyway (or retry)
- Kafka connection lost → Reconnect automatically

---

## Best Practices

### 1. **Idempotency**
- Make operations idempotent (safe to retry)
- Use event_id to detect duplicates
- Example: Don't send same notification twice

### 2. **Error Handling**
- Never fail the main operation if Kafka fails
- Use dead letter queue for failed messages
- Log everything

### 3. **Message Design**
- Include all necessary data (avoid database lookups in consumer)
- Use consistent schema
- Version your events (for future changes)

### 4. **Consumer Groups**
- Use meaningful group names: `notifier-service-v1`
- Each service instance should be in same group (for load balancing)

### 5. **Monitoring**
- Monitor lag (how far behind consumer is)
- Monitor throughput (messages/second)
- Set up alerts for errors

### 6. **Testing**
- Test with Kafka down (producer should handle gracefully)
- Test message ordering
- Test consumer rebalancing

---

## Common Patterns

### Pattern 1: Event Sourcing
Store all events, reconstruct state from events.

**Not needed for your use case** (you have database).

### Pattern 2: CQRS (Command Query Responsibility Segregation)
Separate write (commands) and read (queries) models.

**Your case:** API writes, notifier reads (similar concept).

### Pattern 3: Saga Pattern
Distributed transactions across services.

**Not needed yet**, but useful when you have multiple services.

### Pattern 4: Outbox Pattern
Write to database and Kafka in transaction.

**Your case:** 
1. Save task to DB
2. Save event to "outbox" table in same transaction
3. Separate process reads outbox and publishes to Kafka

**Why:** Ensures consistency (both succeed or both fail).

---

## Troubleshooting

### Common Issues

1. **Consumer not receiving messages**
   - Check consumer group name
   - Check topic name (case-sensitive)
   - Check offset (might be at end)
   - Check if messages were actually sent

2. **Messages duplicated**
   - Check idempotency settings
   - Check retry logic
   - Use event_id for deduplication

3. **Consumer lagging**
   - Too many messages, not enough consumers
   - Consumer processing too slow
   - Add more consumers or optimize processing

4. **Connection errors**
   - Check Kafka broker address
   - Check network/firewall
   - Check Kafka is running

---

## Step-by-Step Implementation Plan

### Step 1: Setup Kafka
- [ ] Install Kafka (Docker recommended)
- [ ] Create topic: `task-scheduled`
- [ ] Test with command-line tools (kafka-console-producer/consumer)

### Step 2: Add Due Date to Task Model
- [ ] Add `DueDate` field to Task struct
- [ ] Update database schema (migration)
- [ ] Update API handlers to accept due_date

### Step 3: Implement Producer
- [ ] Add Kafka client library (e.g., `confluent-kafka-go` or `sarama`)
- [ ] Initialize producer in main.go
- [ ] Create event struct
- [ ] Publish event when task created/updated with due_date
- [ ] Handle errors gracefully

### Step 4: Implement Consumer (Notifier Service)
- [ ] Create new service structure
- [ ] Initialize Kafka consumer
- [ ] Subscribe to `task-scheduled` topic
- [ ] Read messages in loop
- [ ] Parse and process events
- [ ] Print notifications
- [ ] Handle errors and reconnection

### Step 5: Testing
- [ ] Test producer (create task, verify message in Kafka)
- [ ] Test consumer (verify it reads and prints)
- [ ] Test error scenarios (Kafka down, invalid messages)
- [ ] Test with multiple consumers (consumer groups)

### Step 6: Deployment
- [ ] Add Kafka to docker-compose.yml
- [ ] Configure environment variables
- [ ] Update documentation

---

## Key Libraries for Go

### Option 1: `confluent-kafka-go`
- Official Confluent library
- C library wrapper
- Requires C dependencies
- Most feature-complete

### Option 2: `sarama`
- Pure Go implementation
- No C dependencies
- Good performance
- Popular choice

### Option 3: `kafka-go`
- Simple, idiomatic Go
- Easy to use
- Good for learning

**Recommendation for learning:** Start with `kafka-go` (simplest), then move to `sarama` for production.

---

## What "Delegate" Means

Your mentor likely meant **"decouple"** or **"delegate"**:

- **Decouple**: Separate concerns (API doesn't know about notifications)
- **Delegate**: Let Kafka handle message delivery (don't do it yourself)

Both concepts apply here!

---

## Next Steps After Basic Implementation

1. **Add more event types**: Task completed, task deleted
2. **Add multiple consumers**: Analytics service, audit log service
3. **Add message filtering**: Consumer only processes certain events
4. **Add retry logic**: Dead letter queue for failed messages
5. **Add monitoring**: Track message rates, lag, errors
6. **Add schema registry**: Validate message schemas
7. **Add compression**: Reduce network usage
8. **Add encryption**: Secure messages in transit

---

## Summary

**Kafka is perfect for your use case because:**
- ✅ Decouples API from notification logic
- ✅ Handles high volume of scheduled tasks
- ✅ Reliable (messages aren't lost)
- ✅ Scalable (add more notifiers easily)
- ✅ Asynchronous (fast API responses)

**Your architecture:**
1. API creates task → publishes to Kafka
2. Kafka stores event
3. Notifier reads from Kafka → prints notification

**Key concepts to remember:**
- **Producer** = sends messages
- **Consumer** = reads messages
- **Topic** = message category
- **Partition** = parallel processing
- **Consumer Group** = load balancing

**Start simple:**
1. Setup Kafka locally
2. Add producer to API
3. Create notifier service
4. Test end-to-end
5. Iterate and improve

Good luck! 🚀





