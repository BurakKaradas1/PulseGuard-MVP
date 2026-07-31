package config

import (
	"fmt"
	"os"
	"strings"
)

// Sistemin genel ayarlarını tutan yapı
type Config struct {
	SecretKey string
	DBPath    string
	Port      string
}

// loadDotEnv, proje kokundeki .env dosyasini gercek process ortam
// degiskenlerine yukler. Go standart kutuphanesi .env dosyalarini
// otomatik okumadigi (ve bu projede godotenv gibi harici bir paket
// kullanilmadigi) icin, .env dosyasindaki PULSEGUARD_SECRET normalde
// hic bir zaman os.Getenv tarafindan gorulmuyordu. Zaten tanimli
// gercek ortam degiskenlerinin uzerine yazmaz.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // .env dosyasi opsiyoneldir
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

// Ortam değişkenlerini ve varsayılan ayarları döndürür
func LoadConfig() (*Config, error) {
	loadDotEnv(".env")

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
