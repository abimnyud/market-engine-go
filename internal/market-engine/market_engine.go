package marketengine

import (
	"fmt"
	"market-engine-go/internal/models"
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

type MarketEngine struct {
	orderBooks   map[string]*models.OrderBook
	Trades       []models.Trade
	Mu           sync.RWMutex
	Tickers      []string
	TradeChannel chan models.Trade
}

var tickers = []string{"BBCA", "BBRI", "GOTO", "TLKM", "ASII"}
var initialPrices = map[string]float64{
	"BBCA": 8150,
	"BBRI": 3800,
	"GOTO": 65,
	"TLKM": 3400,
	"ASII": 6450,
}

func New() *MarketEngine {
	engine := &MarketEngine{
		Trades:       make([]models.Trade, 0, 1000),
		TradeChannel: make(chan models.Trade, 100),
	}

	return engine
}

func (engine *MarketEngine) StartSimulation() {
	tickerChannel := time.NewTicker(5 * time.Millisecond)

	go func() {
		for range tickerChannel.C {
			engine.updateTrades()
		}
	}()
}

func (engine *MarketEngine) updateTrades() {
	engine.Mu.Lock()
	defer engine.Mu.Unlock()

	symbol := tickers[rand.IntN(len(tickers))]
	basePrice := initialPrices[symbol]

	// Randomize price slightly (+/- 0.5%)
	change := math.Floor((rand.Float64() - 0.5) * (basePrice * 0.01))
	tradePrice := basePrice + change

	newTrade := models.Trade{
		ID:        fmt.Sprintf("TRD-%d", time.Now().UnixNano()),
		Ticker:    symbol,
		Price:     tradePrice,
		Size:      rand.IntN(1000) + 1,
		Side:      []string{"BUY", "SELL"}[rand.IntN(2)],
		Timestamp: time.Now(),
	}

	runningTrades := engine.Trades
	if len(runningTrades) >= 1000 {
		engine.Trades = runningTrades[1:]
	}

	engine.Trades = append(runningTrades, newTrade)

	select {
	case engine.TradeChannel <- newTrade:
	default:
	}
}
