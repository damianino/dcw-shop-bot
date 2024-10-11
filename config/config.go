package config

import (
	"fmt"
	env "github.com/caarlos0/env/v10"
	"time"
)

const (
	CacheDefaultTTL       = time.Hour * 12
	DefaultRequestTimeout = time.Second * 5
	CtxDefaultTimeout     = time.Second * 50
)

type Config struct {
	TelegramCustomerBot struct {
		Token string `env:"TELEGRAM_CUSTOMER_TOKEN"`
	}

	TelegramAdminBot struct {
		Token string `env:"TELEGRAM_ADMIN_TOKEN"`
	}

	DB struct {
		Host     string `env:"DB_HOST"`
		Port     string `env:"DB_PORT"`
		Name     string `env:"DB_DBNAME"`
		Username string `env:"DB_USERNAME"`
		Password string `env:"DB_PASSWORD"`
	}
}

func GetConfig() (*Config, error) {
	envConfig := Config{}

	if err := env.Parse(&envConfig); err != nil {
		return &envConfig, fmt.Errorf("parse telegram config: %w", err)
	}

	return &envConfig, nil
}
