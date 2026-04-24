package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	HTTPAddress        string `yaml:"http-address" env:"HTTP_ADDRESS"`
	ProductServiceAddr string `yaml:"product-service-addr" env:"PRODUCT_SERVICE_ADDR"`
	UserServiceAddr    string `yaml:"user-service-addr" env:"USER_SERVICE_ADDR"`
	OrderServiceAddr   string `yaml:"order-service-addr" env:"ORDER_SERVICE_ADDR"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("fail while loading app config: %w", err)
	}

	return &cfg, nil
}
