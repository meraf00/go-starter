package config

import (
	"time"

	"github.com/chariotplatform/goapi/logger"
)

type AppConfig struct {
	AppName     string
	Environment string
	StartedAt   time.Time
	Swagger     SwaggerConfig
	HTTP        ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	OryKratos   OryKratosConfig
}

type SwaggerConfig struct {
	Title       string
	Version     string
	BasePath    string
	DocEndpoint string
}

type ServerConfig struct {
	Host               string
	Port               int
	Cors               bool
	CORSAllowedOrigins []string
}

type DatabaseConfig struct {
	Database string
	Host     string
	Username string
	Password string
	Port     int
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type OryKratosConfig struct {
	URL string
}

func LoadConfig(logger logger.Log) *AppConfig {
	LoadEnv(logger)

	return &AppConfig{
		AppName:     GetEnvString("APP_NAME", "goapi", false),
		Environment: GetEnvironment("development"),
		StartedAt:   time.Now(),
		Swagger: SwaggerConfig{
			Title:       "goapi API",
			Version:     "1.0.0",
			BasePath:    "/api",
			DocEndpoint: "/docs",
		},
		HTTP: ServerConfig{
			Host:               GetEnvString("HTTP_HOST", "0.0.0.0", false),
			Port:               GetEnvNumber("HTTP_PORT", 8000, false),
			Cors:               true,
			CORSAllowedOrigins: GetEnvStringSlice("CORS_ALLOWED_ORIGINS", []string{}, false),
		},
		Database: DatabaseConfig{
			Database: GetEnvString("DB_NAME", "", true),
			Host:     GetEnvString("DB_HOST", "127.0.0.1", false),
			Username: GetEnvString("DB_USER", "", true),
			Password: GetEnvString("DB_PASS", "", true),
			Port:     GetEnvNumber("DB_PORT", 5432, false),
			SSLMode:  GetEnvString("DB_SSL_MODE", "disable", false),
		},
		Redis: RedisConfig{
			Addr:     GetEnvString("REDIS_ADDR", "127.0.0.1:6379", false),
			Password: GetEnvString("REDIS_PASSWORD", "", false),
			DB:       GetEnvNumber("REDIS_DB", 0, false),
		},
		OryKratos: OryKratosConfig{
			URL: GetEnvString("ORY_KRATOS_URL", "http://localhost:4433", false),
		},
	}
}
