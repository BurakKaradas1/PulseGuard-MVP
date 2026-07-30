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

// DequeueAll diskteki tüm olayları okur
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

	// Buradaki scanner.Err() linter'ı memnun eder
	return events, scanner.Err()
}

// RemoveProcessedEvents sadece C2'ye başarıyla gönderilen logları dosyadan siler.
// Eğer o sırada yeni loglar gelmişse onları korur (Race Condition engeli).
func RemoveProcessedEvents(processedCount int) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	var remainingEvents []Event
	file, err := os.Open(queueFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			remainingEvents = append(remainingEvents, e)
		}
	}

	// Okuma bittikten sonra hata var mı diye kontrol ediyoruz
	if err := scanner.Err(); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// Eğer dosyaya yeni bir veri eklenmemişse, doğrudan dosyayı sil
	if processedCount >= len(remainingEvents) {
		return os.Remove(queueFile)
	}

	// Dosyaya yeni veri eklenmişse gönderilenleri kes at, kalanları tut
	remainingEvents = remainingEvents[processedCount:]

	// Dosyayı tamamen sıfırla ve sadece KALAN verileri yeniden yaz
	file, err = os.OpenFile(queueFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, e := range remainingEvents {
		data, _ := json.Marshal(e)
		file.Write(data)
		file.WriteString("\n")
	}

	return nil
}
