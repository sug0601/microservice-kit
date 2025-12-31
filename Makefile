.PHONY: up down setup test test-direct test-kong test-stream logs clean

# Start all services
up:
	docker compose up -d --build
	@echo "Waiting for services to start..."
	@sleep 15
	@echo "Services are starting. Run 'make setup' to configure Kong."

# Stop all services
down:
	docker compose down

# Setup Kong gRPC routing
setup:
	./setup-kong.sh

# Run all tests
test: test-direct test-kong test-stream

# Test direct connection to gRPC server
test-direct:
	@echo "=== Testing direct gRPC connection ==="
	grpcurl -plaintext -d '{"name": "Direct"}' localhost:50051 hello.HelloService/SayHello

# Test connection through Kong
test-kong:
	@echo "=== Testing gRPC through Kong ==="
	grpcurl -plaintext -proto proto/hello.proto -d '{"name": "Kong"}' localhost:19080 hello.HelloService/SayHello

# Test server streaming through Kong
test-stream:
	@echo "=== Testing gRPC server streaming through Kong ==="
	grpcurl -plaintext -proto proto/hello.proto -d '{"name": "Stream"}' localhost:19080 hello.HelloService/SayHelloServerStream

# List available gRPC services (via reflection)
list-services:
	@echo "=== Services on gRPC server ==="
	grpcurl -plaintext localhost:50051 list
	@echo ""
	@echo "=== Services through Kong ==="
	grpcurl -plaintext localhost:19080 list

# Describe HelloService
describe:
	grpcurl -plaintext localhost:50051 describe hello.HelloService

# View Kong configuration
kong-status:
	@echo "=== Kong Status ==="
	curl -s http://localhost:18001/status | jq
	@echo ""
	@echo "=== Kong Services ==="
	curl -s http://localhost:18001/services | jq '.data'
	@echo ""
	@echo "=== Kong Routes ==="
	curl -s http://localhost:18001/routes | jq '.data'

# View logs
logs:
	docker compose logs -f

# View Kong logs only
logs-kong:
	docker compose logs -f kong

# View gRPC server logs only
logs-grpc:
	docker compose logs -f grpc-server

# Clean up everything
clean:
	docker compose down -v
	docker rmi kong-grpc-server 2>/dev/null || true

# Full setup from scratch
all: up
	@sleep 20
	@$(MAKE) setup
	@echo ""
	@echo "=== Running tests ==="
	@$(MAKE) test

# Help
help:
	@echo "Kong + gRPC Demo Commands:"
	@echo ""
	@echo "  make up          - Start all services"
	@echo "  make down        - Stop all services"
	@echo "  make setup       - Configure Kong gRPC routing"
	@echo "  make test        - Run all tests"
	@echo "  make test-direct - Test direct gRPC connection"
	@echo "  make test-kong   - Test gRPC through Kong"
	@echo "  make test-stream - Test server streaming"
	@echo "  make list-services - List available gRPC services"
	@echo "  make describe    - Describe HelloService"
	@echo "  make kong-status - View Kong configuration"
	@echo "  make logs        - View all logs"
	@echo "  make logs-kong   - View Kong logs"
	@echo "  make logs-grpc   - View gRPC server logs"
	@echo "  make clean       - Clean up everything"
	@echo "  make all         - Full setup and test"
