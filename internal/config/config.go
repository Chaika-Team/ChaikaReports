package config

import (
	"log"

	"github.com/spf13/viper"
)

type StorageConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Keyspace string   `mapstructure:"keyspace"`
	User     string   `mapstructure:"user"`
	Password string   `mapstructure:"password"`
}

type Config struct {
	Cassandra     StorageConfig `mapstructure:"cassandra"`
	CassandraTest StorageConfig `mapstructure:"cassandra-test"`
}

func LoadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config, %s", err)
	}

	return &cfg
}
