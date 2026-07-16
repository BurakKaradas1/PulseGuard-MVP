package main

import (
	"fmt"
	"log"
	"pulseguardv1/internal"
	"time"
)

func main() {
	// 1. YAML yapılandırma dosyasını yükle
	cfg, err := internal.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("[-] Failed to load configuration: %v", err)
	}

	// 2. Aktif kontrolleri dinamik olarak belirle
	var checkers []internal.Checker

	if cfg.Checks.SystemIntegrity {
		checkers = append(checkers, &internal.SystemChecker{})
	}
	if cfg.Checks.AnalysisDetection {
		checkers = append(checkers, &internal.AnalysisChecker{})
	}
	if cfg.Checks.NetworkStatus {
		// YAML'dan okunan toplayıcı (C2) adresini enjekte ediyoruz
		checkers = append(checkers, &internal.NetworkChecker{URL: cfg.Agent.CollectorURL})
	}

	// 3. Ticker süresini YAML dosyasındaki değere göre dinamik başlat
	ticker := time.NewTicker(cfg.Agent.Interval)
	defer ticker.Stop()

	fmt.Printf("[+] PulseGuard Agent started. Monitoring every %s...\n", cfg.Agent.Interval)

	for range ticker.C {
		for _, c := range checkers {
			event := c.Check()
			if event.Passed {
				fmt.Printf("[%s] [OK] %s: %s\n", event.Level, c.Name(), event.Message)
			} else {
				fmt.Printf("[%s] [FAIL] %s: %s\n", event.Level, c.Name(), event.Message)
			}
		}
	}
}
