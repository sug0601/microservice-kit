package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ksqldb-demo/kafka"
	"github.com/example/ksqldb-demo/ksqldb"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	ksqldbURL := getEnv("KSQLDB_URL", "http://localhost:8088")

	switch os.Args[1] {
	case "produce":
		runProducer(kafkaBroker)
	case "consume":
		runConsumer(kafkaBroker)
	case "query":
		runQuery(ksqldbURL)
	case "stream":
		runStreamQuery(ksqldbURL)
	case "demo":
		runDemo(kafkaBroker, ksqldbURL)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ksqldb-demo <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  produce  - Produce sample orders to Kafka")
	fmt.Println("  consume  - Consume orders from Kafka")
	fmt.Println("  query    - Query aggregated data from ksqlDB")
	fmt.Println("  stream   - Subscribe to ksqlDB push query")
	fmt.Println("  demo     - Run full demo (produce + query)")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runProducer(broker string) {
	producer, err := kafka.NewProducer(broker, "orders")
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	orders := []kafka.Order{
		{OrderID: "O101", CustomerID: "C001", Product: "MacBook Pro", Quantity: 1, Price: 2499.00},
		{OrderID: "O102", CustomerID: "C002", Product: "iPhone 15", Quantity: 2, Price: 999.00},
		{OrderID: "O103", CustomerID: "C001", Product: "AirPods Pro", Quantity: 1, Price: 249.00},
		{OrderID: "O104", CustomerID: "C003", Product: "iPad Air", Quantity: 1, Price: 799.00},
		{OrderID: "O105", CustomerID: "C002", Product: "Apple Watch", Quantity: 1, Price: 399.00},
	}

	for _, order := range orders {
		if err := producer.SendOrder(order); err != nil {
			log.Printf("Failed to send order %s: %v", order.OrderID, err)
		} else {
			log.Printf("Sent: %s - %s x%d ($%.2f)", order.OrderID, order.Product, order.Quantity, order.Price)
		}
		time.Sleep(500 * time.Millisecond)
	}

	producer.Flush()
	log.Println("All orders sent!")
}

func runConsumer(broker string) {
	consumer, err := kafka.NewConsumer(broker, "orders", "demo-group")
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	log.Println("Consuming orders (Ctrl+C to stop)...")
	if err := consumer.Consume(ctx, func(order kafka.Order) {
		log.Printf("Received: %s - %s (Customer: %s, Total: $%.2f)",
			order.OrderID, order.Product, order.CustomerID,
			float64(order.Quantity)*order.Price)
	}); err != nil {
		log.Printf("Consumer stopped: %v", err)
	}
}

func runQuery(ksqldbURL string) {
	client := ksqldb.NewClient(ksqldbURL)

	log.Println("Querying ORDER_TOTALS from ksqlDB...")
	results, err := client.PullQuery("SELECT * FROM ORDER_TOTALS;")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Println()
	fmt.Println("Customer Order Totals:")
	fmt.Println("─────────────────────────────────────")
	for _, row := range results {
		fmt.Printf("  Customer: %-6s | Orders: %v | Total: $%v\n",
			row["CUSTOMER_ID"], row["ORDER_COUNT"], row["TOTAL_AMOUNT"])
	}
	fmt.Println("─────────────────────────────────────")
}

func runStreamQuery(ksqldbURL string) {
	client := ksqldb.NewClient(ksqldbURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	log.Println("Subscribing to HIGH_VALUE_ORDERS stream (Ctrl+C to stop)...")
	err := client.PushQuery(ctx, "SELECT * FROM HIGH_VALUE_ORDERS EMIT CHANGES;", func(row map[string]interface{}) {
		log.Printf("High-value order: %s - %s (Customer: %s, Qty: %v, Price: $%v)",
			row["ORDER_ID"], row["PRODUCT"], row["CUSTOMER_ID"],
			row["QUANTITY"], row["PRICE"])
	})
	if err != nil {
		log.Printf("Stream ended: %v", err)
	}
}

func runDemo(broker, ksqldbURL string) {
	log.Println("=== ksqlDB Stream Processing Demo ===")
	log.Println()

	// Produce orders
	log.Println("Step 1: Sending orders to Kafka...")
	runProducer(broker)

	// Wait for processing
	log.Println()
	log.Println("Step 2: Waiting for ksqlDB to process...")
	time.Sleep(2 * time.Second)

	// Query results
	log.Println()
	log.Println("Step 3: Querying aggregated results...")
	runQuery(ksqldbURL)

	log.Println()
	log.Println("Demo complete!")
}
