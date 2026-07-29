package market

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

// MarketMaker simulates trading activity with multiple bot personas.
// Used in Phase 0-1 (off-chain) to create price movement and volume.
type MarketMaker struct {
	engine      *Engine
	bots        []MMBot
	intervalMin time.Duration
	intervalMax time.Duration
	stopCh      chan struct{}
}

type MMBot struct {
	ID       string
	Name     string
	Persona  string // aggressive_buyer / cautious_seller / hoarder / swing
	Balance  decimal.Decimal
	IsActive bool
}

type MMConfig struct {
	BotCount     int
	IntervalMin  time.Duration
	IntervalMax  time.Duration
	InitialPrice decimal.Decimal
}

func NewMarketMaker(engine *Engine, cfg MMConfig) *MarketMaker {
	mm := &MarketMaker{
		engine:      engine,
		intervalMin: cfg.IntervalMin,
		intervalMax: cfg.IntervalMax,
		stopCh:      make(chan struct{}),
	}

	personas := []string{"aggressive_buyer", "cautious_seller", "hoarder", "swing"}
	names := []string{"CryptoWhale88", "TONdiamond", "MechaFan2026", "PixelTrader", "AlphaHunter"}

	for i := 0; i < cfg.BotCount; i++ {
		mm.bots = append(mm.bots, MMBot{
			ID:       fmt.Sprintf("mm_%d", i),
			Name:     names[i%len(names)],
			Persona:  personas[i%len(personas)],
			Balance:  decimal.NewFromInt(1000),
			IsActive: true,
		})
	}

	return mm
}

// Start begins the market making loop.
func (mm *MarketMaker) Start() {
	go mm.loop()
}

// Stop halts the market maker.
func (mm *MarketMaker) Stop() {
	close(mm.stopCh)
}

func (mm *MarketMaker) loop() {
	for {
		select {
		case <-mm.stopCh:
			return
		default:
			interval := mm.randomInterval()
			time.Sleep(interval)
			mm.executeOne()
		}
	}
}

func (mm *MarketMaker) randomInterval() time.Duration {
	diff := mm.intervalMax - mm.intervalMin
	return mm.intervalMin + time.Duration(rand.Int63n(int64(diff)))
}

func (mm *MarketMaker) executeOne() {
	// Pick random active bot
	bot := mm.bots[rand.Intn(len(mm.bots))]
	if !bot.IsActive {
		return
	}

	// Decide action based on persona
	action := mm.decideAction(bot)
	if action == "hold" {
		return
	}

	// TODO: Generate order with random price around current market
	// TODO: Submit to engine
	// TODO: Occasionally generate "big order" for announcements
}

func (mm *MarketMaker) decideAction(bot MMBot) string {
	switch bot.Persona {
	case "aggressive_buyer":
		if rand.Float64() < 0.7 {
			return "buy"
		}
	case "cautious_seller":
		if rand.Float64() < 0.3 {
			return "sell"
		}
	case "hoarder":
		if rand.Float64() < 0.8 {
			return "buy"
		}
	case "swing":
		if rand.Float64() < 0.5 {
			return "buy"
		}
		return "sell"
	}
	return "hold"
}
