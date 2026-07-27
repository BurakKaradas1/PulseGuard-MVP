package main

import (
	"fmt"
	"log"
	"net/http"

	"pulseguard-collector/internal/api"
	"pulseguard-collector/internal/config"
	"pulseguard-collector/internal/storage"
)

// logRequest middleware logs incoming requests to the terminal
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] Incoming Request: %s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

// CORS Middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Can be left as "*" for development or set to a specific React port.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-PulseGuard-Signature")

		// Return 200 OK directly for Preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// 1. Load configuration (Fails safely if the environment variable is missing)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("[-] Failed to start server: ", err)
	}

	// 2. Initialize database layer
	repo, err := storage.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatal("[-] Database connection error: ", err)
	}

	fmt.Println("[+] SQLite Database initialized successfully.")
	fmt.Println("[+] PulseGuard C2 Server (Collector) Starting...")

	// 3. Create API Server instance (Inject Interface and SecretKey)
	apiServer := api.NewAPIServer(repo, cfg.SecretKey)

	// 4. REST API Endpoints
	http.HandleFunc("/api/v1/events", corsMiddleware(logRequest(apiServer.HandleReceiveEvents)))
	http.HandleFunc("/api/v1/dashboard/hosts", corsMiddleware(logRequest(apiServer.HandleGetHosts)))
	http.HandleFunc("/api/v1/dashboard/events", corsMiddleware(logRequest(apiServer.HandleGetEvents)))
	http.HandleFunc("/api/v1/dashboard/hosts/detail", corsMiddleware(logRequest(apiServer.HandleGetHostDetail)))
	http.HandleFunc("/api/v1/dashboard/hosts/threshold", corsMiddleware(logRequest(apiServer.HandleSetThreshold)))
	http.HandleFunc("/api/v1/agent/register", corsMiddleware(logRequest(apiServer.HandleRegisterHost)))

	// 5. Start the server
	fmt.Printf("[+] Listening on port %s (SQLite active...)\n", cfg.Port)
	err = http.ListenAndServe(cfg.Port, nil)
	if err != nil {
		log.Fatal("[-] Server failed: ", err)
	}
}
