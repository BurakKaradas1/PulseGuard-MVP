package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const secretKey = "super-secret-pulseguard-key"

// Veriyi gizli anahtarla şifrelediğimiz yer
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
func SendBatch(events []Event, url string) error {
	if len(events) == 0 {
		return nil
	}

	data, err := json.Marshal(events)
	if err != nil {
		return err
	}
	//Verinin dijital mührü
	signature := generateSignature(data, secretKey)

	//Özel HTTP başlıkları ekleyebilmek için yeni Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

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
