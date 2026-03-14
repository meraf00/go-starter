package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/chariotplatform/goapi/logger"
	"github.com/joho/godotenv"
)

const missingValueErrorTemplate = "Required environment variable %s is undefined and has no default"

// Load environment variables based on GO_APP_ENV
func LoadEnv(logger logger.Log) {
	err := godotenv.Load()
	if err != nil {
		logger.Warn(err)
		return
	}

	logger.Infof("Loaded .env")
}

// Environment types
var validEnvironments = []string{"development", "production", "test"}

// EnvironmentConfig struct
type EnvironmentConfig struct {
	Environment string
}

// GetEnvironment validates and retrieves the NODE_ENV variable
func GetEnvironment(defaultValue string) string {
	env := os.Getenv("GO_APP_ENV")
	if env == "" {
		env = defaultValue
		err := os.Setenv("GO_APP_ENV", env)
		if err != nil {
			fmt.Print("error setting GO_APP_ENV")
		}
	}

	if slices.Contains(validEnvironments, env) {
		return env
	}

	panic(fmt.Sprintf("Invalid NODE_ENV value. Accepted values: %v", validEnvironments))
}

// GetEnvString retrieves a string environment variable with a fallback
func GetEnvString(key string, defaultValue string, required bool) string {
	value := os.Getenv(key)
	if value == "" && defaultValue == "" && required {
		panic(fmt.Sprintf(missingValueErrorTemplate, key))
	}
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvStringSlice retrieves a slice of strings from an environment variable, split by commas, with a fallback
func GetEnvStringSlice(key string, defaultValue []string, required bool) []string {
	value := os.Getenv(key)
	if value == "" && len(defaultValue) == 0 && required {
		panic(fmt.Sprintf(missingValueErrorTemplate, key))
	}
	if value == "" {
		return defaultValue
	}
	origins := strings.Split(value, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

// GetEnvNumber retrieves a number environment variable with a fallback
func GetEnvNumber(key string, defaultValue int, required bool) int {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == 0 && required {
			panic(fmt.Sprintf(missingValueErrorTemplate, key))
		}
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Invalid number format for %s: %s", key, value))
	}
	return intValue
}
