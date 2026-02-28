// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Path string
}

type EmailConfig struct {
	ResendAPIKey string
	FromEmail    string
	TestEmail    string `json:"testEmail"`
}

type AppConfig struct {
    Server      ServerConfig
    Database    DatabaseConfig
    Email       EmailConfig
    Debug       bool   // true = разработка, false = продакшен
    LogLevel    string
}

// Вспомогательные методы
func (c *AppConfig) IsProd() bool {
    return !c.Debug
}

func (c *AppConfig) IsDev() bool {
    return c.Debug
}

func Load() *AppConfig {
	loadEnvFile()
	return &AppConfig{
		Server: ServerConfig{
			Port:         normalizePort(getEnv("PORT", ":8080")),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "taskflow.db"),
		},
		Email: EmailConfig{
			ResendAPIKey: getEnv("RESEND_API_KEY", ""),
			FromEmail:    getEnv("EMAIL_FROM", "noreply@resend.dev"),
			TestEmail:    getEnv("TEST_EMAIL", ""),
		},
		Debug:    getEnvAsBool("DEBUG", false),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func loadEnvFile() {
    err := godotenv.Load()
    if err != nil {
        fmt.Println("⚠️ .env not found in current directory, trying parent...")

        // Пробуем найти .env в корне проекта
        dir, _ := os.Getwd()
        for {
            envPath := filepath.Join(dir, ".env")
            if err := godotenv.Load(envPath); err == nil {
                fmt.Printf("✅ Loaded .env from: %s\n", envPath)
                return
            }

            parent := filepath.Dir(dir)
            if parent == dir {
                break
            }
            dir = parent
        }

        fmt.Println("❌ No .env file found!")
    } else {
        fmt.Println("✅ Loaded .env from current directory")
    }
}

// normalizePort приводит порт к формату ":8080"
func normalizePort(port string) string {
	port = strings.TrimSpace(port)

	if port == "" {
		return ":8080"
	}

	if strings.HasPrefix(port, ":") {
		if _, err := strconv.Atoi(port[1:]); err == nil {
			return port
		}
		return ":8080"
	}

	if _, err := strconv.Atoi(port); err == nil {
		return ":" + port
	}

	return ":8080"
}

// getEnv читает переменную окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
