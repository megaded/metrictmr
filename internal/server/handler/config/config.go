package config

import "flag"

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
	config.Address = flag.String("a", defaultAddr, "server endpoint")
	flag.Parse()
	return config
}
