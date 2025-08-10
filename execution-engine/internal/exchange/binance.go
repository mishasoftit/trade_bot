package exchange

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

// BinanceClient implements the exchange interface for Binance
type BinanceClient struct {
	client      *futures.Client
	streams     map[string]chan TradeEvent
	streamMutex sync.Mutex
	connected   bool
}

// NewBinanceClient creates a new Binance client
func NewBinanceClient() *BinanceClient {
	return &BinanceClient{
		client:  futures.NewClient("", ""), // API keys will be set via config
		streams: make(map[string]chan TradeEvent),
	}
}

// Connect establishes connection to Binance
func (b *BinanceClient) Connect() error {
	// Test connectivity
	_, err := b.client.NewPingService().Do(context.Background())
	if err != nil {
		return err
	}
	b.connected = true
	log.Println("Connected to Binance")
	return nil
}

// Disconnect closes all connections
func (b *BinanceClient) Disconnect() error {
	b.streamMutex.Lock()
	defer b.streamMutex.Unlock()
	
	for symbol := range b.streams {
		close(b.streams[symbol])
		delete(b.streams, symbol)
	}
	b.connected = false
	return nil
}

// PlaceOrder places an order on Binance
func (b *BinanceClient) PlaceOrder(ctx context.Context, o *order.Order) (string, error) {
	if !b.connected {
		return "", ErrNotConnected
	}

	orderType := binance.OrderTypeLimit
	if o.Type == order.Market {
		orderType = binance.OrderTypeMarket
	}

	side := binance.SideTypeBuy
	if o.Side == order.Sell {
		side = binance.SideTypeSell
	}

	res, err := b.client.NewCreateOrderService().
		Symbol(o.Symbol).
		Side(side).
		Type(orderType).
		TimeInForce(binance.TimeInForceTypeGTC).
		Quantity(o.QuantityString()).
		Price(o.PriceString()).
		Do(ctx)

	if err != nil {
		return "", err
	}
	return res.OrderID, nil
}

// StreamTrades opens a real-time trade stream for a symbol
func (b *BinanceClient) StreamTrades(ctx context.Context, symbol string) (chan TradeEvent, error) {
	if !b.connected {
		return nil, ErrNotConnected
	}

	b.streamMutex.Lock()
	defer b.streamMutex.Unlock()

	if ch, exists := b.streams[symbol]; exists {
		return ch, nil
	}

	ch := make(chan TradeEvent, 100)
	b.streams[symbol] = ch

	wsHandler := func(event *futures.WsAggTradeEvent) {
		price, _ := event.Price.Float64()
		qty, _ := event.Quantity.Float64()
		ch <- TradeEvent{
			Symbol:    event.Symbol,
			Price:     price,
			Quantity:  qty,
			Timestamp: time.Unix(0, event.Time*int64(time.Millisecond)),
		}
	}

	errHandler := func(err error) {
		log.Printf("Binance stream error: %v", err)
	}

	_, done, err := futures.WsAggTradeServe(symbol, wsHandler, errHandler)
	if err != nil {
		return nil, err
	}

	go func() {
		<-done
		b.streamMutex.Lock()
		defer b.streamMutex.Unlock()
		if ch, ok := b.streams[symbol]; ok {
			close(ch)
			delete(b.streams, symbol)
		}
	}()

	return ch, nil
}

// Implement other required methods (CancelOrder, GetOrderStatus, GetBalance)...