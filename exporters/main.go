package main

import (
	"net/http"
	"time"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	exchangeLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_api_latency_seconds",
			Help: "Latency of exchange API calls in seconds",
		},
		[]string{"exchange"},
	)

	orderFillRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "order_fill_rate",
			Help: "Percentage of orders filled successfully",
		},
		[]string{"exchange", "symbol"},
	)

	riskCalcTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "risk_calculation_time_seconds",
			Help: "Time taken for risk calculations in seconds",
		},
		[]string{"service"},
	)

	circuitBreakerStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_triggered",
			Help: "Circuit breaker status (1 = triggered, 0 = normal)",
		},
		[]string{"service"},
	)
)

func init() {
	prometheus.MustRegister(exchangeLatency)
	prometheus.MustRegister(orderFillRate)
	prometheus.MustRegister(riskCalcTime)
	prometheus.MustRegister(circuitBreakerStatus)
}

func main() {
	// Start background tasks to collect metrics
	go collectExchangeLatency()
	go collectFillRates()
	go collectRiskMetrics()
	go collectCircuitBreakers()

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9100", nil)
}

func collectExchangeLatency() {
	for {
		// In production, this would call exchange APIs and measure latency
		// For now, we'll simulate with random values
		exchangeLatency.WithLabelValues("binance").Set(0.2)
		exchangeLatency.WithLabelValues("kraken").Set(0.15)
		exchangeLatency.WithLabelValues("coinbase").Set(0.25)
		time.Sleep(5 * time.Second)
	}
}

func collectFillRates() {
	for {
		// In production, this would query order databases
		// For now, we'll simulate with random values
		orderFillRate.WithLabelValues("binance", "BTCUSDT").Set(0.92)
		orderFillRate.WithLabelValues("kraken", "ETHUSD").Set(0.85)
		orderFillRate.WithLabelValues("coinbase", "SOLUSD").Set(0.78)
		time.Sleep(10 * time.Second)
	}
}

func collectRiskMetrics() {
	for {
		// In production, this would query risk services
		riskCalcTime.WithLabelValues("strategy-service").Set(1.2)
		riskCalcTime.WithLabelValues("risk-controller").Set(0.8)
		time.Sleep(3 * time.Second)
	}
}

func collectCircuitBreakers() {
	for {
		// In production, this would query circuit breaker services
		circuitBreakerStatus.WithLabelValues("strategy-service").Set(0)
		circuitBreakerStatus.WithLabelValues("risk-controller").Set(0)
		time.Sleep(2 * time.Second)
	}
}