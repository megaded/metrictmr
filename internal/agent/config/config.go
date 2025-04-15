package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	defaultAddr    = "localhost:8080"
	reportInterval = 10
	pollInterval   = 2
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
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

func (c *Config) GetKey() string {
	return c.Key
}

func GetConfig() *Config {
	config := &Config{}
	setEnvParam(config)
	setCmdParam(config)
	return config
}

func setEnvParam(c *Config) {
	env.Parse(c)
}

func setCmdParam(c *Config) {
	address := flag.String("a", defaultAddr, "server endpoint")
	reportInterval := flag.Int64("r", reportInterval, "reportInterval")
	pollInterval := flag.Int64("p", pollInterval, "pollInterval")
	key := flag.String("k", "", "key")
	flag.Parse()
	if c.Address == "" {
		c.Address = *address
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = *reportInterval
	}
	if c.PollInterval == 0 {
		c.PollInterval = *pollInterval
	}
	if c.Key == "" {
		c.Key = *key
	}
}
