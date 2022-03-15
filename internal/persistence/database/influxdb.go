package database

import (
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// InfluxDBConfig - config entry describing a influxdb config
type InfluxDBConfig struct {
	Host     string `json:"host" toml:"host"`
	Database string `json:"database" toml:"database"`
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
}

// InfluxDBConnect - opens a connection to influxdb and returns the connection object
func InfluxDBConnect(conf *InfluxDBConfig) (client.Client, error) {
	cl, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     conf.Host,
		Username: conf.Username,
		Password: conf.Password,
	})
	if err != nil {
		return nil, err
	}
	return cl, nil
}
