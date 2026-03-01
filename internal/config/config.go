package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config holds every configuration value the application needs
type Config struct {
	AppEnv string

	Server   ServerConfig
	Logger   *zap.Logger
	DB       DBConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN returns the PostgreSQL connection string
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	// S3Endpoint allows pointing to LocalStack or MinIO in dev/test.
	S3Endpoint      string
	PresignedURLTTL time.Duration
}

type SecurityConfig struct {
	BcryptCost    int
	RateLimiteRPS float64
}

// Load reads configuration from .env.
func Load() (*Config, error) {
	// Load .env only in non-production environments
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load() // non-fatal if file is absent
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h")) // 7 days
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	connLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}
	presignedURLTTL, err := time.ParseDuration(getEnv("AWS_PRESIGNED_URL_TTL", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid AWS_PRESIGNED_URL_TTL: %w", err)
	}

	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	maxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	rps, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "100"), 64)

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:            mustEnv("DB_HOST"),
			Port:            mustEnv("DB_PORT"),
			User:            mustEnv("DB_USER"),
			Password:        mustEnv("DB_PASSWORD"),
			Name:            mustEnv("DB_NAME"),
			SSLMode:         mustEnv("DB_SSL_MODE"),
			MaxOpenConns:    maxOpen,
			MaxIdleConns:    maxIdle,
			ConnMaxLifetime: connLifetime,
		},
		JWT: JWTConfig{
			Secret:     mustEnv("JWT_SECRET"),
			AccessTTL:  accessTTL,
			RefreshTTL: refreshTTL,
		},

		AWS: AWSConfig{
			Region:          mustEnv("AWS_REGION"),
			AccessKeyID:     mustEnv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: mustEnv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:        mustEnv("AWS_S3_BUCKET"),
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", ""),
			PresignedURLTTL: presignedURLTTL,
		},
		Security: SecurityConfig{
			BcryptCost:    bcryptCost,
			RateLimiteRPS: rps,
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// mustEnv panics with a descriptive message when a required variable is absent.
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
