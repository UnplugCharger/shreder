#!/bin/bash

# Kill any existing processes
pkill -f "shreder.*port" 2>/dev/null || true
sleep 1

# Start server 1 in background
echo "Starting server 1 on port 8060..."
./bin/shreder -port=:8060 -peers=http://localhost:8061 > server1.log 2>&1 &
SERVER1_PID=$!

# Wait a moment before starting server 2
sleep 2

# Start server 2 in background  
echo "Starting server 2 on port 8061..."
./bin/shreder -port=:8061 -peers=http://localhost:8060 > server2.log 2>&1 &
SERVER2_PID=$!

# Wait for servers to start
sleep 3

echo "Servers started. Running tests..."

# Run the test
./test_replication.sh

# Show server logs
echo -e "\n=== Server 1 Log ==="
tail -20 server1.log

echo -e "\n=== Server 2 Log ==="
tail -20 server2.log

# Clean up
echo "Cleaning up..."
kill $SERVER1_PID $SERVER2_PID 2>/dev/null
wait $SERVER1_PID $SERVER2_PID 2>/dev/null

echo "Test completed!"
