# Temporal Saga Pattern Demo

Temporalを使った分散トランザクション（Sagaパターン）のデモプロジェクト。

## アーキテクチャ

```
┌─────────────────────────────────────────────────────┐
│                 Temporal Server                      │
│                 (localhost:7233)                     │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                   Saga Worker                        │
│                                                      │
│  Workflow: OrderSagaWorkflow                         │
│  Activities:                                         │
│    - Step1Activity (注文作成)                         │
│    - Step2Activity (決済処理)                         │
│    - Step3Activity (在庫引当)                         │
│  Compensations:                                      │
│    - CompensateStep1 (注文取消)                       │
│    - CompensateStep2 (返金処理)                       │
│    - CompensateStep3 (在庫戻し)                       │
└─────────────────────────────────────────────────────┘
```

## Sagaパターンとは

Sagaパターンは、分散システムにおける長時間トランザクションを管理するためのパターン。
各ステップが失敗した場合、それまでに完了したステップを逆順で補償（ロールバック）する。

### 正常フロー
```
Step1 (実行) → Step2 (実行) → Step3 (実行) → 完了
```

### 失敗時の補償フロー（Step3で失敗した場合）
```
Step1 (実行) → Step2 (実行) → Step3 (失敗)
    → CompensateStep2 (返金) → CompensateStep1 (注文取消) → ロールバック完了
```

## セットアップ

### 必要なもの
- Docker & Docker Compose
- Go 1.23+

### 起動

```bash
# サービス起動
make up

# 起動確認
docker compose ps
```

### 停止

```bash
make down

# ボリュームも削除する場合
make clean
```

## 使い方

### 正常ケースの実行

```bash
make saga-run
```

出力例:
```
========================================
           SAGA RESULT
========================================
Order ID:    order-1234567890
Success:     true
Compensated: false

Steps executed:
  1. [✓] Step1: Order order-1234567890 created
  2. [✓] Step2: Payment for order-1234567890 processed
  3. [✓] Step3: Inventory for order-1234567890 reserved
========================================
```

### 失敗ケースの実行（補償処理）

```bash
# Step3で失敗させる（Step2, Step1の補償が実行される）
make saga-fail-step3

# Step2で失敗させる（Step1の補償が実行される）
make saga-fail-step2

# Step1で失敗させる（補償なし）
make saga-fail-step1
```

出力例（Step3失敗時）:
```
========================================
           SAGA RESULT
========================================
Order ID:    order-1234567890
Success:     false
Compensated: true
Error:       Step3 failed: simulated error

Steps executed:
  1. [✓] Step1: Order order-1234567890 created
  2. [✓] Step2: Payment for order-1234567890 processed
========================================
```

### カスタムOrder IDで実行

```bash
make saga-order ORDER=my-custom-order-123
```

## Temporal Web UI

http://localhost:8088 でワークフローの実行履歴を確認できる。

```bash
make temporal-ui  # ブラウザで開く
```

Web UIで確認できること:
- ワークフローの実行状態
- 各Activityの実行時間
- 失敗時のリトライ回数
- 補償処理の実行履歴

## ログ確認

```bash
# 全サービスのログ
make logs

# Saga Workerのログ
make logs-saga

# Temporalサーバーのログ
make logs-temporal
```

## プロジェクト構成

```
temporal/
├── docker-compose.yml      # Temporal Server + UI + Worker
├── Makefile                # コマンド集
├── README.md
├── saga-worker/            # Temporal Worker
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go             # Worker起動
│   ├── workflow.go         # Sagaワークフロー定義
│   └── activities.go       # 各ステップ + 補償処理
└── saga-client/            # テストクライアント
    ├── go.mod
    ├── go.sum
    └── main.go
```

## サービス一覧

| サービス | ポート | 説明 |
|---------|-------|------|
| temporal | 7233 | Temporal Server (gRPC) |
| temporal-ui | 8088 | Temporal Web UI |
| temporal-db | - | PostgreSQL (内部) |
| saga-worker | - | Saga Worker |

## コマンド一覧

```bash
make help
```

| コマンド | 説明 |
|---------|------|
| `make up` | サービス起動 |
| `make down` | サービス停止 |
| `make clean` | 全削除（ボリューム含む） |
| `make saga-run` | 正常ケース実行 |
| `make saga-fail-step1` | Step1で失敗 |
| `make saga-fail-step2` | Step2で失敗（補償あり） |
| `make saga-fail-step3` | Step3で失敗（補償あり） |
| `make temporal-ui` | Web UIを開く |
| `make logs` | 全ログ表示 |
| `make logs-saga` | Workerログ表示 |

## 参考リンク

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Saga Pattern](https://microservices.io/patterns/data/saga.html)
