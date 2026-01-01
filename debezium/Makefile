.PHONY: up down logs ps clean build rebuild test-insert test-update test-delete redis-streams redis-keys

# 環境起動
up:
	docker-compose up -d

# イメージビルド
build:
	docker-compose build

# イメージ再ビルド（キャッシュなし）
rebuild:
	docker-compose build --no-cache debezium-server

# 環境停止
down:
	docker-compose down

# 環境停止（ボリューム削除）
clean:
	docker-compose down -v

# ログ確認
logs:
	docker-compose logs -f

# Debezium Serverのログ確認
logs-debezium:
	docker logs -f debezium-server

# コンテナ状態確認
ps:
	docker-compose ps

# テストデータ挿入
test-insert:
	docker exec -i postgres psql -U postgres -d testdb -c \
		"INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com');"

# テストデータ更新
test-update:
	docker exec -i postgres psql -U postgres -d testdb -c \
		"UPDATE users SET email = 'updated@example.com' WHERE name = 'Test User';"

# テストデータ削除
test-delete:
	docker exec -i postgres psql -U postgres -d testdb -c \
		"DELETE FROM users WHERE name = 'Test User';"

# Redis Streamsの内容確認
redis-streams:
	docker exec redis redis-cli XRANGE testdb.public.users - +

# Redis Streamsの件数確認
redis-len:
	docker exec redis redis-cli XLEN testdb.public.users

# Redisキー一覧
redis-keys:
	docker exec redis redis-cli KEYS '*'

# PostgreSQLに接続
psql:
	docker exec -it postgres psql -U postgres -d testdb

# Redisに接続
redis-cli:
	docker exec -it redis redis-cli

# 初期テーブル作成
init-db:
	docker exec -i postgres psql -U postgres -d testdb -c \
		"CREATE TABLE IF NOT EXISTS users ( \
			id SERIAL PRIMARY KEY, \
			name VARCHAR(100) NOT NULL, \
			email VARCHAR(100) NOT NULL, \
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP \
		);"

# サンプルデータ投入
seed:
	docker exec -i postgres psql -U postgres -d testdb -c \
		"INSERT INTO users (name, email) VALUES \
			('Alice', 'alice@example.com'), \
			('Bob', 'bob@example.com');"

# 初期化（起動 + テーブル作成 + サンプルデータ）
setup: up
	@echo "Waiting for services to be ready..."
	@sleep 10
	@$(MAKE) init-db
	@$(MAKE) seed
	@echo "Setup complete!"
