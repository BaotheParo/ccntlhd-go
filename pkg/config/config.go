package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	AppName string `mapstructure:"app_name"`
	Env     string `mapstructure:"env"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config/config.yaml") // Or viper.AddConfigPath("config") + viper.SetConfigName("config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
