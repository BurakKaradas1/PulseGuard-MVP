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

	fmt.Println("[*] Registering agent to Collector...")

	err = internal.RegisterAgent(cfg.Agent.CollectorURL)

	if err != nil {

		log.Fatalf("[-] Ajan kaydı başarısız oldu: %v", err)

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

	// Hardware checkerlar

	checkers = append(checkers, &internal.CpuChecker{Threshold: cfg.Thresholds.CPU})

	checkers = append(checkers, &internal.RamChecker{Threshold: cfg.Thresholds.RAM})

	checkers = append(checkers, &internal.DiskChecker{Threshold: cfg.Thresholds.Disk})

	// Ticker süresi YAML dosyasındaki değere göre dinamik başlat

	ticker := time.NewTicker(cfg.Agent.Interval)

	defer ticker.Stop()

	fmt.Printf("[+] PulseGuard Agent started. Monitoring every %s...\n", cfg.Agent.Interval)

	for {

		<-ticker.C

		fmt.Println("\n[*] Running scheduled checks...")

		//Tüm sensörleri çalıştır kuyruğa at

		for _, c := range checkers {

			result := c.Check()

			// YENİ: RAM'e append etmek yerine diske yazıyoruz

			err := internal.EnqueueEvent(result)

			if err != nil {

				fmt.Printf("[!] Disk yazma hatası: %v\n", err)

			}

			status := "[OK]"

			if !result.Passed {

				status = "[FAIL]"

			}

			fmt.Printf("[%s] %s %s: %s\n", result.Level, status, c.Name(), result.Message)

		}

		// Diskte biriken tüm verileri oku

		queuedEvents, _ := internal.DequeueAll()

		if len(queuedEvents) > 0 {

			eventsURL := cfg.Agent.CollectorURL + "/api/v1/events"

			// C2 sunucusuna toplu olarak gönder

			err := internal.SendBatch(queuedEvents, eventsURL)

			if err != nil {

				fmt.Printf("[!] Failed to push events to C2. Veriler diskte (WAL) güvende. | Current queue size: %d\n", len(queuedEvents))

			}
		} else {
			fmt.Printf("[+] Successfully pushed %d events to C2. Truncating processed WAL entries.\n", len(queuedEvents))
			// Sadece başarıyla gönderilenlerin sayısını veriyoruz, gerisini WAL koruyacak
			internal.RemoveProcessedEvents(len(queuedEvents))
		}

	}

}
