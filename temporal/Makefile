.PHONY: up down logs saga-setup saga-run saga-fail-step1 saga-fail-step2 saga-fail-step3 temporal-ui help

# Start all services
up:
	docker compose up -d --build
	@echo "Waiting for Temporal to start..."
	@sleep 30
	@echo "Services started. Visit http://localhost:8088 for Temporal UI"

# Stop all services
down:
	docker compose down

# View all logs
logs:
	docker compose logs -f

# View Temporal logs
logs-temporal:
	docker compose logs -f temporal

# View Saga worker logs
logs-saga:
	docker compose logs -f saga-worker

# Setup saga-client dependencies
saga-setup:
	cd saga-client && go mod tidy

# Run saga workflow (success case)
saga-run:
	@echo "=== Running Saga Workflow (Success) ==="
	cd saga-client && go run main.go

# Run saga workflow with failure at step1
saga-fail-step1:
	@echo "=== Running Saga Workflow (Fail at Step1) ==="
	cd saga-client && go run main.go -fail step1

# Run saga workflow with failure at step2
saga-fail-step2:
	@echo "=== Running Saga Workflow (Fail at Step2) ==="
	cd saga-client && go run main.go -fail step2

# Run saga workflow with failure at step3
saga-fail-step3:
	@echo "=== Running Saga Workflow (Fail at Step3) ==="
	cd saga-client && go run main.go -fail step3

# Run saga with custom order ID
saga-order:
	cd saga-client && go run main.go -order $(ORDER)

# Open Temporal Web UI
temporal-ui:
	@echo "Opening Temporal Web UI at http://localhost:8088"
	open http://localhost:8088 || xdg-open http://localhost:8088 || echo "Visit http://localhost:8088"

# Clean up everything
clean:
	docker compose down -v

# Help
help:
	@echo "Temporal Saga Demo Commands:"
	@echo ""
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - View all logs"
	@echo "  make logs-temporal  - View Temporal logs"
	@echo "  make logs-saga      - View Saga worker logs"
	@echo ""
	@echo "  make saga-setup     - Setup saga-client dependencies"
	@echo "  make saga-run       - Run saga (success case)"
	@echo "  make saga-fail-step1 - Run saga with failure at step1"
	@echo "  make saga-fail-step2 - Run saga with failure at step2 (compensation)"
	@echo "  make saga-fail-step3 - Run saga with failure at step3 (compensation)"
	@echo "  make saga-order ORDER=xxx - Run saga with custom order ID"
	@echo ""
	@echo "  make temporal-ui    - Open Temporal Web UI"
	@echo "  make clean          - Clean up everything"
