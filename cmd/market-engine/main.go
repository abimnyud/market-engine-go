package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	marketv1 "market-engine-go/gen/go/market/v1"
	grpcserver "market-engine-go/internal/grpc"
	marketengine "market-engine-go/internal/market-engine"
)

func main() {
	engine := marketengine.New()

	log.Println("Market Engine Simulation Starting...")
	engine.StartSimulation()

	log.Println("Press Ctrl+C to stop")

	port := ":50051"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()

	marketv1.RegisterMarketServiceServer(server, &grpcserver.MarketServer{Engine: engine})
	reflection.Register(server)

	log.Printf("gRPC Server listening on %s", port)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
