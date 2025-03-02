package config

import (
	"flag"
)

const (
	defaultAddr    = "localhost:8080"
	reportInterval = 10
	pollInterval   = 10
)

type Config struct {
	Address        string
	ReportInterval int64
	PollInterval   int64
}

func (c *Config) GetAddress() string {
	return c.Address
}

func (c *Config) GetReportInterval() int64 {
	return c.ReportInterval
}

func (c *Config) GetPoolInterval() int64 {
	return c.PollInterval
}

func GetConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.Address, "a", defaultAddr, "server endpoint")
	flag.Int64Var(&config.ReportInterval, "r", reportInterval, "reportInterval")
	flag.Int64Var(&config.PollInterval, "p", reportInterval, "pollInterval")
	flag.Parse()
	return config
}
