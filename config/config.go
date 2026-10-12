package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type AuthConfig struct {
	BearerToken string `json:"bearer_token"`
}

type WhatsAppConfig struct {
	DefaultRecipient string  `json:"default_recipient"`
	SessionDBPath    string  `json:"session_db_path"`
	DeviceName       string  `json:"device_name"`
	RateLimitSeconds float64 `json:"rate_limit_seconds"`
}

type LogConfig struct {
	Level string `json:"level"`
}

type Config struct {
	Server   ServerConfig   `json:"server"`
	Auth     AuthConfig     `json:"auth"`
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Log      LogConfig      `json:"log"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
			path = envPath
		} else {
			path = "config.json"
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		WhatsApp: WhatsAppConfig{
			SessionDBPath:    "./storage/wa_session.db",
			DeviceName:       "WA Notification Gateway",
			RateLimitSeconds: 1.5,
		},
		Log: LogConfig{
			Level: "info",
		},
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if cfg.Server.Port <= 0 {
		cfg.Server.Port = 8080
	}
	if cfg.WhatsApp.SessionDBPath == "" {
		cfg.WhatsApp.SessionDBPath = "./storage/wa_session.db"
	}
	if cfg.WhatsApp.RateLimitSeconds <= 0 {
		cfg.WhatsApp.RateLimitSeconds = 1.0
	}

	return cfg, nil
}
