package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/megaded/metrictmr/internal/logger"
)

const (
	defaultAddr          = "localhost:8080"
	defaultStoreInternal = 0
	defaultFilePath      = "metric.txt"
	defaultRestore       = true
)

type Config struct {
	Address       string `env:"ADDRESS" json:"address"`
	StoreInterval *int   `env:"STORE_INTERVAL" json:"store_interval"`
	FilePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore       *bool  `env:"RESTORE,init" json:"restore"`
	DBConnString  string `env:"DATABASE_DSN" json:"database_dsn"`
	Key           string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key"`
}

func (c *Config) GetAddress() string {
	return c.Address
}

func (c *Config) GetFilePath() (fp string, isDefault bool) {
	return c.FilePath, c.FilePath == defaultFilePath
}

func GetConfig() *Config {
	config := &Config{}
	var configPath string
	flag.StringVar(&configPath, "c", "", "config file")
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()
	if configPath != "" {
		data, err := readJSONFile(configPath)
		if err != nil {
			logger.Log.Error(err.Error())
			panic(err)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			logger.Log.Error(err.Error())
			panic(err)
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

func readJSONFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл не найден: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	return data, nil
}
