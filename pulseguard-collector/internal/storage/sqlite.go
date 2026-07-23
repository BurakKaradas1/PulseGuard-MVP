package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type HostFullDetail struct {
	ID              string
	Hostname        string
	IPAddress       string
	OS              string
	Status          string
	LastSeen        time.Time
	MaxCpuUsage     int
	MaxRamUsage     int
	MaxDiskUsage    int
	ErrorAlertLimit int
}
type SQLiteRepository struct {
	db *sql.DB
}

// ID'si verilen hostun tüm bilgilerini SQLite'tan getirir
func (r *SQLiteRepository) GetHostByID(hostID string) (HostFullDetail, error) {
	var h HostFullDetail
	query := `
        SELECT id, hostname, ip_address, os, status, last_seen, max_cpu_usage, max_ram_usage, max_disk_usage, error_alert_limit 
        FROM hosts 
        WHERE id = ?`

	err := r.db.QueryRow(query, hostID).Scan(
		&h.ID, &h.Hostname, &h.IPAddress, &h.OS, &h.Status, &h.LastSeen,
		&h.MaxCpuUsage, &h.MaxRamUsage, &h.MaxDiskUsage, &h.ErrorAlertLimit,
	)
	return h, err
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
    CREATE TABLE IF NOT EXISTS hosts (
        id TEXT PRIMARY KEY,
        hostname TEXT,
        ip_address TEXT,
        os TEXT,
        status TEXT,
        last_seen DATETIME,
        max_cpu_usage INTEGER,
        max_ram_usage INTEGER,
        max_disk_usage INTEGER,
        error_alert_limit INTEGER,
        cpu_usage INTEGER DEFAULT 0,
        ram_usage INTEGER DEFAULT 0,
        disk_usage INTEGER DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        host_id TEXT,
        level TEXT,
        message TEXT,
        passed BOOLEAN,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(host_id) REFERENCES hosts(id)
    );

    CREATE TABLE IF NOT EXISTS thresholds (
        level TEXT PRIMARY KEY,
        enabled BOOLEAN
    );
    
    CREATE TABLE IF NOT EXISTS processed_batches (
        signature TEXT PRIMARY KEY,
        processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	fmt.Println("[+] SQLite Database initialized successfully.")
	return &SQLiteRepository{db: db}, nil
}

// Veritabanındaki tüm hostları alır
func (r *SQLiteRepository) GetHosts() ([]Host, error) {
	query := "SELECT id, hostname, last_seen, cpu_usage, ram_usage, disk_usage FROM hosts"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		err := rows.Scan(&h.ID, &h.Hostname, &h.LastSeen, &h.CpuUsage, &h.RamUsage, &h.DiskUsage)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}

	// Linter'ın istediği son güvenlik kontrolü
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

func (s *SQLiteRepository) SaveEvent(level, message string, passed bool) error {
	query := "INSERT INTO events(level, message, passed) VALUES (?, ?, ?)"
	_, err := s.db.Exec(query, level, message, passed)
	return err
}

func (r *SQLiteRepository) GetEvents(levelFilter string, timeRange string) ([]Event, error) {
	query := "SELECT level, message FROM events WHERE 1=1"
	var args []interface{}

	if levelFilter != "" {
		query += " AND level = ?"
		args = append(args, levelFilter)
	}
	// Eğer zaman filtresi varsa sorguya ekle
	if timeRange != "" {
		duration, err := time.ParseDuration(timeRange)
		if err == nil {
			// Şu anki zamandan istenilen süreyi çıkararak bir başlangıç noktası bul
			cutoffTime := time.Now().Add(-duration).UTC()
			// Log kayıt tarihi bu başlangıç noktasından büyük/eşit olanlar
			query += " AND created_at >= ?"
			args = append(args, cutoffTime)
		}
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		// Sadece yeni şemada olan alanları okuyoruz
		err := rows.Scan(&e.Level, &e.Message)
		if err != nil {
			return nil, err
		}
		e.Passed = true // Struct hata vermesin diye varsayılan değer atıyoruz
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *SQLiteRepository) UpdateHostMetricsFromLog(hostname string, cpu, ram, disk *int) error {
	// Dinamik olarak hangi metrik geldiyse sadece onu güncelleyen esnek bir SQL yapıcısı
	var setClauses []string
	var args []interface{}

	if cpu != nil {
		setClauses = append(setClauses, "cpu_usage = ?")
		args = append(args, *cpu)
	}
	if ram != nil {
		setClauses = append(setClauses, "ram_usage = ?")
		args = append(args, *ram)
	}
	if disk != nil {
		setClauses = append(setClauses, "disk_usage = ?")
		args = append(args, *disk)
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, hostname)
	query := fmt.Sprintf("UPDATE hosts SET %s WHERE hostname = ? OR id LIKE ?", strings.Join(setClauses, ", "))
	// Ajan ID'si genellikle hostname + "-agent" şeklinde olduğu için her iki ihtimale de bakıyoruz
	args = append(args, hostname+"-agent")

	_, err := r.db.Exec(query, args...)
	return err
}

// Eşik ayarlarını SQLite'a kalıcı olarak kaydeder
func (r *SQLiteRepository) UpdateHostThreshold(hostID string, cpu int, ram int, disk int, errLimit int) error {
	query := `
        UPDATE hosts 
        SET max_cpu_usage = ?, max_ram_usage = ?, max_disk_usage = ?, error_alert_limit = ? 
        WHERE id = ?`

	_, err := r.db.Exec(query, cpu, ram, disk, errLimit, hostID)
	return err
}

// Ajan ilk bağlandığında veya var olan ajan veri gönderdiğinde onu sisteme kaydeder
func (r *SQLiteRepository) RegisterHost(hostID, hostname, ip, os string) error {
	query := `
		INSERT INTO hosts (id, hostname, ip_address, os, status, last_seen) 
		VALUES (?, ?, ?, ?, 'healthy', CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET 
			hostname = excluded.hostname,
			ip_address = excluded.ip_address,
			os = excluded.os,
			last_seen = CURRENT_TIMESTAMP;`

	_, err := r.db.Exec(query, hostID, hostname, ip, os)
	return err
}

func (r *SQLiteRepository) IsBatchProcessed(signature string) bool {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM processed_batches WHERE signature = ?", signature).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (r *SQLiteRepository) MarkBatchProcessed(signature string) error {
	_, err := r.db.Exec("INSERT INTO processed_batches (signature) VALUES (?)", signature)
	return err
}
func (s *SQLiteRepository) Close() {
	s.db.Close()
}
