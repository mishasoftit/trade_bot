package exchange

import (
	"context"
	"time"

	"github.com/trading-system/execution-engine/internal/order"
)

// Interface defines the contract for exchange implementations
type Interface interface {
	Connect() error
	Disconnect() error
	PlaceOrder(ctx context.Context, o *order.Order) (string, error)
	CancelOrder(orderID string) error
	GetOrderStatus(orderID string) (order.Status, error)
	StreamTrades(ctx context.Context, symbol string) (chan TradeEvent, error)
	GetBalance(currency string) (float64, error)
}

// TradeEvent represents a real-time trade event
type TradeEvent struct {
	Symbol    string
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

// Manager handles multiple exchange connections
type Manager struct {
	exchanges map[string]Interface
}

// NewManager creates a new exchange manager
func NewManager(exchanges map[string]Interface) *Manager {
	return &Manager{exchanges: exchanges}
}

// GetExchange returns an exchange by name
func (m *Manager) GetExchange(name string) (Interface, bool) {
	ex, ok := m.exchanges[name]
	return ex, ok
}

// PlaceOrder places an order on the specified exchange
func (m *Manager) PlaceOrder(ctx context.Context, exchangeName string, o *order.Order) (string, error) {
	ex, ok := m.GetExchange(exchangeName)
	if !ok {
		return "", ErrExchangeNotFound
	}
	return ex.PlaceOrder(ctx, o)
}

// Common errors
var (
	ErrExchangeNotFound = errors.New("exchange not found")
	ErrNotConnected     = errors.New("exchange not connected")
)