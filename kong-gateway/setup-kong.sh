#!/bin/bash

# Wait for Kong to be ready
echo "Waiting for Kong to be ready..."
until curl -s http://localhost:18001/status > /dev/null 2>&1; do
    sleep 2
done
echo "Kong is ready!"

# Create gRPC Service
echo "Creating gRPC service..."
curl -s -X POST http://localhost:18001/services \
  --data "name=grpc-hello-service" \
  --data "protocol=grpc" \
  --data "host=grpc-server" \
  --data "port=50051"

echo ""

# Create gRPC Route
echo "Creating gRPC route..."
curl -s -X POST http://localhost:18001/services/grpc-hello-service/routes \
  --data "name=grpc-hello-route" \
  --data "protocols[]=grpc" \
  --data "paths[]=/hello.HelloService"

echo ""

# Verify configuration
echo ""
echo "=== Configured Services ==="
curl -s http://localhost:18001/services | jq '.data[] | {name, protocol, host, port}'

echo ""
echo "=== Configured Routes ==="
curl -s http://localhost:18001/routes | jq '.data[] | {name, protocols, paths}'

echo ""
echo "Kong gRPC configuration complete!"
echo ""
echo "Test commands:"
echo "  # Direct to gRPC server:"
echo "  grpcurl -plaintext -d '{\"name\": \"World\"}' localhost:50051 hello.HelloService/SayHello"
echo ""
echo "  # Through Kong:"
echo "  grpcurl -plaintext -d '{\"name\": \"World\"}' localhost:19080 hello.HelloService/SayHello"
