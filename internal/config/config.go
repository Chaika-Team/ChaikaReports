// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
	"log"
)

type StorageConfig struct {
	Hosts    []string
	Keyspace string
	User     string
	Password string
}

type Config struct {
	Cassandra StorageConfig
}

func LoadConfig() *Config {
	viper.SetConfigFile("config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config, %s", err)
	}

	return &cfg
}
