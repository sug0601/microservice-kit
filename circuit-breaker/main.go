package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Circuit Breakerの3つの状態:
// - Closed: 正常。リクエストを通す
// - Open: 障害検知。リクエストを即座に失敗させる
// - Half-Open: 回復確認中。一部のリクエストを通して様子見

var cb *gobreaker.CircuitBreaker[[]byte]

func init() {
	cb = gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:        "external-api",
		MaxRequests: 3,                // Half-Open時に許可するリクエスト数
		Interval:    10 * time.Second, // Closed状態でカウントをリセットする間隔
		Timeout:     5 * time.Second,  // Open→Half-Openに移行するまでの時間

		// Open状態に移行する条件
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 連続5回失敗 または 失敗率50%以上（最低10リクエスト）
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.ConsecutiveFailures >= 5 ||
				(counts.Requests >= 10 && failureRatio >= 0.5)
		},

		// 状態変化時のコールバック
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("🔄 [%s] State changed: %s → %s\n", name, from, to)
		},

		// Half-Open→Closedに戻る条件（デフォルト: MaxRequests分成功したら）
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	})
}

// Circuit Breaker経由でAPIを呼ぶ
func callAPI(url string) ([]byte, error) {
	return cb.Execute(func() ([]byte, error) {
		fmt.Printf("  → Calling %s...\n", url)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("server error: %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	})
}

func main() {
	fmt.Println("=== Circuit Breaker Demo ===")
	fmt.Println()

	// 存在しないURL（失敗する）
	badURL := "http://localhost:19999/api"
	// 成功するURL
	goodURL := "https://httpbin.org/get"

	// 1. 連続失敗でOpenになる様子
	fmt.Println("📍 Phase 1: 連続失敗させてOpenにする")
	for i := 1; i <= 7; i++ {
		fmt.Printf("\n[Request %d] State: %s\n", i, cb.State())
		_, err := callAPI(badURL)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 2. Open状態では即座にエラー
	fmt.Println("\n📍 Phase 2: Open状態（リクエストは実行されない）")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n[Request %d] State: %s\n", i, cb.State())
		_, err := callAPI(badURL)
		if err != nil {
			fmt.Printf("  ⚡ Rejected: %v\n", err)
		}
	}

	// 3. Timeout後にHalf-Openへ
	fmt.Println("\n📍 Phase 3: Timeout待ち（5秒）...")
	time.Sleep(6 * time.Second)

	// 4. Half-Open状態で成功するURLを呼ぶ
	fmt.Println("\n📍 Phase 4: Half-Open → 成功してClosedに戻る")
	for i := 1; i <= 5; i++ {
		fmt.Printf("\n[Request %d] State: %s\n", i, cb.State())
		body, err := callAPI(goodURL)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
		} else {
			fmt.Printf("  ✅ Success: %d bytes\n", len(body))
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n📍 Final State: %s\n", cb.State())
	fmt.Println("\n=== Demo Complete ===")
}
