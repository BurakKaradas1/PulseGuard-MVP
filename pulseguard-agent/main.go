package main

import (
	"fmt"
	"log"
	"pulseguard-agent/internal"
	"time"
)

func main() {
	// 1. YAML yapılandırma dosyası
	cfg, err := internal.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("[-] Failed to load configuration: %v", err)
	}

	var checkers []internal.Checker

	if cfg.Checks.SystemIntegrity {
		checkers = append(checkers, &internal.SystemChecker{})
	}
	if cfg.Checks.AnalysisDetection {
		checkers = append(checkers, &internal.AnalysisChecker{})
	}
	if cfg.Checks.NetworkStatus {
		// YAML'dan okunan toplayıcı adresini enjekte ediyoruz
		checkers = append(checkers, &internal.NetworkChecker{URL: cfg.Agent.CollectorURL})
	}

	// Hardware Checkers with YAML thresholds
	checkers = append(checkers, &internal.CpuChecker{Threshold: cfg.Thresholds.CPU})
	checkers = append(checkers, &internal.RamChecker{Threshold: cfg.Thresholds.RAM})
	checkers = append(checkers, &internal.DiskChecker{Threshold: cfg.Thresholds.Disk})

	// 3. Ticker süresi YAML dosyasındaki değere göre dinamik başlat
	ticker := time.NewTicker(cfg.Agent.Interval)
	defer ticker.Stop()

	fmt.Printf("[+] PulseGuard Agent started. Monitoring every %s...\n", cfg.Agent.Interval)

	//Kuyruk
	var eventQueue []internal.Event

	for {
		<-ticker.C
		fmt.Println("\n[*] Running scheduled checks...")

		//Tüm sensörleri çalıştır kuyruğa at
		for _, c := range checkers {
			result := c.Check()
			eventQueue = append(eventQueue, result)

			status := "[OK]"
			if !result.Passed {
				status = "[FAIL]"
			}
			fmt.Printf("[%s] %s %s: %s\n", result.Level, status, c.Name(), result.Message)
		}

		//Kuyruktaki tüm verileri C2 sunucusuna toplu olarak gönder
		err := internal.SendBatch(eventQueue, cfg.Agent.CollectorURL)

		if err != nil {
			fmt.Printf("[!] Failed to push events to C2. Hata Detayı: %v | Current queue size: %d\n", err, len(eventQueue))
		} else {
			// Gönderim başarılı olursa kuyruğu boşalt
			fmt.Printf("[+] Successfully pushed %d events to C2. Clearing queue.\n", len(eventQueue))
			eventQueue = nil
		}
	}
}
