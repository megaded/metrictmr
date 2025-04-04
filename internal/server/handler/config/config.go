package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	defaultAddr          = "localhost:8080"
	defaultStoreInternal = 0
	defaultFilePath      = "metric.txt"
	defaultRestore       = true
)

type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval *int   `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       *bool  `env:"RESTORE,init"`
}

func (c *Config) GetAddress() string {
	return c.Address
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
	if c.StoreInterval == nil {
		internal := 0
		c.StoreInterval = &internal
		flag.IntVar(c.StoreInterval, "i", defaultStoreInternal, "store internal")
	}
	if c.FilePath == "" {
		flag.StringVar(&c.FilePath, "f", defaultFilePath, "file path")
	}
	if c.Restore == nil {
		restore := false
		c.Restore = &restore
		flag.BoolVar(c.Restore, "r", defaultRestore, "restore")
	}
	flag.Parse()
}
