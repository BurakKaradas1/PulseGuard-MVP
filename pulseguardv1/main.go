package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net" // Eksik olan ağ paketini ekledik
)

// AES-256 Secret Key
var aesSecretKey = []byte("pulseguard-super-secret-key-32bt")

// SystemMetrics struct to hold system metrics
type SystemMetrics struct {
	TotalRAM  uint64
	FreeRAM   uint64
	RAMUsage  float64
	CPUUsage  float64
	OpenPorts []uint32 // Açık olan (Listening) portların listesini tutacak
}

// EncryptedPayload
type EncryptedPayload struct {
	Data string `json:"encrypted_data"`
}

// encryptAES
func encryptAES(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// getSystemStats function to get system metrics
func getSystemStats() SystemMetrics {
	var metrics SystemMetrics

	// Reading RAM info
	vMem, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting virtual memory info: %v", err)
	} else {
		metrics.TotalRAM = vMem.Total / 1024 / 1024    // Convert bytes to MB
		metrics.FreeRAM = vMem.Available / 1024 / 1024 // Convert bytes to MB
		metrics.RAMUsage = vMem.UsedPercent
	}

	// Reading CPU info
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		log.Printf("Error getting CPU info: %v", err)
	} else if len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	// Network Port Info (Artık fonksiyonun İÇİNDE)
	connections, err := net.Connections("tcp")
	if err != nil {
		log.Printf("Error getting network connections: %v", err)
	} else {
		for _, conn := range connections {
			if conn.Status == "LISTEN" {
				metrics.OpenPorts = append(metrics.OpenPorts, uint32(conn.Laddr.Port))
			}
		}
	}

	return metrics
}

//sendDataToC2

func sendDataToC2(metrics SystemMetrics) {
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Printf("Error marshalling metrics to JSON: %v", err)
		return
	}

	encryptedString, err := encryptAES(jsonData, aesSecretKey)
	if err != nil {
		log.Println("Encryption error:", err)
	}

	fmt.Println("\n[->] Payload Encrypted (AES-256-GCM). Sending to C2...")
	payload := EncryptedPayload{Data: encryptedString}
	payloadJSON, _ := json.Marshal(payload)

	//fmt.Println("\n[->] JSON Package Prepared. Sending to Central Server (C2)...")
	c2URL := "http://localhost:8080/receive"

	req, err := http.NewRequest("POST", c2URL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return
	}

	//Mühürleme
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "PULSE-GUARD-SECRET-999")
	client := &http.Client{}
	resp, err := client.Do(req)
	// resp, err := http.Post(c2URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending data to C2: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("[+] Data successfully sent to the center! Server Response:", resp.Status)
}
func main() {
	fmt.Println("PulseGuard OS Integration Starting...")
	fmt.Println("Scanning system metrics...")

	currentStats := getSystemStats()
	sendDataToC2(currentStats)

	fmt.Println("\n--- Real System Metrics ---")
	fmt.Printf("Total RAM: %d MB\n", currentStats.TotalRAM)
	fmt.Printf("Free RAM: %d MB\n", currentStats.FreeRAM)
	fmt.Printf("RAM Usage: %.2f%%\n", currentStats.RAMUsage)
	fmt.Printf("CPU Usage: %.2f%%\n", currentStats.CPUUsage)

	fmt.Println("\n--- Security Scan: Open Ports ---")
	fmt.Printf("Found %d open ports:\n", len(currentStats.OpenPorts))

	limit := 10
	if len(currentStats.OpenPorts) < 10 {
		limit = len(currentStats.OpenPorts)
	}

	// Döngüyü burada başlatıp sadece portları yazdırırken kullanıyoruz
	for i := 0; i < limit; i++ {
		fmt.Printf(" > Port: %d is OPEN\n", currentStats.OpenPorts[i])
	} // Döngüyü KAPANMA (}) ile bitirdik

	// Diğer işlemler döngü bittikten sonra çalışır
	if len(currentStats.OpenPorts) > 10 {
		fmt.Printf("...and %d more open ports.\n", len(currentStats.OpenPorts)-10)
	}
	fmt.Println("--------------------------------")

	sendDataToC2(currentStats)
}
