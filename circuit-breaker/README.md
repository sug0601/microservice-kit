# Circuit Breaker Pattern

sony/gobreakerを使ったCircuit Breakerパターンの実装例。

## Circuit Breakerとは

外部サービスへの呼び出しを監視し、障害を検知したら一時的にリクエストを遮断するパターン。

```
┌─────────────────────────────────────────────────────────────┐
│                    Circuit Breaker                          │
│                                                             │
│   ┌────────┐    失敗が閾値超え    ┌────────┐               │
│   │ Closed │ ──────────────────→ │  Open  │               │
│   │ (正常) │                      │ (遮断) │               │
│   └────────┘                      └────────┘               │
│       ↑                               │                     │
│       │                               │ Timeout経過         │
│       │                               ↓                     │
│       │    成功が閾値超え      ┌───────────┐               │
│       └─────────────────────── │ Half-Open │               │
│                                │ (様子見)  │               │
│            失敗 ──────────────→└───────────┘               │
│                    (Openに戻る)                             │
└─────────────────────────────────────────────────────────────┘
```

## 3つの状態

| 状態 | 説明 |
|------|------|
| **Closed** | 正常状態。リクエストを通す。失敗をカウント中 |
| **Open** | 障害検知。リクエストを即座にエラーで返す（APIは呼ばない） |
| **Half-Open** | 回復確認中。一部リクエストを通して成功/失敗を判定 |

## 実行

```bash
go run main.go
```

## 出力例

```
=== Circuit Breaker Demo ===

📍 Phase 1: 連続失敗させてOpenにする

[Request 1] State: closed
  → Calling http://localhost:19999/api...
  ❌ Error: connection refused

...

[Request 5] State: closed
🔄 [external-api] State changed: closed → open
  ❌ Error: connection refused

📍 Phase 2: Open状態（リクエストは実行されない）

[Request 1] State: open
  ⚡ Rejected: circuit breaker is open   ← APIを呼ばずに即エラー

📍 Phase 3: Timeout待ち（5秒）...

📍 Phase 4: Half-Open → 成功してClosedに戻る
🔄 [external-api] State changed: open → half-open

[Request 1] State: half-open
  → Calling https://httpbin.org/get...
  ✅ Success: 272 bytes

...

🔄 [external-api] State changed: half-open → closed

📍 Final State: closed
```

## 設定パラメータ

```go
gobreaker.Settings{
    Name:        "external-api",
    MaxRequests: 3,                // Half-Open時に許可するリクエスト数
    Interval:    10 * time.Second, // Closed状態でカウントをリセットする間隔
    Timeout:     5 * time.Second,  // Open→Half-Openに移行するまでの時間

    ReadyToTrip: func(counts gobreaker.Counts) bool {
        // Open状態に移行する条件
        return counts.ConsecutiveFailures >= 5
    },

    OnStateChange: func(name string, from, to gobreaker.State) {
        // 状態変化時のコールバック（ログ、メトリクス送信など）
    },
}
```

## なぜ必要か

Circuit Breakerがない場合：

```
サービスA → サービスB（障害中）
    │
    └─→ タイムアウトまで待つ（10秒）
    └─→ リトライ（また10秒待つ）
    └─→ リトライ（また10秒待つ）
    └─→ サービスAもタイムアウト
    └─→ ユーザーは30秒以上待たされる
    └─→ サービスAのスレッドが枯渇
    └─→ 障害が連鎖（カスケード障害）
```

Circuit Breakerがある場合：

```
サービスA → Circuit Breaker → サービスB（障害中）
    │              │
    │              └─→ 5回失敗を検知 → Open状態へ
    │
    └─→ 即座にエラーを返す（待ち時間なし）
    └─→ フォールバック処理へ
    └─→ ユーザーへ即座にレスポンス
    └─→ サービスBの回復を邪魔しない
```

## 実践的な使い方

```go
// サービスごとにCircuit Breakerを作成
var (
    orderServiceCB   = newCircuitBreaker("order-service")
    paymentServiceCB = newCircuitBreaker("payment-service")
)

// API呼び出しをラップ
func callOrderService(ctx context.Context, orderID string) (*Order, error) {
    result, err := orderServiceCB.Execute(func() (*Order, error) {
        return orderClient.GetOrder(ctx, orderID)
    })

    if errors.Is(err, gobreaker.ErrOpenState) {
        // Circuit Breakerがオープン → フォールバック
        return getCachedOrder(orderID), nil
    }

    return result, err
}
```

## 参考

- [sony/gobreaker](https://github.com/sony/gobreaker)
- [Circuit Breaker Pattern - Martin Fowler](https://martinfowler.com/bliki/CircuitBreaker.html)
