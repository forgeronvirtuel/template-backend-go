package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port uint   `mapstructure:"port"`
}

// Load reads configuration from viper and returns a Config struct
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)

	// Unmarshal into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		return fmt.Errorf("server port cannot be 0")
	}
	if c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}
	return nil
}

// Address returns the formatted host:port address
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
