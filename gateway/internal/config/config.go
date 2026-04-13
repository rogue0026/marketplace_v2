package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	HTTPAddress        string `yaml:"http-address"`
	ProductServiceAddr string `yaml:"product-service-addr"`
	UserServiceAddr    string `yaml:"user-service-addr"`
	OrderServiceAddr   string `yaml:"order-service-addr"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("fail while loading app config: %w", err)
	}

	return &cfg, nil
}
