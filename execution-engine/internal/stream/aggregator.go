package stream

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/trading-system/execution-engine/internal/exchange"
)

// Aggregator aggregates trade streams from multiple exchanges
type Aggregator struct {
	exchangeManager *exchange.Manager
	streams         map[string]map[string]chan exchange.TradeEvent
	aggregated      map[string]chan exchange.TradeEvent
	mu              sync.RWMutex
}

// NewAggregator creates a new stream aggregator
func NewAggregator(em *exchange.Manager) *Aggregator {
	return &Aggregator{
		exchangeManager: em,
		streams:         make(map[string]map[string]chan exchange.TradeEvent),
		aggregated:      make(map[string]chan exchange.TradeEvent),
	}
}

// Start initializes the aggregator
func (a *Aggregator) Start(ctx context.Context) {
	log.Println("Starting stream aggregator")
	go a.monitorStreams(ctx)
}

// GetStream returns the aggregated trade stream for a symbol
func (a *Aggregator) GetStream(symbol string) (<-chan exchange.TradeEvent, error) {
	a.mu.RLock()
	ch, exists := a.aggregated[symbol]
	a.mu.RUnlock()

	if exists {
		return ch, nil
	}

	return a.createStream(ctx, symbol)
}

func (a *Aggregator) createStream(ctx context.Context, symbol string) (<-chan exchange.TradeEvent, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring lock
	if ch, exists := a.aggregated[symbol]; exists {
		return ch, nil
	}

	// Create new aggregated channel
	aggCh := make(chan exchange.TradeEvent, 1000)
	a.aggregated[symbol] = aggCh

	// Initialize exchange streams for this symbol
	a.streams[symbol] = make(map[string]chan exchange.TradeEvent)

	// Get all exchanges
	exchanges := a.exchangeManager.GetAllExchanges()
	for name, ex := range exchanges {
		exCh, err := ex.StreamTrades(ctx, symbol)
		if err != nil {
			log.Printf("Error creating %s stream for %s: %v", name, symbol, err)
			continue
		}

		a.streams[symbol][name] = exCh
		go a.fanIn(name, symbol, exCh, aggCh)
	}

	return aggCh, nil
}

func (a *Aggregator) fanIn(exchangeName, symbol string, in <-chan exchange.TradeEvent, out chan<- exchange.TradeEvent) {
	for trade := range in {
		// Add exchange name to trade event
		trade.Exchange = exchangeName
		select {
		case out <- trade:
		default:
			log.Printf("Aggregator channel full for %s, dropping trade", symbol)
		}
	}
}

func (a *Aggregator) monitorStreams(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.checkStreamHealth(ctx)
		}
	}
}

func (a *Aggregator) checkStreamHealth(ctx context.Context) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for symbol, exchanges := range a.streams {
		for exchangeName, ch := range exchanges {
			// Check if channel is closed
			if ch == nil {
				log.Printf("Reconnecting %s stream for %s", exchangeName, symbol)
				go a.reconnectStream(ctx, exchangeName, symbol)
			}
		}
	}
}

func (a *Aggregator) reconnectStream(ctx context.Context, exchangeName, symbol string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get exchange
	ex, ok := a.exchangeManager.GetExchange(exchangeName)
	if !ok {
		log.Printf("Exchange %s not found during reconnect", exchangeName)
		return
	}

	// Close old channel if exists
	if oldCh, exists := a.streams[symbol][exchangeName]; exists {
		close(oldCh)
	}

	// Create new stream
	newCh, err := ex.StreamTrades(ctx, symbol)
	if err != nil {
		log.Printf("Error reconnecting %s stream for %s: %v", exchangeName, symbol, err)
		return
	}

	a.streams[symbol][exchangeName] = newCh
	go a.fanIn(exchangeName, symbol, newCh, a.aggregated[symbol])
}