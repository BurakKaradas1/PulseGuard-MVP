package config

import (
	"fmt"
	"os"
)

// Sistemin genel ayarlarını tutan yapı
type Config struct {
	SecretKey string
	DBPath    string
	Port      string
}

// Ortam değişkenlerini ve varsayılan ayarları döndürür
func LoadConfig() (*Config, error) {
	secret := os.Getenv("PULSEGUARD_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SECURITY ALERT: environment variable not found")
	}

	return &Config{
		SecretKey: secret,
		DBPath:    "./pulseguard.db",
		Port:      ":8080",
	}, nil
}
