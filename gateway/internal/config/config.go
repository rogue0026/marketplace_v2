package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	HTTPAddress                string `yaml:"http-address" env:"HTTP_ADDRESS"`
	ProductServiceAddress      string `yaml:"product-service-addr" env:"PRODUCT_SERVICE_ADDRESS"`
	UserServiceAddress         string `yaml:"user-service-addr" env:"USER_SERVICE_ADDRESS"`
	OrderServiceAddress        string `yaml:"order-service-addr" env:"ORDER_SERVICE_ADDRESS"`
	NotificationServiceAddress string `yaml:"notification-service-addr" env:"NOTIFICATION_SERVICE_ADDRESS"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("fail while loading app config: %w", err)
	}

	return &cfg, nil
}
