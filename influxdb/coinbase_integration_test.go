// +build integration

package influxdb_test

import (
	"encoding/json"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/hill-daniel/cryptoflux"
	"github.com/hill-daniel/cryptoflux/influxdb"
	"github.com/influxdata/influxdb1-client/v2"
	"log"
	"os"
	"testing"
	"time"
)

const influxDbName = "cryptocurrencies"

var EPSILON = 0.00000001

func Test_should_store_coin_series_to_influx(t *testing.T) {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("failed to connect to docker daemon: %s", err)
	}
	container, err := dockerClient.CreateContainer(createOptions("influxdb"))
	if err != nil {
		t.Fatalf("failed to create docker container: %s", err)
	}
	defer func() {
		if err := dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		}); err != nil {
			t.Fatalf("failed to remove container: %s", err)
		}
	}()
	startContainer(dockerClient, container, t)
	influxClient := createInfluxClient(err, t)
	setupDataBase(influxClient, t)
	coinBase := influxdb.NewInfluxCoinBase(influxClient)
	fixedDate := time.Date(2018, 11, 14, 13, 37, 0, 0, time.UTC)
	series := []asset.Coin{{
		Name:          "Bitcoin",
		Symbol:        "BTC",
		Timestamp:     fixedDate,
		PriceInDollar: 3874.8765},
	}
	_ = os.Setenv("INFLUX_DB_NAME", influxDbName)

	err = coinBase.Store(series)

	if err != nil {
		t.Fatal(err)
	}
	res, err := queryDB(influxClient, client.Query{
		Command:  "SELECT * FROM BTC",
		Database: influxDbName,
	})
	if err != nil {
		t.Fatal(err)
	}
	checkResult(res, t, fixedDate)
}

func createOptions(image string) docker.CreateContainerOptions {
	exposedCadvPort := map[docker.Port]struct{}{
		"8086/tcp": {}}

	containerConf := docker.Config{
		ExposedPorts: exposedCadvPort,
		Image:        image,
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"8086/tcp": {docker.PortBinding{HostPort: "8086"}}}

	hostConf := docker.HostConfig{
		PortBindings:    portBindings,
		PublishAllPorts: true,
		Privileged:      false,
	}

	return docker.CreateContainerOptions{
		Name:       "influxdb",
		Config:     &containerConf,
		HostConfig: &hostConf,
	}
}

func startContainer(dockerClient *docker.Client, container *docker.Container, t *testing.T) {
	err := dockerClient.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		t.Fatalf("failed to start docker container: %s", err)
	}
	// wait for container to wake up
	if err := waitStarted(dockerClient, container.ID, 5*time.Second); err != nil {
		t.Fatalf("failed to reach influxdb for testing, aborting.")
	}
	container, err = dockerClient.InspectContainer(container.ID)
	if err != nil {
		t.Fatalf("failed to inspect container: %s", err)
	}
}

func waitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("failed to wait for container %s to start for %v", id, maxWait)
}

func createInfluxClient(err error, t *testing.T) client.Client {
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:8086", "localhost"),
		Username: influxDbName,
		Password: "c6528f8536922Av",
	})
	if err != nil {
		t.Fatal("failed to create influx client", err)
	}
	return influxClient
}

func setupDataBase(influxClient client.Client, t *testing.T) {
	// wait for influxDb to accept connections
	err := waitReady(influxClient, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = queryDB(influxClient, client.Query{
		Command: fmt.Sprintf("CREATE DATABASE %s", influxDbName),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queryDB(influxClient, client.Query{
		Command: fmt.Sprintf("CREATE USER cryptoflux WITH PASSWORD '%s'", "c6528f8536922Av"),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queryDB(influxClient, client.Query{
		Command: fmt.Sprintf("GRANT ALL PRIVILEGES TO %s", "cryptoflux"),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func waitReady(influxClient client.Client, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		_, _, err := influxClient.Ping(500 * time.Millisecond)
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("failed to ping influxdb for %v", maxWait)
}

func queryDB(clnt client.Client, q client.Query) (res []client.Result, err error) {
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func checkResult(res []client.Result, t *testing.T, fixedDate time.Time) {
	if len(res[0].Series[0].Values) != 1 {
		t.Errorf("expected result count of 1, got %d", len(res[0].Series[0].Values))
	}
	row := res[0].Series[0].Values[0]
	timestamp, err := time.Parse(time.RFC3339, row[0].(string))
	if err != nil {
		log.Fatal(err)
	}
	if timestamp != fixedDate {
		t.Errorf("expected %v, got %v", fixedDate, timestamp)
	}
	name := row[1].(string)
	if name != "Bitcoin" {
		t.Errorf("expected %s, got %s", "Bitcoin", name)
	}
	priceInDollar, _ := row[2].(json.Number).Float64()
	if !floatEquals(priceInDollar, 3874.8765) {
		t.Errorf("expected %f, got %f", priceInDollar, 3874.8765)
	}
}

func floatEquals(a, b float64) bool {
	return (a-b) < EPSILON && (b-a) < EPSILON
}
