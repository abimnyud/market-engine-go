package grpcserver

import (
	"context"
	"log"
	marketv1 "market-engine-go/gen/go/market/v1"
	marketengine "market-engine-go/internal/market-engine"
	"time"
)

type MarketServer struct {
	marketv1.UnimplementedMarketServiceServer
	Engine *marketengine.MarketEngine
}

func (server *MarketServer) GetTickers(ctx context.Context, req *marketv1.GetTickersRequest) (*marketv1.GetTickersResponse, error) {
	var res []*marketv1.TickerData
	for _, v := range server.Engine.Tickers {
		res = append(res, v)
	}

	return &marketv1.GetTickersResponse{Tickers: res}, nil
}

func (server *MarketServer) StreamTickers(stream marketv1.MarketService_StreamTickersServer) error {
	updateChannel := make(chan *marketv1.StreamTickersResponse, 100)
	activeGenerators := make(map[string]context.CancelFunc)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	log.Printf("[StreamTickers] Client connected")

	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				log.Println("[StreamTickers] Client closed connection")
				return
			}

			log.Printf("[StreamTickers] Processing %v", req.Symbols)
			newSymbols := make(map[string]bool)
			for _, s := range req.Symbols {
				newSymbols[s] = true
			}

			var unsubscribedSymbols []string
			for symbol, stop := range activeGenerators {
				if !newSymbols[symbol] {
					unsubscribedSymbols = append(unsubscribedSymbols, symbol)
					stop()
					delete(activeGenerators, symbol)
				}
			}
			if len(unsubscribedSymbols) > 0 {
				log.Printf("[StreamTickers] Unsubscribing: %v", unsubscribedSymbols)
			}

			for _, symbol := range req.Symbols {
				if _, exists := activeGenerators[symbol]; !exists {
					if _, ok := server.Engine.Tickers[symbol]; !ok {
						log.Printf("[StreamTickers] Ticker unavailable: %v", symbol)
						continue
					}

					tickerCtx, tickerCancel := context.WithCancel(ctx)
					activeGenerators[symbol] = tickerCancel
					go server.Engine.RunPriceGenerator(tickerCtx, symbol, updateChannel)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("[StreamTickers] Client disconnected")

			return nil
		case update := <-updateChannel:
			if err := stream.Send(update); err != nil {
				log.Printf("[StreamTickers] Send failed: %v", err)
				return err
			}
		}
	}
}

func (server *MarketServer) StreamTrades(req *marketv1.StreamTradesRequest, stream marketv1.MarketService_StreamTradesServer) error {
	intervalMs := req.GetIntervalMs()
	if intervalMs <= 0 {
		intervalMs = 100
	}

	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	log.Printf("[StreamTrades] Client connected: streaming every %v", req.GetIntervalMs())

	for {
		select {
		case <-stream.Context().Done():
			log.Println("[StreamTrades] Client disconnected")
			return stream.Context().Err()
		case <-ticker.C:
			server.Engine.Mu.RLock()
			if len(server.Engine.Trades) == 0 {
				server.Engine.Mu.RUnlock()
				continue
			}

			latest := server.Engine.Trades[len(server.Engine.Trades)-1]
			server.Engine.Mu.RUnlock()

			const timeFormatRFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
			err := stream.Send(&marketv1.StreamTradesResponse{
				Id:        latest.ID,
				Ticker:    latest.Ticker,
				Price:     latest.Price,
				Size:      int32(latest.Size),
				Side:      latest.Side,
				Timestamp: latest.Timestamp.Format(timeFormatRFC3339Milli),
			})

			if err != nil {
				log.Printf("[StreamTrades] Send failed: %v", err)
				return err
			}
		}
	}
}
