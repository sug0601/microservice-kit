.PHONY: run build clean test fmt lint

# 実行
run:
	go run main.go

# ビルド
build:
	go build -o bin/circuit-breaker main.go

# クリーン
clean:
	rm -rf bin/

# テスト
test:
	go test -v ./...

# フォーマット
fmt:
	go fmt ./...

# lint (golangci-lintが必要)
lint:
	golangci-lint run

# 依存関係の整理
tidy:
	go mod tidy

# ヘルプ
help:
	@echo "Usage:"
	@echo "  make run    - デモを実行"
	@echo "  make build  - バイナリをビルド"
	@echo "  make clean  - ビルド成果物を削除"
	@echo "  make test   - テストを実行"
	@echo "  make fmt    - コードをフォーマット"
	@echo "  make lint   - lintを実行"
	@echo "  make tidy   - go mod tidy"
