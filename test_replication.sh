#!/bin/bash

# Test script for distributed cache replication

echo "Testing distributed cache replication..."

# Wait a moment for servers to start
sleep 2

echo "Setting key 'test1' with value 'hello' on server 1 (port 8060)..."
curl -X POST http://localhost:8060/set \
  -H "Content-Type: application/json" \
  -d '{"key": "test1", "value": "hello"}'

echo -e "\n\nSetting key 'test2' with value 'world' on server 2 (port 8061)..."
curl -X POST http://localhost:8061/set \
  -H "Content-Type: application/json" \
  -d '{"key": "test2", "value": "world"}'

echo -e "\n\nWaiting for replication..."
sleep 2

echo -e "\n\nGetting key 'test1' from server 1:"
curl "http://localhost:8060/get?key=test1"

echo -e "\n\nGetting key 'test1' from server 2:"
curl "http://localhost:8061/get?key=test1"

echo -e "\n\nGetting key 'test2' from server 1:"
curl "http://localhost:8060/get?key=test2"

echo -e "\n\nGetting key 'test2' from server 2:"
curl "http://localhost:8061/get?key=test2"

echo -e "\n\nTest completed!"
