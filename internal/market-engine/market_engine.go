package marketengine

import (
	"context"
	"fmt"
	"maps"
	"market-engine-go/internal/models"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	marketv1 "market-engine-go/gen/go/market/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type MarketEngine struct {
	orderBooks    map[string]*models.OrderBook
	Trades        []models.Trade
	Mu            sync.RWMutex
	Tickers       map[string]*marketv1.TickerData
	TradeChannel  chan models.Trade
	CurrentPrices map[string]int64
}

func New() *MarketEngine {
	engine := &MarketEngine{
		Trades:        make([]models.Trade, 0, 1000),
		TradeChannel:  make(chan models.Trade, 100),
		CurrentPrices: make(map[string]int64),
		Tickers: map[string]*marketv1.TickerData{
			"BBCA": {Symbol: "BBCA", Price: 8150, Name: "Bank Central Asia Tbk"},
			"BBRI": {Symbol: "BBRI", Price: 3800, Name: "Bank Rakyat Indonesia (Persero) Tbk"},
			"GOTO": {Symbol: "GOTO", Price: 65, Name: "Goto Gojek Tokopedia Tbk"},
			"TLKM": {Symbol: "TLKM", Price: 3400, Name: "Telkom Indonesia Tbk"},
			"ASII": {Symbol: "ASII", Price: 6450, Name: "Astra International Tbk"},
			"SUPA": {Symbol: "SUPA", Price: 1230, Name: "Superbank Indonesia Tbk"},
			"BMRI": {Symbol: "BMRI", Price: 5175, Name: "Bank Mandiri (Persero) Tbk"},
			"ADRO": {Symbol: "ADRO", Price: 1900, Name: "Alamtri Resources Indonesia Tbk"},
			"ANTM": {Symbol: "ANTM", Price: 3070, Name: "Aneka Tambang Tbk"},
			"UNVR": {Symbol: "UNVR", Price: 2770, Name: "Unilever Indonesia Tbk"},
			"INDF": {Symbol: "INDF", Price: 6750, Name: "Indofood Sukses Makmur Tbk"},
			"ICBP": {Symbol: "ICBP", Price: 8425, Name: "Indofood CBP Sukses Makmur Tbk"},
			"PTBA": {Symbol: "PTBA", Price: 2270, Name: "Bukit Asam Tbk"},
			"BBNI": {Symbol: "BBNI", Price: 4340, Name: "Bank Negara Indonesia (Persero) Tbk"},
			"ITMG": {Symbol: "ITMG", Price: 21575, Name: "Indo Tambangraya Megah Tbk"},
			"KLBF": {Symbol: "KLBF", Price: 1200, Name: "Kalbe Farma Tbk"},
			"UNTR": {Symbol: "UNTR", Price: 29800, Name: "United Tractors Tbk"},
			"MDKA": {Symbol: "MDKA", Price: 2190, Name: "Merdeka Copper Gold Tbk"},
			"AADI": {Symbol: "AADI", Price: 7050, Name: "Adaro Andalan Indonesia Tbk"},
			"ISAT": {Symbol: "ISAT", Price: 2430, Name: "Indosat Tbk"},
			"BRPT": {Symbol: "BRPT", Price: 3510, Name: "Barito Pacific Tbk"},
		},
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

	symbols := slices.Collect(maps.Keys(engine.Tickers))
	symbol := symbols[rand.IntN(len(symbols))]
	basePrice := engine.Tickers[symbol].Price

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

func (engine *MarketEngine) RunPriceGenerator(ctx context.Context, symbol string, channel chan<- *marketv1.StreamTickersResponse) {
	for {
		interval := time.Duration(100+rand.IntN(2000-100)) * time.Millisecond

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			updated := engine.calculateNextPrice(symbol)

			if updated != nil {
				channel <- updated
			}
		}
	}
}

func (engine *MarketEngine) calculateNextPrice(symbol string) *marketv1.StreamTickersResponse {
	lastPrice, exists := engine.CurrentPrices[symbol]
	if !exists {
		lastPrice = int64(engine.Tickers[symbol].Price)
	}

	// TODO: Make it using the fraction system
	changeAmount := int32((rand.IntN(10) - 5) * 1)

	newPrice := lastPrice + int64(changeAmount)

	if newPrice < 50 {
		return nil
	}

	engine.CurrentPrices[symbol] = newPrice

	updated := &marketv1.StreamTickersResponse{
		Symbol:    symbol,
		Price:     float64(newPrice),
		Change:    wrapperspb.Int32(changeAmount),
		Timestamp: time.Now().UnixMilli(),
	}

	return updated
}
