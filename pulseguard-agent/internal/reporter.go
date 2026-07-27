package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Veriyi gizli anahtarla şifrelediğimiz yer
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// Ajan kayıt paket yapısı
type RegisterPayload struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip_address"`
	OS       string `json:"os"`
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// Ajanın Collector'e kendini kaydettirmesi
func RegisterAgent(collectorURL string) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	payload := RegisterPayload{
		ID:       hostname + "-agent",
		Hostname: hostname,
		IP:       getLocalIP(),
		OS:       runtime.GOOS,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON oluşturma hatası %v", err)
	}

	registerEndpoint := fmt.Sprintf("%s/api/v1/agent/register", collectorURL)
	resp, err := http.Post(registerEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Collector'e ulaşılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Kayıt reddedildi, Durum Kodu: %d", resp.StatusCode)
	}

	fmt.Printf("[+] Ajan başarıyla Collector'e kaydedildi! (Host: %s, IP: %s)\n", payload.Hostname, payload.IP)
	return nil
}

func SendBatch(events []Event, fullURL string) error {
	if len(events) == 0 {
		return nil
	}

	// 1. Ortam değişkeninden gizli anahtarı oku (İngilizce hata mesajı ile)
	secretKey := os.Getenv("PULSEGUARD_SECRET")
	if secretKey == "" {
		return fmt.Errorf("PULSEGUARD_SECRET environment variable is not defined, data transmission aborted")
	}

	data, err := json.Marshal(events)
	if err != nil {
		return err
	}

	// 2. Şifreleme işlemini dinamik anahtarla yap
	signature := generateSignature(data, secretKey)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PulseGuard-Signature", signature)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server error, status code: %d", resp.StatusCode)
	}

	return nil
}
