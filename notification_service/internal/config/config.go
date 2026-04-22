package config

import "github.com/ilyakaznacheev/cleanenv"

type AppConfig struct {
	DatabaseURL       string   `yaml:"database-url"`
	GRPCServerAddress string   `yaml:"grpc-server-address"`
	KafkaBrokers      []string `yaml:"kafka-brokers"`
	KafkaGroupID      string   `yaml:"kafka-group-id"`
}

func Load(cfgPath string) (*AppConfig, error) {
	cfg := AppConfig{}
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
