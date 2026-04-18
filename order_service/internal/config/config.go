package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DatabaseURL           string   `yaml:"database-url"`
	GRPCServerAddress     string   `yaml:"grpc-server-address"`
	UserServiceAddress    string   `yaml:"user-service-address"`
	ProductServiceAddress string   `yaml:"product-service-address"`
	KafkaBrokers          []string `yaml:"kafka-brokers"`
}

func Load(cfgPath string) (*Config, error) {
	cfg := Config{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
