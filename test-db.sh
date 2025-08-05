#!/bin/bash

# Test Docker Compose commands for PostgreSQL test database

echo "=== PostgreSQL Test Database Commands ==="
echo

echo "Start test database:"
echo "docker-compose -f docker-compose.test.yml up -d"
echo

echo "Stop test database:"
echo "docker-compose -f docker-compose.test.yml down"
echo

echo "Stop and remove volumes (clean slate):"
echo "docker-compose -f docker-compose.test.yml down -v"
echo

echo "Connect to test database:"
echo "docker exec -it postgres-test psql -U testuser -d testdb"
echo

echo "View logs:"
echo "docker-compose -f docker-compose.test.yml logs -f postgres-test"
echo

echo "=== Connection Details ==="
echo "Host: localhost"
echo "Port: 5433"
echo "Database: testdb"
echo "User: testuser"
echo "Password: testpass"
echo

echo "=== Connection String ==="
echo "postgresql://testuser:testpass@localhost:5433/testdb"
