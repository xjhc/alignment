#!/bin/bash
set -euo pipefail

# Remove any existing Redis database files that might cause compatibility issues
rm -f dump.rdb /tmp/dump.rdb

# Start Redis server
echo "Starting Redis server..."
redis-server --port 6379 --daemonize yes --logfile /tmp/redis.log

# Wait a moment for Redis to start
sleep 2

# Test Redis connection
if redis-cli ping > /dev/null 2>&1; then
    echo "Redis server started successfully"
else
    echo "ERROR: Failed to start Redis server"
    if [ -f /tmp/redis.log ]; then
        echo "Redis log:"
        cat /tmp/redis.log
    fi
    exit 1
fi