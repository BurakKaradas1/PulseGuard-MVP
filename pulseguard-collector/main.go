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

// CORS Middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Geliştirme ortamı için "*" bırakılabilir veya spesifik React portu yazılabilir.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-PulseGuard-Signature")

		// Preflight (OPTIONS) isteğine doğrudan 200 OK dön
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

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
	http.HandleFunc("/api/v1/events", corsMiddleware(logRequest(apiServer.HandleReceiveEvents)))
	http.HandleFunc("/api/v1/dashboard/hosts", corsMiddleware(logRequest(apiServer.HandleGetHosts)))
	http.HandleFunc("/api/v1/dashboard/events", corsMiddleware(logRequest(apiServer.HandleGetEvents)))
	http.HandleFunc("/api/v1/dashboard/hosts/detail", corsMiddleware(logRequest(apiServer.HandleGetHostDetail)))
	http.HandleFunc("/api/v1/dashboard/hosts/threshold", corsMiddleware(logRequest(apiServer.HandleSetThreshold)))
	http.HandleFunc("/api/v1/agent/register", corsMiddleware(logRequest(apiServer.HandleRegisterHost)))
	// 4. Sunucuyu ayağa kaldır
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
