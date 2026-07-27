package internal

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

const queueFile = "pulseguard_offline.wal"

var queueMutex sync.Mutex

// EnqueueEvent fonksiyonu yeni olayı WAL dosyasına yazar
func EnqueueEvent(e Event) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	file, err := os.OpenFile(queueFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	file.Write(data)
	file.WriteString("\n")
	return nil
}

func DequeueAll() ([]Event, error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	var events []Event
	file, err := os.Open(queueFile)
	if err != nil {
		if os.IsNotExist(err) {
			return events, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			events = append(events, e)
		}
	}
	return events, scanner.Err()
}

func ClearQueue() error {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return os.Remove(queueFile)
}
