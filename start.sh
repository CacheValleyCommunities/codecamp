#!/bin/bash

# Kill any processes using port 8082
echo "🔍 Checking for processes using port 8082..."
PIDS=$(lsof -ti :8082)

if [ ! -z "$PIDS" ]; then
    echo "⚠️  Found processes using port 8082: $PIDS"
    echo "💀 Killing processes..."
    kill -9 $PIDS
    sleep 1
    echo "✅ Port 8082 cleared"
else
    echo "✅ Port 8082 is already free"
fi

# Start the Go server
echo "🚀 Starting Go server..."
go run main.go
