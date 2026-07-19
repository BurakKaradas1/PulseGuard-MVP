package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"pulseguard-collector/internal/api"
	"pulseguard-collector/internal/storage"
)

// İleride confige alınacak
const secretKey = "super-secret-pulseguard-key"

var repo storage.Repository

// Bütünlük kanıtı
func verifySignature(payload []byte, signature string) bool {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// Ajandan gelen batch isteklerini karşılayan endpoint
func handleReceiveEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ajanın mührünü al
	agentSignature := r.Header.Get("X-PulseGuard-Signature")
	if agentSignature == "" {
		http.Error(w, "Missing Signature", http.StatusUnauthorized)
		return
	}

	// HTTP gövdesini oku
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Mühür kontrolü
	if !verifySignature(bodyBytes, agentSignature) {
		fmt.Println("[!] SECURITY ALERT: Invalid signature detected! Potential tampering")
		http.Error(w, "Invalid Signature", http.StatusUnauthorized)
		return
	}

	// Mühür geçerliyse storage paketindeki Event struct'ını kullanarak JSON ayrıştır
	var events []storage.Event
	err = json.NewDecoder(r.Body).Decode(&events)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Eventleri arayüz fonksiyonu üzerinden SQLite'a kaydet
	for _, event := range events {
		err = repo.SaveEvent(event.Level, event.Message, event.Passed)
		if err != nil {
			fmt.Printf("Failed to insert event: %v\n", err)
		}
	}

	fmt.Printf("[+] Successfully verified and stored %d events from agent.\n", len(events))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Batch processed successfully"))
}

// Sunucuya gelen her isteği terminale basar.
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] Incoming Request: %s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

func main() {
	// 1. Veritabanı katmanını başlat
	var err error
	repo, err = storage.NewSQLiteRepository("./pulseguard.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("[+] SQLite Database initialized successfully.")
	fmt.Println("PulseGuard C2 Server (Collector) Starting...")
	fmt.Println("Listening on port 8080 (SQLite active...)")

	// 2. API Server örneğini oluştur
	apiServer := api.NewAPIServer(repo)

	// 3. REST API Uç Noktaları
	// Ajanın (Agent) log gönderdiği rota
	http.HandleFunc("/api/v1/events", logRequest(handleReceiveEvents))

	// Dashboard'un (React) veri çekeceği rotalar
	http.HandleFunc("/api/v1/dashboard/hosts", logRequest(apiServer.HandleGetHosts))   // Filo görünümü
	http.HandleFunc("/api/v1/dashboard/events", logRequest(apiServer.HandleGetEvents)) // Zaman çizelgesi

	// 4. Sunucuyu ayağa kaldır
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
