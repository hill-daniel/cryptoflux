package asset

// ResponseObject represents data from a ticker.
type ResponseObject struct {
	Data []CmcQuote `json:"data"`
}

// CmcQuote represents data from a ticker.
type CmcQuote struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Quote  Quote  `json:"quote"`
}

// Quote represents data from a ticker.
type Quote struct {
	Usd Price `json:"USD"`
}

// Price represents the price in USD.
type Price struct {
	Price float64 `json:"price"`
}
