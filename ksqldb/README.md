# ksqlDB Stream Processing Demo

Kafkaストリーム処理のためのksqlDB環境とGoサンプルアプリケーション。

## アーキテクチャ

```
                                    ┌─────────────┐
                                    │   Go App    │
                                    │  (Producer) │
                                    └──────┬──────┘
                                           │ produce
                                           ▼
┌─────────────┐     ┌─────────┐     ┌─────────────────┐     ┌──────────────┐
│  Zookeeper  │────▶│  Kafka  │────▶│ Schema Registry │────▶│ ksqlDB Server│
│   :2181     │     │  :9092  │     │     :8082       │     │    :8088     │
└─────────────┘     └─────────┘     └─────────────────┘     └──────┬───────┘
                                                                   │
                                    ┌─────────────┐                │
                                    │   Go App    │◀───────────────┘
                                    │   (Query)   │  REST API
                                    └─────────────┘
```

## クイックスタート

```bash
# 1. インフラ起動
make up

# 2. ストリーム/テーブル作成
make sample

# 3. Goアプリでデモ実行
make run-demo
```

## デモ出力例

```
=== ksqlDB Stream Processing Demo ===

Step 1: Sending orders to Kafka...
Sent: O101 - MacBook Pro x1 ($2499.00)
Sent: O102 - iPhone 15 x2 ($999.00)
Sent: O103 - AirPods Pro x1 ($249.00)
Sent: O104 - iPad Air x1 ($799.00)
Sent: O105 - Apple Watch x1 ($399.00)
All orders sent!

Step 2: Waiting for ksqlDB to process...

Step 3: Querying aggregated results...

Customer Order Totals:
─────────────────────────────────────
  Customer: C001   | Orders: 3 | Total: $2748
  Customer: C002   | Orders: 2 | Total: $2397
─────────────────────────────────────

Demo complete!
```

## Go サンプルアプリケーション

### コマンド一覧

| コマンド | 説明 |
|---------|------|
| `make run-demo` | フルデモ（送信 → 処理待ち → クエリ） |
| `make run-produce` | 注文データをKafkaに送信 |
| `make run-query` | ksqlDBから集計結果を取得 |
| `make run-stream` | プッシュクエリでリアルタイム購読 |
| `make run-consume` | Kafkaから直接メッセージを消費 |

### アプリケーション構成

```
app/
├── main.go           # CLIエントリポイント
├── go.mod
├── Dockerfile
├── kafka/
│   ├── producer.go   # Kafka Producer (confluent-kafka-go)
│   └── consumer.go   # Kafka Consumer
└── ksqldb/
    └── client.go     # ksqlDB REST APIクライアント
```

### ksqlDB クライアント機能

```go
client := ksqldb.NewClient("http://localhost:8088")

// プルクエリ（テーブルから即時取得）
results, err := client.PullQuery("SELECT * FROM ORDER_TOTALS;")

// プッシュクエリ（ストリームをリアルタイム購読）
client.PushQuery(ctx, "SELECT * FROM orders EMIT CHANGES;", func(row map[string]interface{}) {
    fmt.Printf("New order: %v\n", row)
})

// DDL/DML文の実行
client.ExecuteStatement("CREATE STREAM ...")
```

## ksqlDB ストリーム処理

### ストリーム作成

```sql
CREATE STREAM orders (
  order_id VARCHAR KEY,
  customer_id VARCHAR,
  product VARCHAR,
  quantity INT,
  price DECIMAL(10,2)
) WITH (
  KAFKA_TOPIC = 'orders',
  PARTITIONS = 1,
  VALUE_FORMAT = 'JSON'
);
```

### 集計テーブル（リアルタイム更新）

```sql
CREATE TABLE order_totals AS
  SELECT
    customer_id,
    COUNT(*) AS order_count,
    SUM(quantity * price) AS total_amount
  FROM orders
  GROUP BY customer_id
  EMIT CHANGES;
```

### 派生ストリーム（フィルタリング）

```sql
CREATE STREAM high_value_orders AS
  SELECT * FROM orders
  WHERE (quantity * price) > 100
  EMIT CHANGES;
```

## Makefileコマンド

### インフラ

| コマンド | 説明 |
|---------|------|
| `make up` | 全コンテナ起動 |
| `make down` | 全コンテナ停止 |
| `make ps` | コンテナ状態確認 |
| `make logs` | ログ表示 |
| `make clean` | コンテナ+ボリューム削除 |

### ksqlDB

| コマンド | 説明 |
|---------|------|
| `make cli` | ksqlDB CLIに接続 |
| `make sample` | サンプルストリーム/テーブル作成 |
| `make streams` | ストリーム一覧 |
| `make tables` | テーブル一覧 |
| `make queries` | 実行中クエリ一覧 |
| `make topics` | Kafkaトピック一覧 |

### Goアプリ

| コマンド | 説明 |
|---------|------|
| `make build` | Dockerイメージビルド |
| `make run-demo` | デモ実行 |
| `make run-produce` | 注文データ送信 |
| `make run-query` | 集計結果クエリ |
| `make run-stream` | リアルタイム購読 |
| `make run-consume` | Kafka消費 |

## ポート一覧

| サービス | ポート | 用途 |
|---------|-------|------|
| Zookeeper | 2181 | Kafka調整 |
| Kafka | 9092 | 外部クライアント接続 |
| Kafka | 29092 | コンテナ間通信 |
| Schema Registry | 8082 | スキーマ管理 |
| ksqlDB Server | 8088 | REST API / CLI |

## 参考リンク

- [ksqlDB Documentation](https://docs.ksqldb.io/)
- [ksqlDB Syntax Reference](https://docs.ksqldb.io/en/latest/developer-guide/ksqldb-reference/)
- [ksqlDB REST API](https://docs.ksqldb.io/en/latest/developer-guide/api/)
- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)
