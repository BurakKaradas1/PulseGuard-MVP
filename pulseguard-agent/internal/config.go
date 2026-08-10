package internal

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type CustomCommand struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type CustomHTTP struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

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

	CustomChecks struct {
		Commands      []CustomCommand `yaml:"commands"`
		HttpEndpoints []CustomHTTP    `yaml:"http_endpoints"`
	} `yaml:"custom_checks"`
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

// LoadDotEnv, proje kokunde bulunan bir .env dosyasini okuyup icindeki
// KEY=VALUE satirlarini gercek process ortam degiskeni olarak yukler.

func LoadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		// .env dosyasi opsiyoneldir (ornegin production'da gercek
		// ortam degiskenleri zaten set edilmis olabilir); yoksa sessizce gec.
		return
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

		// Gercek bir ortam degiskeni zaten tanimliysa .env'deki degeri
		// kullanmayarak gercek environment'in her zaman oncelikli
		// olmasini sagliyoruz.
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
