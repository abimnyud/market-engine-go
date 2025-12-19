package grpcserver

import (
	"log"
	marketv1 "market-engine-go/gen/go/market/v1"
	marketengine "market-engine-go/internal/market-engine"
	"time"
)

type MarketServer struct {
	marketv1.UnimplementedMarketServiceServer
	Engine *marketengine.MarketEngine
}

func (server *MarketServer) StreamTrades(req *marketv1.StreamTradesRequest, stream marketv1.MarketService_StreamTradesServer) error {
	intervalMs := req.GetIntervalMs()
	if intervalMs <= 0 {
		intervalMs = 500
	}

	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	log.Printf("Client connected: streaming every %v", req.GetIntervalMs())

	for {
		select {
		case <-stream.Context().Done():
			log.Println("Client disconnected")
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
				log.Printf("Send failed: %v", err)
				return err // Connection closed
			}
		case <-stream.Context().Done():
			log.Println("Client disconnected")
			return stream.Context().Err()
		}
	}
}
