package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	OAuth    OAuthConfig
	TMDB     TMDBConfig
	Session  SessionConfig
}

type ServerConfig struct {
	Env  string
	Port string
	Host string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	TLS      bool
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	CallbackHost       string
}

type TMDBConfig struct {
	APIKey       string
	BaseURL      string
	ImageBaseURL string
}

type SessionConfig struct {
	SecretKey string
}

// Load reads environment variables and returns a Config struct
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Env:  getEnv("NODE_ENV", "local"),
			Port: getEnv("PORT", "4000"),
			Host: getEnv("HOST", "http://localhost:4000"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			TLS:      getEnv("REDIS_TLS", "false") == "true",
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			CallbackHost:       getEnv("HOST", "http://localhost:4000"),
		},
		TMDB: TMDBConfig{
			APIKey:       getEnv("TMDB_KEY", ""),
			BaseURL:      getEnv("TMDB_URL", "https://api.themoviedb.org"),
			ImageBaseURL: getEnv("TMDB_IMAGE_URL", "https://image.tmdb.org/t/p/w500"),
		},
		Session: SessionConfig{
			SecretKey: getEnv("SECRET_KEY", ""),
		},
	}

	// Validate required fields
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Session.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY is required")
	}
	if len(cfg.Session.SecretKey) < 32 {
		return nil, fmt.Errorf("SECRET_KEY must be at least 32 characters")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// IsDevelopment returns true if running in development/local mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "local" || c.Server.Env == "development"
}

// RedisAddr returns the Redis address in host:port format
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}
