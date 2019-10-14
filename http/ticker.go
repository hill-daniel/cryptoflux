package http

import (
	"encoding/json"
	"fmt"
	"github.com/hill-daniel/cryptoflux"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	envTickerEndpoint = "TICKER_ENDPOINT"
	envTickerAPIKey   = "TICKER_API_KEY"
)

// Client is just a http get abstraction.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Clock is an abstraction for receiving time.
type Clock func() time.Time

// Ticker receives a Series of Coins for each time Tick is called.
type Ticker struct {
	client Client
	clock  Clock
}

// NewTicker creates a new Ticker.
func NewTicker(client Client, clock Clock) *Ticker {
	return &Ticker{
		client: client,
		clock:  clock,
	}
}

// Tick fetches a Series of Coins from given endpoint.
func (t *Ticker) Tick() (asset.Series, error) {
	response, err := retrieve(t.client)
	if err != nil {
		return nil, err
	}

	responseObject, err := extractResponseObject(response.Body)
	if err != nil {
		return nil, err
	}

	series, err := mapToSeries(responseObject, t.clock())
	if err != nil {
		return nil, err
	}
	return series, nil
}

func retrieve(client Client) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, os.Getenv(envTickerEndpoint), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to crate response")
	}
	addQueryParams(req)
	req.Header.Set("X-CMC_PRO_API_KEY", os.Getenv(envTickerAPIKey))

	response, err := client.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve response from ticker")
	}
	if !isSuccessful(response) {
		return nil, fmt.Errorf("failed to retrieve successful response, got status [%d, %s]", response.StatusCode, response.Status)
	}
	return response, nil
}

func addQueryParams(req *http.Request) {
	q := req.URL.Query()
	q.Add("start", "1")
	q.Add("limit", "500")
	q.Add("convert", "USD")
	req.URL.RawQuery = q.Encode()
}

func isSuccessful(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode <= 299
}

func extractResponseObject(body io.ReadCloser) (asset.ResponseObject, error) {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return asset.ResponseObject{}, errors.Wrap(err, "failed to read response from ticker")
	}

	var responseObject asset.ResponseObject
	err = json.Unmarshal(bytes, &responseObject)
	if err != nil {
		return asset.ResponseObject{}, errors.Wrap(err, "failed to marshal response to quote")
	}
	return responseObject, nil
}

func mapToSeries(responseObject asset.ResponseObject, timestamp time.Time) (asset.Series, error) {
	series := make([]asset.Coin, len(responseObject.Data))
	for i, data := range responseObject.Data {
		series[i] = asset.Coin{
			Symbol:        data.Symbol,
			Name:          data.Name,
			PriceInDollar: data.Quote.Usd.Price,
			Timestamp:     timestamp,
		}
	}
	return series, nil
}
