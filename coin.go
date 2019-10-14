package asset

import (
	"time"
)

// CoinBase provides functionality to store series of coins.
type CoinBase interface {
	Store(series Series) error
}

// Series is a custom type to enable sorting on a Coin slice.
type Series []Coin

// Coin represents a crypto currency asset.
type Coin struct {
	Name          string
	Symbol        string
	Timestamp     time.Time
	PriceInDollar float64
}

func (p Series) Len() int {
	return len(p)
}

func (p Series) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Series) Less(i, j int) bool {
	return p[i].Timestamp.Before(p[j].Timestamp)
}
