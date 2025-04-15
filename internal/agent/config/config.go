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
	if c.Address == "" {
		flag.StringVar(&c.Address, "a", defaultAddr, "server endpoint")
	}
	if c.ReportInterval == 0 {
		flag.Int64Var(&c.ReportInterval, "r", reportInterval, "reportInterval")
	}
	if c.PollInterval == 0 {
		flag.Int64Var(&c.PollInterval, "p", pollInterval, "pollInterval")
	}
	if c.Key == "" {
		flag.StringVar(&c.Key, "k", "", "key")
	}
	flag.Parse()
}
