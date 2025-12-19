package models

import "time"

type Order struct {
	Price     float64 `json:"price"`
	Volume    int     `json:"volume"`
	Frequency int     `json:"frequency"`
}

type Trade struct {
	ID        string    `json:"id"`
	Ticker    string    `json:"ticker"`
	Price     float64   `json:"price"`
	Size      int       `json:"size"`
	Side      string    `json:"side"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderBook struct {
	Bids []Order `json:"bids"`
	Asks []Order `json:"asks"`
}
