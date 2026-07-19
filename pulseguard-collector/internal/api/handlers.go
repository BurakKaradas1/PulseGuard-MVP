package api

import (
	"encoding/json"
	"net/http"
	"time"

	"pulseguard-collector/internal/storage"
)

type APIServer struct {
	repo storage.Repository
}

func NewAPIServer(repo storage.Repository) *APIServer {
	return &APIServer{
		repo: repo,
	}
}

type HostStatus struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// Filo görünümü
func (s *APIServer) HandleGetHosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mockHosts := []HostStatus{
		{
			ID:       "host-001",
			Hostname: "db-server-main",
			Status:   "healthy", // Renders Green in React
			LastSeen: time.Now(),
		},
		{
			ID:       "host-002",
			Hostname: "web-node-01",
			Status:   "critical", // Renders Red in React
			LastSeen: time.Now().Add(-15 * time.Minute),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockHosts)
}

// Zaman ve raporlar için filtrelenmiş etkinlikleri döndürür
func (s *APIServer) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read filters from the URL query parameters
	levelFilter := r.URL.Query().Get("level") // e.g., "ERROR", "WARNING"

	// Real implementation will pass levelFilter to the database query
	events, err := s.repo.GetEvents(levelFilter)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
