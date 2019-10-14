package main

import (
	"fmt"
	"github.com/hill-daniel/cryptoflux"
	"github.com/hill-daniel/cryptoflux/http"
	"github.com/hill-daniel/cryptoflux/influxdb"
	"github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
	net "net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	enInfluxDbHost      = "INFLUX_HOST"
	envInfluxDbUser     = "INFLUX_USER"
	envInfluxDbPassword = "INFLUX_PW"
	envLogLevel         = "LOG_LEVEL"
	envPollInterval     = "POLL_INTERVAL"
)

func init() {
	lvl, err := log.ParseLevel(os.Getenv(envLogLevel))
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(lvl)
	log.SetOutput(os.Stdout)
}

func main() {
	httpClient := &net.Client{Timeout: time.Second * 10}
	quoteTicker := http.NewTicker(httpClient, time.Now)
	coinBase := createCoinBase()

	done := make(chan struct{})
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(getPollInterval())
	log.Infof("Started cryptoflux application, polling ticker...")

	go func() {
		for {
			select {
			case <-ticker.C:
				series, err := fetchQuotes(quoteTicker)
				if err != nil {
					log.Error(err)
					continue
				}
				if err = storeSeries(series, coinBase); err != nil {
					log.Error(err)
				}
			case <-sigc:
				log.Info("Got an interrupt, stopping...")
				ticker.Stop()
				close(done)
			}
		}
	}()
	<-done
}

func getPollInterval() time.Duration {
	interval, err := strconv.Atoi(os.Getenv(envPollInterval))
	if err != nil {
		log.Warnf("failed to parse poll interval from [%s]. Using fallback.", os.Getenv(envPollInterval))
		interval = 10
	}
	return time.Duration(interval) * time.Minute
}

func createCoinBase() asset.CoinBase {
	influxClient, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:8086", os.Getenv(enInfluxDbHost)),
		Username: os.Getenv(envInfluxDbUser),
		Password: os.Getenv(envInfluxDbPassword),
	})
	return influxdb.NewInfluxCoinBase(influxClient)
}

func fetchQuotes(quoteTicker *http.Ticker) (asset.Series, error) {
	start := time.Now()
	series, err := quoteTicker.Tick()
	if err != nil {
		return nil, err
	}
	log.Infof("received series in %f seconds", time.Since(start).Seconds())
	return series, nil
}

func storeSeries(series asset.Series, coinBase asset.CoinBase) error {
	start := time.Now()
	err := coinBase.Store(series)
	if err != nil {
		return err
	}
	log.Infof("wrote %d coins in %f seconds", len(series), time.Since(start).Seconds())
	return nil
}
