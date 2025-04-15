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
	DBConnString  string `env:"DATABASE_DSN"`
	Key           string `env:"KEY"`
}

func (c *Config) GetAddress() string {
	return c.Address
}

func (c *Config) GetFilePath() (fp string, isDefault bool) {
	return c.FilePath, c.FilePath == defaultFilePath
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
	dBConnString := flag.String("d", "", "db conn string")
	storeInterval := flag.Int("i", defaultStoreInternal, "store internal")
	filePath := flag.String("f", defaultFilePath, "file path")
	restore := flag.Bool("r", defaultRestore, "restore")
	key := flag.String("k", "", "key")
	flag.Parse()
	if c.Address == "" {
		c.Address = *address
	}
	if c.DBConnString == "" {
		c.DBConnString = *dBConnString
	}
	if c.StoreInterval == nil {
		c.StoreInterval = storeInterval
	}
	if c.FilePath == "" {
		c.FilePath = *filePath
	}
	if c.Restore == nil {
		c.Restore = restore
	}
	if c.Key == "" {
		c.Key = *key
	}
}
