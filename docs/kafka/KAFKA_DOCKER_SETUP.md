# Kafka Docker Setup - Exact Steps

## Quick Setup Guide

### Step 1: Update docker-compose.yml

Add these services to your existing `docker-compose.yml` file:

```yaml
services:
  # ... your existing services ...

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
      - "9092:9092"      # External access (from your host machine)
      - "9093:9093"      # Internal Docker network
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092,PLAINTEXT_INTERNAL://kafka:29092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT_INTERNAL
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"  # Auto-create topics (good for development)
    healthcheck:
      test: ["CMD", "kafka-broker-api-versions", "--bootstrap-server", "localhost:9092"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### Step 2: Update API Service Dependencies

Add Kafka to your API service's `depends_on`:

```yaml
  api:
    # ... existing config ...
    depends_on:
      cockroachdb:
        condition: service_healthy
      db-init:
        condition: service_completed_successfully
      redis:
        condition: service_healthy
      kafka:                    # ← Add this
        condition: service_healthy
    environment:
      # ... existing env vars ...
      KAFKA_BROKER_URL: kafka:29092  # ← Add this (for Docker network)
```

### Step 3: Start Services

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View Kafka logs
docker-compose logs kafka

# Follow logs in real-time
docker-compose logs -f kafka
```

### Step 4: Verify Kafka is Running

```bash
# Check if Kafka is healthy
docker exec taskmanager-kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092

# List topics (should be empty initially)
docker exec taskmanager-kafka kafka-topics \
  --list \
  --bootstrap-server localhost:9092
```

### Step 5: Test Kafka (Optional but Recommended)

**Terminal 1 - Producer:**
```bash
docker exec -it taskmanager-kafka kafka-console-producer \
  --bootstrap-server localhost:9092 \
  --topic test-topic
```
Then type messages and press Enter.

**Terminal 2 - Consumer:**
```bash
docker exec -it taskmanager-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic test-topic \
  --from-beginning
```
You should see messages from Terminal 1 appear here.

### Step 6: Create Your Topic (Optional)

Since `KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"`, topics are created automatically when first used. But you can create manually:

```bash
docker exec taskmanager-kafka kafka-topics \
  --create \
  --bootstrap-server localhost:9092 \
  --topic task-scheduled \
  --partitions 3 \
  --replication-factor 1
```

### Step 7: Update Environment Variables

Add to your `.env` file (for local development):

```env
# Kafka Configuration
KAFKA_BROKER_URL=localhost:9092
```

For Docker, use: `kafka:29092` (internal Docker network)

## Complete docker-compose.yml Example

Here's how your complete file should look:

```yaml
services:
  # CockroachDB Database
  cockroachdb:
    image: cockroachdb/cockroach:latest-v23.2
    container_name: taskmanager-cockroachdb
    command: start-single-node --insecure
    ports:
      - "26257:26257"
      - "8081:8080"  
    volumes:
      - cockroachdb-data:/cockroach/cockroach-data
    healthcheck:
      test: ["CMD-SHELL", "cockroach sql --insecure -e 'SELECT 1' || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  # Database Initialization
  db-init:
    image: cockroachdb/cockroach:latest-v23.2
    container_name: taskmanager-db-init
    entrypoint: ["/bin/sh", "-c"]
    command:
      - |
        sleep 5
        cockroach sql --insecure --host=cockroachdb:26257 --execute="CREATE DATABASE IF NOT EXISTS taskmanager;"
    depends_on:
      cockroachdb:
        condition: service_healthy
    restart: "no"

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: taskmanager-redis
    ports:
      - "6380:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Zookeeper (for Kafka)
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

  # Go API Application
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: taskmanager-api
    ports:
      - "3000:8080"
    volumes:
      - ./keys:/app/keys:ro
    environment:
      - DATABASE_URL=postgresql://root@cockroachdb:26257/taskmanager?sslmode=disable
      - REDIS_URL=redis:6379
      - PORT=8080
      - JWT_PRIVATE_KEY_PATH=/app/keys/private.pem
      - JWT_PUBLIC_KEY_PATH=/app/keys/public.pem
      - KAFKA_BROKER_URL=kafka:29092
    depends_on:
      cockroachdb:
        condition: service_healthy
      db-init:
        condition: service_completed_successfully
      redis:
        condition: service_healthy
      kafka:
        condition: service_healthy
    restart: unless-stopped

volumes:
  cockroachdb-data:
  redis-data:
```

## Troubleshooting

### Issue: Port 9092 already in use
**Solution:** Change port mapping:
```yaml
ports:
  - "9093:9092"  # Use different host port
```

### Issue: Kafka won't start
**Solution:** Check Zookeeper is healthy first:
```bash
docker-compose logs zookeeper
docker-compose ps zookeeper
```

### Issue: Can't connect from host machine
**Solution:** Use `localhost:9092` (external listener)

### Issue: Can't connect from Docker container
**Solution:** Use `kafka:29092` (internal listener)

### Issue: Topics not auto-creating
**Solution:** Check `KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"` is set

## Useful Commands

```bash
# View all topics
docker exec taskmanager-kafka kafka-topics --list --bootstrap-server localhost:9092

# Describe a topic
docker exec taskmanager-kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic task-scheduled

# Delete a topic (careful!)
docker exec taskmanager-kafka kafka-topics --delete --bootstrap-server localhost:9092 --topic task-scheduled

# View consumer groups
docker exec taskmanager-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list

# View consumer lag
docker exec taskmanager-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group notifier-service
```

## Next Steps

After Kafka is running:
1. Install Go Kafka library
2. Create producer in API service
3. Create consumer in notifier service
4. Test end-to-end

See `KAFKA_QUESTIONS_ANSWERED.md` for implementation details!

