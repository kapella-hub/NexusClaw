package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	DSN      string `mapstructure:"dsn"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	TokenSecret string        `mapstructure:"token_secret"`
	TokenExpiry time.Duration `mapstructure:"token_expiry"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// OAuthProviderConfig holds settings for a single OAuth provider.
type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
}

// Config is the root configuration for the application.
type Config struct {
	Server   ServerConfig                   `mapstructure:"server"`
	Database DatabaseConfig                 `mapstructure:"database"`
	Redis    RedisConfig                    `mapstructure:"redis"`
	Auth     AuthConfig                     `mapstructure:"auth"`
	Log      LogConfig                      `mapstructure:"log"`
	OAuth    map[string]OAuthProviderConfig `mapstructure:"oauth"`
}

// Load reads configuration from the file at path and environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.dsn", "postgres://localhost:5432/nexusclaw?sslmode=disable")
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.min_conns", 2)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("auth.token_secret", "")
	v.SetDefault("auth.token_expiry", 24*time.Hour)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("nexusclaw")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
	}
	v.AutomaticEnv()

	// Tolerate missing config file; env vars and defaults still apply.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && path != "" {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
