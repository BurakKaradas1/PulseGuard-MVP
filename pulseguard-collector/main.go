package main

import (
	"fmt"
	"log"
	"net/http"

	"pulseguard-collector/internal/api"
	"pulseguard-collector/internal/storage"
)

// İleride konfigürasyon dosyasından (YAML/ENV) okunacak
const secretKey = "super-secret-pulseguard-key"

// logRequest middleware'i, gelen her isteği terminale yazar
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] Incoming Request: %s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

func main() {
	// 1. Veritabanı katmanını başlat
	repo, err := storage.NewSQLiteRepository("./pulseguard.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("[+] SQLite Database initialized successfully.")
	fmt.Println("PulseGuard C2 Server (Collector) Starting...")
	fmt.Println("Listening on port 8080 (SQLite active...)")

	// 2. API Server örneğini oluştur (Interface ve SecretKey'i enjekte et)
	apiServer := api.NewAPIServer(repo, secretKey)

	// 3. REST API Uç Noktaları
	http.HandleFunc("/api/v1/events", logRequest(apiServer.HandleReceiveEvents))                   // Ajan girişi
	http.HandleFunc("/api/v1/dashboard/hosts", logRequest(apiServer.HandleGetHosts))               // Filo görünümü
	http.HandleFunc("/api/v1/dashboard/events", logRequest(apiServer.HandleGetEvents))             // Zaman çizelgesi
	http.HandleFunc("/api/v1/dashboard/hosts/detail", logRequest(apiServer.HandleGetHostDetail))   // GET
	http.HandleFunc("/api/v1/dashboard/hosts/threshold", logRequest(apiServer.HandleSetThreshold)) // POST
	http.HandleFunc("/api/v1/agent/register", logRequest(apiServer.HandleRegisterHost))
	// 4. Sunucuyu ayağa kaldır
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
