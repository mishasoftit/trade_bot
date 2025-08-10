package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trading-system/execution-engine/internal/exchange"
	"github.com/trading-system/execution-engine/internal/order"
	"github.com/trading-system/execution-engine/internal/reconciliation"
	"github.com/trading-system/execution-engine/internal/risk"
	"github.com/trading-system/execution-engine/internal/stream"
	"github.com/trading-system/execution-engine/pkg/db"
	"github.com/trading-system/execution-engine/pkg/metrics"
)

func main() {
	// Initialize configuration
	// TODO: Load from config file/env vars

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize TimescaleDB connection
	dbConn, err := db.ConnectTimescaleDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to TimescaleDB: %v", err)
	}
	defer dbConn.Close(ctx)

	// Initialize Prometheus metrics
	metricsServer := metrics.NewServer(":9090")
	go metricsServer.Start()

	// Initialize risk controller client
	riskClient := risk.NewClient("http://localhost:8080") // Update with actual risk controller URL

	// Initialize exchanges
	binance := exchange.NewBinanceClient()
	kraken := exchange.NewKrakenClient()
	coinbase := exchange.NewCoinbaseClient()

	// Create exchange manager
	exchangeManager := exchange.NewManager(map[string]exchange.Interface{
		"binance": binance,
		"kraken":  kraken,
		"coinbase": coinbase,
	})

	// Initialize WebSocket stream aggregator
	streamAggregator := stream.NewAggregator(exchangeManager)
	go streamAggregator.Start(ctx)

	// Initialize order manager with anti-slippage
	orderManager := order.NewManager(exchangeManager, riskClient, dbConn)
	orderManager.SetSlippageProtection(true)

	// Initialize reconciliation system
	reconciler := reconciliation.NewReconciler(exchangeManager, dbConn)
	go reconciler.Run(ctx, 5*time.Minute) // Reconcile every 5 minutes

	// Start order processing
	go orderManager.ProcessOrders(ctx)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down execution engine...")
	cancel()
	time.Sleep(2 * time.Second) // Allow goroutines to clean up
}