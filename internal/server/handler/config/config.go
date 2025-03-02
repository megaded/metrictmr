package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	defaultAddr = "localhost:8080"
)

type Config struct {
	Address *string
}

func (c *Config) GetAddress() string {
	return *c.Address
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
	flag.Parse()
}
