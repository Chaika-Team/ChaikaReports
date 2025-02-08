package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type StorageConfig struct {
	Hosts         []string      `mapstructure:"hosts" validate:"required,min=1"`
	Keyspace      string        `mapstructure:"keyspace" validate:"required"`
	User          string        `mapstructure:"user" validate:"required"`
	Password      string        `mapstructure:"password" validate:"required"`
	Timeout       time.Duration `mapstructure:"timeout" validate:"required"`
	RetryDelay    time.Duration `mapstructure:"retry_delay" validate:"required"`
	RetryAttempts int           `mapstructure:"retry_attempts" validate:"required"`
}

type ServerConfig struct {
	Port    string        `mapstructure:"port" validate:"required"`
	Timeout time.Duration `mapstructure:"timeout" validate:"required"`
}

type Config struct {
	Cassandra     StorageConfig `mapstructure:"cassandra"`
	CassandraTest StorageConfig `mapstructure:"cassandra-test"`
	Server        ServerConfig  `mapstructure:"http-server"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHAIKA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	return validate.Struct(cfg)
}
