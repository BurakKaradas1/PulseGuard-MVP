package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"pulseguard-collector/internal/storage"
)

type ThresholdConfig struct {
	MaxCpuUsage     int `json:"max_cpu_usage"`
	MaxRamUsage     int `json:"max_ram_usage"`
	ErrorAlertLimit int `json:"error_alert_limit"`
}

type HostDetail struct {
	HostStatus
	IPAddress string          `json:"ip_address"`
	OS        string          `json:"os"`
	Threshold ThresholdConfig `json:"threshold"`
}

type APIServer struct {
	repo      storage.Repository
	secretKey string
}

// 1. Belirli bir hostun detaylarını getirir (GET)
func (s *APIServer) HandleGetHostDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostID := r.URL.Query().Get("id")
	if hostID == "" {
		http.Error(w, "Missing host id", http.StatusBadRequest)
		return
	}

	// Veritabanından gerçek veriyi çek
	hostData, err := s.repo.GetHostByID(hostID)
	if err != nil {
		http.Error(w, "Host bulunamadı veya henüz sisteme kayıt olmadı", http.StatusNotFound)
		return
	}

	// Gelen ham veriyi JSON formatına oturt
	detail := HostDetail{
		HostStatus: HostStatus{
			ID:       hostData.ID,
			Hostname: hostData.Hostname,
			Status:   hostData.Status,
			LastSeen: hostData.LastSeen,
		},
		IPAddress: hostData.IPAddress,
		OS:        hostData.OS,
		Threshold: ThresholdConfig{
			MaxCpuUsage:     hostData.MaxCpuUsage,
			MaxRamUsage:     hostData.MaxRamUsage,
			ErrorAlertLimit: hostData.ErrorAlertLimit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// 2. Bir host için yeni alarm eşikleri (Threshold) belirler (POST)
func (s *APIServer) HandleSetThreshold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostID := r.URL.Query().Get("id")
	if hostID == "" {
		http.Error(w, "Missing host id", http.StatusBadRequest)
		return
	}

	// React'ten gelen yeni ayarları JSON olarak oku
	var newConfig ThresholdConfig
	err := json.NewDecoder(r.Body).Decode(&newConfig)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	err = s.repo.UpdateHostThreshold(
		hostID,
		newConfig.MaxCpuUsage,
		newConfig.MaxRamUsage,
		newConfig.ErrorAlertLimit,
	)

	if err != nil {
		fmt.Printf("[!] Veritabanı Hatası (UpdateThreshold): %v\n", err)
		http.Error(w, "Veritabanına kaydedilemedi", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[+] Threshold updated for host %s: CPU:%d%% RAM:%d%%\n", hostID, newConfig.MaxCpuUsage, newConfig.MaxRamUsage)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Threshold configuration saved successfully"))
}

func NewAPIServer(repo storage.Repository, secretKey string) *APIServer {
	return &APIServer{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (s *APIServer) verifySignature(payload []byte, signature string) bool {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func (s *APIServer) HandleReceiveEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentSignature := r.Header.Get("X-PulseGuard-Signature")
	if agentSignature == "" {
		http.Error(w, "Missing Signature", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if !s.verifySignature(bodyBytes, agentSignature) {
		fmt.Println("[!] SECURITY ALERT: Invalid signature detected! Potential tampering")
		http.Error(w, "Invalid Signature", http.StatusUnauthorized)
		return
	}

	//Eğer bu imza daha önce geldiyse, paketi tekrar işleme diyoruz
	if s.repo.IsBatchProcessed(agentSignature) {
		fmt.Println("[-] Idempotency: Bu paket daha önce işlendi, atlanıyor.")
		w.WriteHeader(http.StatusOK) // Ajana "tamam aldım" diyoruz ki tekrar göndermeye çalışmasın
		w.Write([]byte("Batch already processed"))
		return
	}

	var events []storage.Event
	err = json.NewDecoder(r.Body).Decode(&events)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	for _, event := range events {
		// Global repo yerine yapı içindeki s.repo arayüzünü kullanıyoruz
		err = s.repo.SaveEvent(event.Level, event.Message, event.Passed)
		if err != nil {
			fmt.Printf("Failed to insert event: %v\n", err)
		}
	}

	s.repo.MarkBatchProcessed(agentSignature)

	fmt.Printf("[+] Successfully verified and stored %d events from agent.\n", len(events))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Batch processed successfully"))
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

	dbHosts, err := s.repo.GetHosts()
	if err != nil {
		http.Error(w, "Veritabanından hostlar okunamadı", http.StatusInternalServerError)
		return
	}

	var hosts []HostStatus
	for _, h := range dbHosts {
		hosts = append(hosts, HostStatus{
			ID:       h.ID,
			Hostname: h.Hostname,
			Status:   "healthy", // Şimdilik varsayılan atıyoruz, ileride eklenebilir
			LastSeen: h.LastSeen,
		})
	}

	if hosts == nil {
		hosts = []HostStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

// Zaman ve raporlar için filtrelenmiş etkinlikleri döndürür
func (s *APIServer) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read filters from the URL query parameters
	levelFilter := r.URL.Query().Get("level") // "ERROR", "WARNING"
	timeRange := r.URL.Query().Get("range")
	// Real implementation will pass levelFilter to the database query
	events, err := s.repo.GetEvents(levelFilter, timeRange)
	if err != nil {
		fmt.Printf("[!] Veritabanı Hatası (GetEvents): %v\n", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// 3. Ajanın sisteme ilk bağlandığında kendini kaydetmesini sağlar (POST)
func (s *APIServer) HandleRegisterHost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ajandan gelecek olan JSON paketinin yapısı
	var req struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		IP       string `json:"ip_address"`
		OS       string `json:"os"`
	}

	// Gelen veriyi oku
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Veritabanına kaydet (Geçmişte storage tarafında yazdığımız fonksiyonu çağırıyoruz)
	err := s.repo.RegisterHost(req.ID, req.Hostname, req.IP, req.OS)
	if err != nil {
		fmt.Printf("[!] Veritabanı Hatası (RegisterHost): %v\n", err)
		http.Error(w, "Veritabanına kaydedilemedi", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[+] YENİ AJAN KAYDEDİLDİ: %s (%s)\n", req.Hostname, req.IP)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Host registered successfully to PulseGuard"))
}
