package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caarlos0/env"
	"github.com/megaded/metrictmr/internal/logger"
)

const (
	defaultAddr    = "localhost:8080"
	reportInterval = 10
	pollInterval   = 2
)

type Config struct {
	Address        string `env:"ADDRESS" json:"address"`
	ReportInterval int64  `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int64  `env:"POLL_INTERVAL" json:"poll_interval"`
	Key            string `env:"KEY"`
	RateLimit      *int   `env:"RATE_LIMIT"`
	CryptoKey      string `evn:"CRYPTO_KEY" json:"crypto_key"`
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

func (c *Config) GetRateLimit() int {
	return *c.RateLimit
}

func (c *Config) GetCryptoKeyPath() string {
	return *&c.CryptoKey
}

func GetConfig() *Config {
	config := &Config{}
	var configPath string
	flag.StringVar(&configPath, "c", "", "config file")
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()
	if configPath != "" {
		data, err := readJSONFile(configPath)
		if err == nil {
			err = json.Unmarshal(data, &config)
			if err == nil {
				logger.Log.Fatal(err.Error())
			}
		}
	}
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
	rateLimit := flag.Int("l", 10, "rate limit")
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
	if c.RateLimit == nil {
		c.RateLimit = rateLimit
	}
}

func readJSONFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл не найден: %s", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	return data, nil
}
