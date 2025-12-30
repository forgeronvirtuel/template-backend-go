package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		setupViper  func()
		wantErr     bool
		expectedCfg *Config
	}{
		{
			name: "default values",
			setupViper: func() {
				viper.Reset()
			},
			wantErr: false,
			expectedCfg: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
			},
		},
		{
			name: "custom values",
			setupViper: func() {
				viper.Reset()
				viper.Set("server.host", "127.0.0.1")
				viper.Set("server.port", 3000)
			},
			wantErr: false,
			expectedCfg: &Config{
				Server: ServerConfig{
					Host: "127.0.0.1",
					Port: 3000,
				},
			},
		},
		{
			name: "invalid port - zero",
			setupViper: func() {
				viper.Reset()
				viper.Set("server.port", 0)
			},
			wantErr: true,
		},
		// Note: Can't test port > 65535 with Viper as uint16 max is 65535
		{
			name: "empty host",
			setupViper: func() {
				viper.Reset()
				viper.Set("server.host", "")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()

			cfg, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCfg.Server.Host, cfg.Server.Host)
			assert.Equal(t, tt.expectedCfg.Server.Port, cfg.Server.Port)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config - localhost",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 3000,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - port zero",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 0,
				},
			},
			wantErr: true,
			errMsg:  "server port cannot be 0",
		},
		// Note: Can't test port > 65535 directly as uint16 max is 65535
		{
			name: "invalid - empty host",
			config: &Config{
				Server: ServerConfig{
					Host: "",
					Port: 8080,
				},
			},
			wantErr: true,
			errMsg:  "server host cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerConfigAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   ServerConfig
		expected string
	}{
		{
			name: "default address",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			expected: "0.0.0.0:8080",
		},
		{
			name: "localhost",
			config: ServerConfig{
				Host: "localhost",
				Port: 3000,
			},
			expected: "localhost:3000",
		},
		{
			name: "127.0.0.1",
			config: ServerConfig{
				Host: "127.0.0.1",
				Port: 9000,
			},
			expected: "127.0.0.1:9000",
		},
		{
			name: "high port",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 65535,
			},
			expected: "0.0.0.0:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := tt.config.Address()
			assert.Equal(t, tt.expected, addr)
		})
	}
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("port boundary - min valid", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Host: "0.0.0.0",
				Port: 1,
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("port boundary - max valid", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Host: "0.0.0.0",
				Port: 65535,
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	// Note: Can't test port > 65535 directly as uint16 max is 65535
	// The type system prevents invalid values at compile time
}
