package influxdb

import (
	"github.com/hill-daniel/cryptoflux"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
)

const envInfluxDbName = "INFLUX_DB_NAME"

// InfluxCoinBase provides functionality to store Series of Coins in Influxdb.
type InfluxCoinBase struct {
	client client.Client
}

// NewInfluxCoinBase creates a new InfluxCoinBase.
func NewInfluxCoinBase(client client.Client) *InfluxCoinBase {
	return &InfluxCoinBase{client: client}
}

// Store stores the given Series of Coins into Influxdb.
func (i InfluxCoinBase) Store(series asset.Series) error {
	defer func() {
		if err := i.client.Close(); err != nil {
			log.Error("failed to close influx client", err)
		}
	}()
	batchPoints, err := createBatchPoints()
	if err != nil {
		return errors.Wrap(err, "failed to create batch points for series")
	}
	for _, coin := range series {
		p, err := createPoint(coin)
		if err != nil {
			return err
		}
		batchPoints.AddPoint(p)
	}
	if err := i.client.Write(batchPoints); err != nil {
		return errors.Wrap(err, "failed to store batch for series")
	}
	return nil
}

func createBatchPoints() (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  os.Getenv(envInfluxDbName),
		Precision: "ns",
	})
	return bp, err
}

func createPoint(coin asset.Coin) (*client.Point, error) {
	tags := map[string]string{}
	fields := map[string]interface{}{
		"name":          coin.Name,
		"priceInDollar": coin.PriceInDollar,
	}
	point, err := client.NewPoint(
		coin.Symbol,
		tags,
		fields,
		coin.Timestamp,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create point for coin [%v]", coin)
	}
	return point, nil
}
