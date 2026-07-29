package market

import (
	"sync"

	"github.com/shopspring/decimal"
)

// Engine is the core order book + matching engine.
// Phase 1 (off-chain): all matching happens here in-memory.
// Phase 2 (on-chain): matching still here, settlement moves to TON contracts.
type Engine struct {
	mu       sync.RWMutex
	buyBook  []Order // sorted by price desc
	sellBook []Order // sorted by price asc
	trades   []Trade
	halted   bool // circuit breaker
}

type Order struct {
	ID       string
	UserID   string
	ItemType string // "pet" / "skill"
	ItemID   string
	Side     string // "buy" / "sell"
	Price    decimal.Decimal
	Currency string
	IsMM     bool // market maker order
}

type Trade struct {
	ID       string
	ItemType string
	ItemID   string
	SellerID string
	BuyerID  string
	Price    decimal.Decimal
	Currency string
	Fee      decimal.Decimal
	IsMM     bool
}

func NewEngine() *Engine {
	return &Engine{}
}

// SubmitOrder adds an order and attempts matching.
func (e *Engine) SubmitOrder(order Order) ([]Trade, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.halted {
		return nil, ErrMarketHalted
	}

	// TODO: Match against opposite book
	// TODO: If no match, add to book
	// TODO: Calculate fee (3-5%)
	// TODO: Record trade

	return nil, nil
}

// CancelOrder removes an open order from the book.
func (e *Engine) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// TODO: Find and remove order from buy/sell book
	return nil
}

// Halt pauses all trading (circuit breaker).
func (e *Engine) Halt() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.halted = true
}

// Resume re-enables trading.
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.halted = false
}

// GetOrderBook returns current bid/ask for display.
func (e *Engine) GetOrderBook(itemType string, limit int) (bids []Order, asks []Order) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// TODO: Filter by itemType, return top N
	return nil, nil
}

var ErrMarketHalted = &MarketError{"market is halted (circuit breaker)"}

type MarketError struct{ msg string }

func (e *MarketError) Error() string { return e.msg }
