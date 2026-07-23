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
	"regexp"
	"strconv"
	"strings"
	"time"

	"pulseguard-collector/internal/storage"
)

type ThresholdConfig struct {
	MaxCpuUsage     int `json:"max_cpu_usage"`
	MaxRamUsage     int `json:"max_ram_usage"`
	MaxDiskUsage    int `json:"max_disk_usage"`
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

// 1. Get details of a specific host (GET)
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

	hostData, err := s.repo.GetHostByID(hostID)
	if err != nil {
		http.Error(w, "Host not found or not registered yet", http.StatusNotFound)
		return
	}

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
			MaxDiskUsage:    hostData.MaxDiskUsage,
			ErrorAlertLimit: hostData.ErrorAlertLimit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// 2. Set new alarm thresholds for a host (POST)
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
		newConfig.MaxDiskUsage,
		newConfig.ErrorAlertLimit,
	)

	if err != nil {
		fmt.Printf("[!] Database Error (UpdateThreshold): %v\n", err)
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[+] Threshold updated for host %s: CPU:%d%% RAM:%d%% DISK:%d%%\n", hostID, newConfig.MaxCpuUsage, newConfig.MaxRamUsage, newConfig.MaxDiskUsage)

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

	// Idempotency & Race Condition Prevention:
	// Attempt to insert signature directly into processed_batches using DB UNIQUE constraint.
	err = s.repo.MarkBatchProcessed(agentSignature)
	if err != nil {
		fmt.Println("[-] Idempotency: This batch has already been processed (Race Condition prevented), skipping.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Batch already processed"))
		return
	}

	var events []storage.Event
	err = json.NewDecoder(r.Body).Decode(&events)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	var latestCpu, latestRam, latestDisk *int

	for _, event := range events {
		err = s.repo.SaveEvent(event.Level, event.Message, event.Passed)
		if err != nil {
			fmt.Printf("Failed to insert event: %v\n", err)
		}

		msgLower := strings.ToLower(event.Message)
		if strings.Contains(msgLower, "cpu") {
			if val, ok := parseMetricValue(event.Message); ok {
				latestCpu = &val
			}
		} else if strings.Contains(msgLower, "ram") {
			if val, ok := parseMetricValue(event.Message); ok {
				latestRam = &val
			}
		} else if strings.Contains(msgLower, "disk") {
			if val, ok := parseMetricValue(event.Message); ok {
				latestDisk = &val
			}
		}
	}

	if latestCpu != nil || latestRam != nil || latestDisk != nil {
		hosts, _ := s.repo.GetHosts()
		for _, h := range hosts {
			_ = s.repo.UpdateHostMetricsFromLog(h.Hostname, latestCpu, latestRam, latestDisk)
		}
	}

	fmt.Printf("[+] Successfully verified and stored %d events from agent and updated metrics.\n", len(events))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Batch processed successfully"))
}

type HostStatus struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
	CpuUsage  int       `json:"cpu_usage"`
	RamUsage  int       `json:"ram_usage"`
	DiskUsage int       `json:"disk_usage"`
}

// Fleet view
func (s *APIServer) HandleGetHosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbHosts, err := s.repo.GetHosts()
	if err != nil {
		http.Error(w, "Failed to read hosts from database", http.StatusInternalServerError)
		return
	}

	var hosts []HostStatus
	for _, h := range dbHosts {
		hosts = append(hosts, HostStatus{
			ID:        h.ID,
			Hostname:  h.Hostname,
			Status:    "healthy",
			LastSeen:  h.LastSeen,
			CpuUsage:  h.CpuUsage,
			RamUsage:  h.RamUsage,
			DiskUsage: h.DiskUsage,
		})
	}

	if hosts == nil {
		hosts = []HostStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

// Returns filtered events for time and reports
func (s *APIServer) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	levelFilter := r.URL.Query().Get("level")
	timeRange := r.URL.Query().Get("range")
	events, err := s.repo.GetEvents(levelFilter, timeRange)
	if err != nil {
		fmt.Printf("[!] Database Error (GetEvents): %v\n", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// 3. Registers the agent when it first connects (POST)
func (s *APIServer) HandleRegisterHost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		IP       string `json:"ip_address"`
		OS       string `json:"os"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := s.repo.RegisterHost(req.ID, req.Hostname, req.IP, req.OS)
	if err != nil {
		fmt.Printf("[!] Database Error (RegisterHost): %v\n", err)
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[+] NEW AGENT REGISTERED: %s (%s)\n", req.Hostname, req.IP)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Host registered successfully to PulseGuard"))
}

func parseMetricValue(message string) (int, bool) {
	re := regexp.MustCompile(`(\d+)%`)
	match := re.FindStringSubmatch(message)
	if len(match) > 1 {
		val, err := strconv.Atoi(match[1])
		if err == nil {
			return val, true
		}
	}
	return 0, false
}
