.PHONY: up down ps logs cli server-logs kafka-logs clean sample build run-demo run-produce run-query run-stream

up:
	docker compose up -d

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f

cli:
	docker exec -it ksqldb-cli ksql http://ksqldb-server:8088

server-logs:
	docker compose logs -f ksqldb-server

kafka-logs:
	docker compose logs -f kafka

clean:
	docker compose down -v

# ksqlDB sample setup
sample:
	@echo "Creating sample stream and table..."
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e \
		"CREATE STREAM IF NOT EXISTS orders (order_id VARCHAR KEY, customer_id VARCHAR, product VARCHAR, quantity INT, price DECIMAL(10,2)) WITH (KAFKA_TOPIC = 'orders', PARTITIONS = 1, VALUE_FORMAT = 'JSON');" 2>/dev/null || true
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e \
		"CREATE TABLE IF NOT EXISTS order_totals AS SELECT customer_id, COUNT(*) AS order_count, SUM(quantity * price) AS total_amount FROM orders GROUP BY customer_id EMIT CHANGES;" 2>/dev/null || true
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e \
		"CREATE STREAM IF NOT EXISTS high_value_orders AS SELECT * FROM orders WHERE (quantity * price) > 100 EMIT CHANGES;" 2>/dev/null || true
	@echo "Done! Run 'make cli' to query data."

streams:
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e "SHOW STREAMS;" 2>/dev/null

tables:
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e "SHOW TABLES;" 2>/dev/null

queries:
	@docker exec ksqldb-cli ksql http://ksqldb-server:8088 -e "SHOW QUERIES;" 2>/dev/null

topics:
	@docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Go App commands
build:
	docker compose build app

run-demo:
	docker compose run --rm app demo

run-produce:
	docker compose run --rm app produce

run-query:
	docker compose run --rm app query

run-stream:
	docker compose run --rm app stream

run-consume:
	docker compose run --rm app consume

# Local development (requires Go and librdkafka)
local-build:
	cd app && go build -o ksqldb-demo .

local-run:
	cd app && ./ksqldb-demo demo

help:
	@echo "Usage:"
	@echo ""
	@echo "Infrastructure:"
	@echo "  make up          - Start all containers"
	@echo "  make down        - Stop all containers"
	@echo "  make ps          - Show container status"
	@echo "  make logs        - Follow all logs"
	@echo "  make clean       - Stop and remove volumes"
	@echo ""
	@echo "ksqlDB:"
	@echo "  make cli         - Connect to ksqlDB CLI"
	@echo "  make sample      - Create sample streams/tables"
	@echo "  make streams     - List all streams"
	@echo "  make tables      - List all tables"
	@echo "  make queries     - List running queries"
	@echo "  make topics      - List Kafka topics"
	@echo ""
	@echo "Go App:"
	@echo "  make build       - Build the Go app Docker image"
	@echo "  make run-demo    - Run full demo (produce + query)"
	@echo "  make run-produce - Produce sample orders to Kafka"
	@echo "  make run-query   - Query aggregated data from ksqlDB"
	@echo "  make run-stream  - Subscribe to ksqlDB push query"
	@echo "  make run-consume - Consume orders from Kafka"
