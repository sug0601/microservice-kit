# Debezium Server + Redis CDC Demo

PostgreSQLの変更をDebezium Serverでキャプチャし、Redis Streamsに送信するCDC（Change Data Capture）デモ環境。

## 構成

```
┌─────────────┐      ┌─────────────────┐      ┌─────────────┐
│  PostgreSQL │ ───▶ │ Debezium Server │ ───▶ │    Redis    │
│   (Source)  │ WAL  │     (CDC)       │      │  (Streams)  │
└─────────────┘      └─────────────────┘      └─────────────┘
```

| サービス | ポート | 説明 |
|----------|--------|------|
| PostgreSQL | 5432 | ソースDB（WAL論理レプリケーション有効） |
| Redis | 6379 | CDCイベントの送信先 |
| Debezium Server | 8080 | CDC処理エンジン |
| Redis Commander | 8081 | Redis管理UI |

## クイックスタート

```bash
# 環境を起動してサンプルデータを投入
make setup

# または個別に実行
make up          # 環境起動
make init-db     # テーブル作成
make seed        # サンプルデータ投入
```

## 使い方

### 基本操作

```bash
make up          # 環境起動
make down        # 環境停止
make clean       # 環境停止 + ボリューム削除
make ps          # コンテナ状態確認
make logs        # 全ログ確認
make logs-debezium  # Debezium Serverログ確認
```

### CDC動作確認

```bash
# データ操作（INSERT/UPDATE/DELETE）
make test-insert   # テストデータ挿入
make test-update   # テストデータ更新
make test-delete   # テストデータ削除

# Redis Streams確認
make redis-streams  # ストリーム内容表示
make redis-len      # イベント件数確認
make redis-keys     # キー一覧
```

### DB/Redis接続

```bash
make psql        # PostgreSQLに接続
make redis-cli   # Redisに接続
```

## CDCイベント形式

Redis Streamsに送信されるイベントの形式：

```json
{
  "payload": {
    "before": null,
    "after": {
      "id": 1,
      "name": "Alice",
      "email": "alice@example.com",
      "created_at": 1767225433210962
    },
    "op": "c",
    "source": {
      "table": "users",
      "schema": "public",
      "db": "testdb"
    }
  }
}
```

### 操作タイプ（op）

| op | 説明 |
|----|------|
| `c` | CREATE（INSERT） |
| `u` | UPDATE |
| `d` | DELETE |
| `r` | READ（スナップショット） |

## Redis Streamsの読み取り

```bash
# 全イベント取得
docker exec redis redis-cli XRANGE testdb.public.users - +

# 最新5件取得
docker exec redis redis-cli XREVRANGE testdb.public.users + - COUNT 5

# Consumer Groupで読み取り
docker exec redis redis-cli XGROUP CREATE testdb.public.users mygroup $ MKSTREAM
docker exec redis redis-cli XREADGROUP GROUP mygroup consumer1 COUNT 1 STREAMS testdb.public.users >
```

## 設定

環境変数（docker-compose.yml）で設定可能：

| 環境変数 | 説明 | デフォルト |
|----------|------|-----------|
| `DEBEZIUM_SOURCE_DATABASE_HOSTNAME` | PostgreSQLホスト | postgres |
| `DEBEZIUM_SOURCE_DATABASE_PORT` | PostgreSQLポート | 5432 |
| `DEBEZIUM_SOURCE_DATABASE_USER` | DBユーザー | postgres |
| `DEBEZIUM_SOURCE_DATABASE_PASSWORD` | DBパスワード | postgres |
| `DEBEZIUM_SOURCE_DATABASE_DBNAME` | DB名 | testdb |
| `DEBEZIUM_SOURCE_TOPIC_PREFIX` | トピックプレフィックス | testdb |
| `DEBEZIUM_SOURCE_SCHEMA_INCLUDE_LIST` | 対象スキーマ | public |
| `DEBEZIUM_SINK_REDIS_ADDRESS` | Redisアドレス | redis:6379 |

## フィルタリング（SMT）

Debezium ServerはSMT（Single Message Transform）を使用してイベントをフィルタリングできます。
本環境ではGroovyスクリプトによるフィルタリングが有効になっています。

### 現在の設定

```yaml
# docker-compose.yml
DEBEZIUM_TRANSFORMS: filter
DEBEZIUM_TRANSFORMS_FILTER_TYPE: io.debezium.transforms.Filter
DEBEZIUM_TRANSFORMS_FILTER_LANGUAGE: jsr223.groovy
DEBEZIUM_TRANSFORMS_FILTER_CONDITION: "value.op == 'c'"  # INSERTのみ
```

### フィルタ条件の例

| 条件 | CONDITION |
|------|-----------|
| INSERTのみ | `value.op == 'c'` |
| UPDATEのみ | `value.op == 'u'` |
| DELETEのみ | `value.op == 'd'` |
| INSERT + UPDATE | `value.op == 'c' \|\| value.op == 'u'` |
| DELETEを除外 | `value.op != 'd'` |
| 特定テーブルのみ | `value.source.table == 'users'` |
| 複合条件 | `value.op == 'c' && value.source.table == 'orders'` |

### フィルタを無効化する

docker-compose.ymlから以下の行を削除またはコメントアウト：

```yaml
# DEBEZIUM_TRANSFORMS: filter
# DEBEZIUM_TRANSFORMS_FILTER_TYPE: io.debezium.transforms.Filter
# DEBEZIUM_TRANSFORMS_FILTER_LANGUAGE: jsr223.groovy
# DEBEZIUM_TRANSFORMS_FILTER_CONDITION: "value.op == 'c'"
```

### テーブル/カラムフィルタ

SMTを使わない基本的なフィルタリング：

```yaml
# 特定テーブルのみキャプチャ
DEBEZIUM_SOURCE_TABLE_INCLUDE_LIST: public.users,public.orders

# 特定テーブルを除外
DEBEZIUM_SOURCE_TABLE_EXCLUDE_LIST: public.logs

# 特定カラムを除外（パスワード等）
DEBEZIUM_SOURCE_COLUMN_EXCLUDE_LIST: public.users.password
```

## UI

- **Redis Commander**: http://localhost:8081
  - Redis内のデータをブラウザで確認可能

## トラブルシューティング

### Debezium Serverが起動しない

```bash
# ログを確認
make logs-debezium

# PostgreSQLの準備待ち
docker-compose restart debezium-server
```

### CDCイベントが流れない

1. PostgreSQLのWAL設定を確認：
```bash
docker exec postgres psql -U postgres -c "SHOW wal_level;"
# → logical であること
```

2. レプリケーションスロットを確認：
```bash
docker exec postgres psql -U postgres -c "SELECT * FROM pg_replication_slots;"
```

### ポートが使用中

```bash
# 使用中のポートを確認
lsof -i :5432
lsof -i :6379

# 他のコンテナを停止
docker stop <container_name>
```
