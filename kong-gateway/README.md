# Kong + gRPC Demo

Kong API Gatewayを使用してgRPCリクエストをプロキシするデモプロジェクト。

## アーキテクチャ

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Client    │─────▶│    Kong     │─────▶│ gRPC Server │
│  (grpcurl)  │      │  (Gateway)  │      │   (Go)      │
└─────────────┘      └─────────────┘      └─────────────┘
                            │
                     ┌──────┴──────┐
                     │             │
              ┌──────▼─────┐ ┌─────▼─────┐
              │ PostgreSQL │ │   Konga   │
              │    (DB)    │ │   (GUI)   │
              └────────────┘ └───────────┘
```

## サービス構成

| サービス | ポート | 説明 |
|---------|--------|------|
| Kong Proxy | 18000 (HTTP), 19080 (gRPC) | APIゲートウェイ |
| Kong Admin API | 18001 | 管理API |
| Kong Admin GUI | 18002 | 管理画面 |
| gRPC Server | 50051 | バックエンドgRPCサーバー |
| Konga | 1337 | Kong管理GUI |
| PostgreSQL | - | Kongデータベース |

## クイックスタート

### 前提条件

- Docker & Docker Compose
- grpcurl（テスト用）
- jq（オプション、JSONフォーマット用）

### 起動

```bash
# 全サービス起動
make up

# Kongルーティング設定
make setup

# または一括実行
make all
```

### テスト

```bash
# 全テスト実行
make test

# gRPCサーバー直接接続テスト
make test-direct

# Kong経由テスト
make test-kong

# サーバーストリーミングテスト
make test-stream
```

## gRPC API

### HelloService

```protobuf
service HelloService {
  rpc SayHello (HelloRequest) returns (HelloResponse);
  rpc SayHelloServerStream (HelloRequest) returns (stream HelloResponse);
}
```

### 手動テスト

```bash
# Kong経由でリクエスト
grpcurl -plaintext -proto proto/hello.proto \
  -d '{"name": "World"}' \
  localhost:19080 hello.HelloService/SayHello

# 直接gRPCサーバーにリクエスト
grpcurl -plaintext \
  -d '{"name": "World"}' \
  localhost:50051 hello.HelloService/SayHello
```

## Docker Compose サービス詳細

### kong-database

Kongのメタデータを保存するPostgreSQLデータベース。

### kong-migration

Kongのデータベースマイグレーションを実行するサービス。
- `kong migrations bootstrap` コマンドで初期スキーマを作成
- 初回起動時のみ実行され、完了後に終了
- このサービスが完了してからKong本体が起動

### kong

Kong API Gateway本体。gRPCプロキシとして動作。

### grpc-server

Go製のgRPCサーバー。HelloServiceを実装。

### konga

Kong管理用のWebインターフェース。

## Makeコマンド一覧

| コマンド | 説明 |
|---------|------|
| `make up` | 全サービス起動 |
| `make down` | 全サービス停止 |
| `make setup` | Kongルーティング設定 |
| `make test` | 全テスト実行 |
| `make test-direct` | gRPCサーバー直接テスト |
| `make test-kong` | Kong経由テスト |
| `make test-stream` | ストリーミングテスト |
| `make list-services` | gRPCサービス一覧 |
| `make describe` | HelloService詳細表示 |
| `make kong-status` | Kong設定確認 |
| `make logs` | 全ログ表示 |
| `make logs-kong` | Kongログのみ表示 |
| `make logs-grpc` | gRPCサーバーログのみ表示 |
| `make clean` | 全クリーンアップ（データ含む） |
| `make all` | 起動→設定→テスト一括実行 |

## ディレクトリ構成

```
.
├── docker-compose.yml    # Docker Compose設定
├── Makefile              # ビルド・テストコマンド
├── setup-kong.sh         # Kongルーティング設定スクリプト
├── proto/
│   └── hello.proto       # gRPCサービス定義
└── server/
    ├── Dockerfile        # gRPCサーバー用Dockerfile
    ├── go.mod            # Goモジュール定義
    └── main.go           # gRPCサーバー実装
```

## トラブルシューティング

### Kongが起動しない

```bash
# ログ確認
make logs-kong

# マイグレーション状態確認
docker compose logs kong-migration
```

### gRPC接続エラー

```bash
# gRPCサーバーの状態確認
make logs-grpc

# Kongルーティング確認
make kong-status
```

### 完全リセット

```bash
make clean
make all
```
