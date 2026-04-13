package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	DatabaseURL       string `yaml:"database-url"`
	GRPCServerAddress string `yaml:"grpc-server-address"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
