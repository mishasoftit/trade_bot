package reconciliation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/trading-system/execution-engine/internal/exchange"
	"github.com/trading-system/execution-engine/internal/order"
	"github.com/trading-system/execution-engine/pkg/db"
)

// Reconciler handles trade reconciliation
type Reconciler struct {
	exchangeManager *exchange.Manager
	db             *db.TimescaleDB
}

// NewReconciler creates a new reconciliation system
func NewReconciler(em *exchange.Manager, db *db.TimescaleDB) *Reconciler {
	return &Reconciler{
		exchangeManager: em,
		db:             db,
	}
}

// Run starts the reconciliation process at regular intervals
func (r *Reconciler) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.Reconcile(ctx); err != nil {
				log.Printf("Reconciliation failed: %v", err)
			}
		}
	}
}

// Reconcile checks for discrepancies between our records and exchange records
func (r *Reconciler) Reconcile(ctx context.Context) error {
	log.Println("Starting reconciliation process")

	// Get all orders that need reconciliation
	orders, err := r.db.GetOrdersForReconciliation(ctx)
	if err != nil {
		return fmt.Errorf("error fetching orders: %w", err)
	}

	for _, o := range orders {
		// Get exchange client
		ex, ok := r.exchangeManager.GetExchange(o.Exchange)
		if !ok {
			log.Printf("Exchange %s not found for order %s", o.Exchange, o.ID)
			continue
		}

		// Get order status from exchange
		status, err := ex.GetOrderStatus(o.ID)
		if err != nil {
			log.Printf("Error getting status for order %s: %v", o.ID, err)
			continue
		}

		// Compare with our records
		if o.Status != status {
			log.Printf("Discrepancy found for order %s: our status=%d, exchange status=%d", 
				o.ID, o.Status, status)
			
			// Update our records
			o.Status = status
			if err := r.db.UpdateOrderStatus(ctx, o.ID, status); err != nil {
				log.Printf("Failed to update order %s: %v", o.ID, err)
			}

			// Additional handling based on status change
			switch status {
			case order.Filled:
				r.handleFilledOrder(ctx, o)
			case order.Cancelled:
				r.handleCancelledOrder(ctx, o)
			}
		}
	}

	log.Println("Reconciliation completed")
	return nil
}

func (r *Reconciler) handleFilledOrder(ctx context.Context, o *order.Order) {
	// Fetch trade details from exchange
	trades, err := r.exchangeManager.GetTradeDetails(o.Exchange, o.ID)
	if err != nil {
		log.Printf("Error fetching trade details for order %s: %v", o.ID, err)
		return
	}

	// Save trade details to database
	for _, t := range trades {
		if err := r.db.LogTrade(ctx, t); err != nil {
			log.Printf("Error logging trade for order %s: %v", o.ID, err)
		}
	}

	// Notify other systems (risk controller, etc.)
	log.Printf("Order %s filled on %s", o.ID, o.Exchange)
}

func (r *Reconciler) handleCancelledOrder(ctx context.Context, o *order.Order) {
	// Update order status in database already handled
	log.Printf("Order %s cancelled on %s", o.ID, o.Exchange)
	
	// Notify strategy service or other components
	// This could be implemented with a message queue or HTTP call
}