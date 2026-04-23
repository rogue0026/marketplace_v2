package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	DatabaseURL       string `yaml:"database-url" env:"DATABASE_URL"`
	GRPCServerAddress string `yaml:"grpc-server-address" env:"GRPC_SERVER_ADDRESS"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
