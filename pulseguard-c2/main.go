package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Ajan ile tamamen aynı gizli anahtara sahip olmalıyız (32 byte)
var aesSecretKey = []byte("pulseguard-super-secret-key-32bt")

// Gelen şifreli paketi alacağımız yapı
type EncryptedPayload struct {
	Data string `json:"encrypted_data"`
}

// Şifre çözüldükten sonra verileri yerleştireceğimiz yapı
type SystemMetrics struct {
	TotalRAM  uint64   `json:"TotalRAM"`
	FreeRAM   uint64   `json:"FreeRAM"`
	RAMUsage  float64  `json:"RAMUsage"`
	CPUUsage  float64  `json:"CPUUsage"`
	OpenPorts []uint32 `json:"OpenPorts"`
}

// Python dashboarduna veri sunabilmek için gelen son veriyi RAM'de saklıyoruz
var (
	lastMetrics SystemMetrics
	metricsLock sync.Mutex
)

// AES-256-GCM Şifre Çözme (Decryption) Fonksiyonu
func decryptAES(encryptedBase64 string, key []byte) ([]byte, error) {
	// Base64 metnini tekrar ham baytlara çevir
	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Nonce (ilk kısım) ve asıl şifreli veriyi (ikinci kısım) ayır
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Kilidi aç!
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// handleReceiveData intercepts the incoming POST requests from agents
func handleReceiveData(w http.ResponseWriter, r *http.Request) {
	// 1. API Key Kontrolü (Güvenlik Kalkanı)
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "PULSE-GUARD-SECRET-999" {
		http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized)
		return
	}

	// 2. Gelen şifreli paketi al
	var payload EncryptedPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid Payload", http.StatusBadRequest)
		return
	}

	// 3. Şifreyi Çöz!
	decryptedJSON, err := decryptAES(payload.Data, aesSecretKey)
	if err != nil {
		fmt.Println("[!] Decryption Failed! Potential tampering detected.")
		http.Error(w, "Decryption Failed", http.StatusBadRequest)
		return
	}

	// 4. Çözülen saf JSON'ı struct'a dönüştür ('err' zaten var olduğu için sadece = kullanıyoruz)
	var metrics SystemMetrics
	err = json.Unmarshal(decryptedJSON, &metrics)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	metricsLock.Lock()
	lastMetrics = metrics
	metricsLock.Unlock()

	// 5. Şifresi çözülen veriyi ekrana bas
	fmt.Println("\n[+] Secure Intelligence Payload Decrypted & Received!")
	fmt.Printf(" -> Target RAM Usage : %% %.2f\n", metrics.RAMUsage)
	fmt.Printf(" -> Target CPU Usage : %% %.2f\n", metrics.CPUUsage)
	fmt.Printf(" -> Open Ports Count : %d\n", len(metrics.OpenPorts))
	fmt.Println("------------------------------------------------")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Secure Payload received")
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastMetrics)
}

func main() {
	fmt.Println("PulseGuard SECURE C2 Server Starting...")
	fmt.Println("Listening on port 8080 (AES-256 Active)...")

	// Route the "/receive" endpoint to our handler function
	http.HandleFunc("/receive", handleReceiveData)
	http.HandleFunc("/stats", handleStats) //Python dashboard buradan okuyacak

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
