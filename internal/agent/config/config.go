package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	defaultAddr    = "localhost:8080"
	reportInterval = 10
	pollInterval   = 10
)

type Config struct {
	Address        *string `env:"HOME"`
	ReportInterval *int64  `env:"HOME"`
	PollInterval   *int64  `env:"HOME"`
}

func (c *Config) GetAddress() string {
	return *c.Address
}

func (c *Config) GetReportInterval() int64 {
	return *c.ReportInterval
}

func (c *Config) GetPoolInterval() int64 {
	return *c.PollInterval
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
	if c.Address == nil || *c.Address == "" {
		flag.StringVar(c.Address, "a", defaultAddr, "server endpoint")
	}
	if c.ReportInterval == nil || *c.PollInterval == 0 {
		flag.Int64Var(c.ReportInterval, "r", reportInterval, "reportInterval")
	}
	if c.PollInterval == nil || *c.PollInterval == 0 {
		flag.Int64Var(c.PollInterval, "p", reportInterval, "pollInterval")
	}
	flag.Parse()
}
