package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port            string
    DBConnection    string
    MaxMessageSize  int64
    ReadTimeout     int
    WriteTimeout    int
    MinUsernameLen  int
    MaxUsernameLen  int
}

func LoadConfig() *Config {
    return &Config{
        Port:           getEnv("PORT", "8080"),
        DBConnection:   getEnv("DB_CONNECTION", "postgres://chatuser:chatpass@localhost:5432/chatdb?sslmode=disable"),
        MaxMessageSize: getEnvAsInt64("MAX_MESSAGE_SIZE", 4096),
        ReadTimeout:    getEnvAsInt("READ_TIMEOUT", 60),
        WriteTimeout:   getEnvAsInt("WRITE_TIMEOUT", 60),
        MinUsernameLen: getEnvAsInt("MIN_USERNAME_LEN", 3),
        MaxUsernameLen: getEnvAsInt("MAX_USERNAME_LEN", 10),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
            return intVal
        }
    }
    return defaultValue
} 