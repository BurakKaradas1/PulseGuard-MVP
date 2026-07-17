package internal

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent struct {
		Interval     time.Duration `yaml:"interval"`
		CollectorURL string        `yaml:"collector_url"`
	} `yaml:"agent"`

	Thresholds struct {
		CPU  int `yaml:"cpu"`
		RAM  int `yaml:"ram"`
		Disk int `yaml:"disk"`
	} `yaml:"thresholds"`

	Checks struct {
		SystemIntegrity   bool `yaml:"system_integrity"`
		AnalysisDetection bool `yaml:"analysis_detection"`
		NetworkStatus     bool `yaml:"network_status"`
	} `yaml:"checks"`
}

// LoadConfig dosyayi okur ve Config struct'ina donusturur
func LoadConfig(filename string) (*Config, error) {
	// 1. Dosyayi diskten oku
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config

	// 2. YAML verisini Go struct'ina cevirir
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// 3. Basariliysa Config'in bellek adresini dondurur
	return &cfg, nil
}
