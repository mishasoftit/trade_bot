package order

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/trading-system/execution-engine/internal/exchange"
	"github.com/trading-system/execution-engine/internal/risk"
	"github.com/trading-system/execution-engine/pkg/db"
)

// Order represents a trading order
type Order struct {
	ID         string
	Symbol     string
	Type       Type
	Side       Side
	Price      float64
	Quantity   float64
	Exchange   string
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
	RetryCount int
}

// Type represents order type
type Type int

const (
	Limit  Type = iota // Limit order
	Market             // Market order
)

// Side represents order side
type Side int

const (
	Buy  Side = iota // Buy order
	Sell             // Sell order
)

// Status represents order status
type Status int

const (
	Pending Status = iota
	SentToExchange
	PartiallyFilled
	Filled
	Cancelled
	Failed
	Rejected
)

// Manager handles order processing
type Manager struct {
	exchangeManager *exchange.Manager
	riskClient      *risk.Client
	db             *db.TimescaleDB
	slippageProtection bool
	maxRetries      int
	retryDelay      time.Duration
	orderChan       chan *Order
	mu              sync.Mutex
}

// NewManager creates a new order manager
func NewManager(
	em *exchange.Manager, 
	rc *risk.Client, 
	db *db.TimescaleDB,
) *Manager {
	return &Manager{
		exchangeManager: em,
		riskClient:      rc,
		db:             db,
		maxRetries:      3,
		retryDelay:      500 * time.Millisecond,
		orderChan:       make(chan *Order, 1000),
	}
}

// SetSlippageProtection enables/disables slippage protection
func (m *Manager) SetSlippageProtection(enabled bool) {
	m.slippageProtection = enabled
}

// SubmitOrder submits an order for processing
func (m *Manager) SubmitOrder(o *Order) {
	o.Status = Pending
	o.CreatedAt = time.Now()
	m.orderChan <- o
}

// ProcessOrders processes orders from the queue
func (m *Manager) ProcessOrders(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-m.orderChan:
			go m.processOrder(ctx, o)
		}
	}
}

func (m *Manager) processOrder(ctx context.Context, o *Order) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Apply slippage protection for limit orders
	if m.slippageProtection && o.Type == Limit {
		// Get current market price
		marketPrice, err := m.getMarketPrice(o.Symbol)
		if err != nil {
			o.Status = Failed
			m.logOrder(o, fmt.Sprintf("Failed to get market price: %v", err))
			return
		}

		// Adjust price if necessary
		if o.Side == Buy && o.Price < marketPrice {
			o.Price = marketPrice * 1.005 // Add 0.5% for buy orders
		} else if o.Side == Sell && o.Price > marketPrice {
			o.Price = marketPrice * 0.995 // Reduce 0.5% for sell orders
		}
	}

	// Check with risk controller
	riskApproved, err := m.riskClient.CheckOrder(o)
	if err != nil {
		o.Status = Failed
		m.logOrder(o, fmt.Sprintf("Risk check failed: %v", err))
		return
	}
	if !riskApproved {
		o.Status = Rejected
		m.logOrder(o, "Rejected by risk controller")
		return
	}

	// Execute with retry logic
	for i := 0; i <= m.maxRetries; i++ {
		o.RetryCount = i
		if i > 0 {
			time.Sleep(m.retryDelay * time.Duration(i))
		}

		// Get exchange client
		ex, ok := m.exchangeManager.GetExchange(o.Exchange)
		if !ok {
			o.Status = Failed
			m.logOrder(o, fmt.Sprintf("Exchange not found: %s", o.Exchange))
			return
		}

		// Place order
		exchangeID, err := ex.PlaceOrder(ctx, o)
		if err != nil {
			o.Status = Failed
			m.logOrder(o, fmt.Sprintf("Placement attempt %d failed: %v", i+1, err))
			continue
		}

		o.ID = exchangeID
		o.Status = SentToExchange
		m.logOrder(o, "Order sent to exchange")
		return
	}
}

func (m *Manager) getMarketPrice(symbol string) (float64, error) {
	// In a real implementation, this would fetch from market data
	// For now, return a mock price
	return 100.0, nil
}

func (m *Manager) logOrder(o *Order, message string) {
	o.UpdatedAt = time.Now()
	// Save to database
	if err := m.db.LogOrder(o); err != nil {
		// Handle error
	}
	// Also log to system
	fmt.Printf("[%s] Order %s: %s\n", o.UpdatedAt.Format(time.RFC3339), o.ID, message)
}

// Implement other methods (CancelOrder, GetOrderStatus, etc.)...