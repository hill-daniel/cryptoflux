package http_test

import (
	"github.com/hill-daniel/cryptoflux"
	"github.com/hill-daniel/cryptoflux/http"
	"io/ioutil"
	net "net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_should_create_coin_series_from_json_response(t *testing.T) {
	fixedDate := time.Date(2018, 11, 14, 13, 37, 0, 0, time.UTC)
	fixedClock := func() time.Time {
		return fixedDate
	}
	ticker := http.NewTicker(&testRestClient{}, fixedClock)

	series, err := ticker.Tick()

	if err != nil {
		t.Error(err)
	}
	if len(series) != 3 {
		t.Error("failed to retrieve stuff")
	}
	priceInDollar, _ := strconv.ParseFloat("3835.36079075", 64)
	expectedBitcoinResult := asset.Coin{
		Symbol:        "BTC",
		Name:          "Bitcoin",
		PriceInDollar: priceInDollar,
		Timestamp:     fixedDate,
	}
	if series[0].Symbol != expectedBitcoinResult.Symbol {
		t.Errorf("Expected %s, got %s", expectedBitcoinResult.Symbol, series[0].Symbol)
	}
	if series[0].Name != expectedBitcoinResult.Name {
		t.Errorf("Expected %s, got %s", expectedBitcoinResult.Name, series[0].Name)
	}
	if series[0].Timestamp != expectedBitcoinResult.Timestamp {
		t.Errorf("Expected %s, got %s", expectedBitcoinResult.Timestamp, series[0].Timestamp)
	}
	if series[0].PriceInDollar != expectedBitcoinResult.PriceInDollar {
		t.Errorf("Expected %f, got %f", expectedBitcoinResult.PriceInDollar, series[0].PriceInDollar)
	}
}

func Test_should_fail_on_unsuccessful_response_from_server(t *testing.T) {
	fixedDate := time.Date(2018, 11, 14, 13, 37, 0, 0, time.UTC)
	fixedClock := func() time.Time {
		return fixedDate
	}
	failedResponse := &net.Response{
		StatusCode: 500,
		Body:       nil,
	}
	ticker := http.NewTicker(&testRestClient{response: failedResponse}, fixedClock)

	series, err := ticker.Tick()

	if series != nil {
		t.Errorf("series should have been nil but was %v", series)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "failed to retrieve successful response") {
		t.Errorf("expected error with message containing: %s, but was: %v", "failed to retrieve successful response", err)
	}
}

type testRestClient struct {
	response *net.Response
}

func (c *testRestClient) Do(req *net.Request) (*net.Response, error) {
	if c.response != nil {
		return c.response, nil
	}
	return &net.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(exampleJsonBody)),
	}, nil
}

const exampleJsonBody = `{
  "status": {
    "timestamp": "2019-01-18T16:15:04.135Z",
    "error_code": 0,
    "error_message": null,
    "elapsed": 8,
    "credit_count": 1
  },
  "data": [
    {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "circulating_supply": 17488062,
      "total_supply": 17488062,
      "max_supply": 21000000,
      "date_added": "2013-04-28T00:00:00.000Z",
      "num_market_pairs": 6801,
      "tags": [
        "mineable"
      ],
      "platform": null,
      "cmc_rank": 1,
      "last_updated": "2019-01-18T16:13:25.000Z",
      "quote": {
        "USD": {
          "price": 3835.36079075,
          "volume_24h": 5221405183.24418,
          "percent_change_1h": 0.190455,
          "percent_change_24h": 0.399281,
          "percent_change_7d": -1.00124,
          "market_cap": 63992784295.12049,
          "last_updated": "2019-01-18T16:13:25.000Z"
        }
      }
    },
    {
      "id": 52,
      "name": "XRP",
      "symbol": "XRP",
      "slug": "ripple",
      "circulating_supply": 41040405095,
      "total_supply": 99991724864,
      "max_supply": 100000000000,
      "date_added": "2013-08-04T00:00:00.000Z",
      "num_market_pairs": 326,
      "tags": [],
      "platform": null,
      "cmc_rank": 2,
      "last_updated": "2019-01-18T16:13:04.000Z",
      "quote": {
        "USD": {
          "price": 0.32529476168,
          "volume_24h": 387813371.10633,
          "percent_change_1h": -0.173428,
          "percent_change_24h": -0.787073,
          "percent_change_7d": -2.81121,
          "market_cap": 13350228794.628683,
          "last_updated": "2019-01-18T16:13:04.000Z"
        }
      }
    },
    {
      "id": 1027,
      "name": "Ethereum",
      "symbol": "ETH",
      "slug": "ethereum",
      "circulating_supply": 104441667.6866,
      "total_supply": 104441667.6866,
      "max_supply": null,
      "date_added": "2015-08-07T00:00:00.000Z",
      "num_market_pairs": 4653,
      "tags": [
        "mineable"
      ],
      "platform": null,
      "cmc_rank": 3,
      "last_updated": "2019-01-18T16:13:19.000Z",
      "quote": {
        "USD": {
          "price": 122.042554949,
          "volume_24h": 2379446316.31767,
          "percent_change_1h": 0.0581254,
          "percent_change_24h": -0.193878,
          "percent_change_7d": -4.15337,
          "market_cap": 12746327967.607079,
          "last_updated": "2019-01-18T16:13:19.000Z"
        }
      }
    }
  ]
}`
