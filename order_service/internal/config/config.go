package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DatabaseURL           string   `yaml:"database-url" env:"DATABASE_URL"`
	GRPCServerAddress     string   `yaml:"grpc-server-address" env:"GRPC_SERVER_ADDRESS"`
	UserServiceAddress    string   `yaml:"user-service-address" env:"USER_SERVICE_ADDRESS"`
	ProductServiceAddress string   `yaml:"product-service-address" env:"PRODUCT_SERVICE_ADDRESS"`
	KafkaBrokers          []string `yaml:"kafka-brokers" env:"KAFKA_BROKERS"`
}

func Load(cfgPath string) (*Config, error) {
	cfg := Config{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
