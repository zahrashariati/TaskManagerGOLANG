#!/bin/bash

# Wait for CockroachDB to be ready
echo "Waiting for CockroachDB to be ready..."
sleep 5

# Create database
echo "Creating database 'taskmanager'..."
docker exec -it taskmanager-cockroachdb ./cockroach sql --insecure --execute="CREATE DATABASE IF NOT EXISTS taskmanager;"

echo "Database 'taskmanager' created successfully!"
echo "You can now start the API with: docker-compose up api"

