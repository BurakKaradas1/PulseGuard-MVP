package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
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

// GetHostByID retrieves all details of a specific host from SQLite
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

// NewSQLiteRepository initializes the SQLite database and creates tables if they don't exist
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
	
	-- UNIQUE constraint applied via PRIMARY KEY to prevent Race Conditions
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

// GetHosts retrieves all registered hosts from the database
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

	// Final security check requested by linter
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// SaveEvent inserts a new agent event into the database
func (s *SQLiteRepository) SaveEvent(level, message string, passed bool) error {
	query := "INSERT INTO events(level, message, passed) VALUES (?, ?, ?)"
	_, err := s.db.Exec(query, level, message, passed)
	return err
}

// GetEvents retrieves events based on optional level and time range filters
func (r *SQLiteRepository) GetEvents(levelFilter string, timeRange string) ([]Event, error) {
	query := "SELECT level, message FROM events WHERE 1=1"
	var args []interface{}

	if levelFilter != "" {
		query += " AND level = ?"
		args = append(args, levelFilter)
	}

	// Add time filter if a range is provided
	if timeRange != "" {
		duration, err := time.ParseDuration(timeRange)
		if err == nil {
			// Find the cutoff point by subtracting the duration from the current time
			cutoffTime := time.Now().Add(-duration).UTC()
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
		err := rows.Scan(&e.Level, &e.Message)
		if err != nil {
			return nil, err
		}
		e.Passed = true // Set default value to prevent struct errors
		events = append(events, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// UpdateHostMetricsFromLog dynamically updates host metrics based on parsed log data
func (r *SQLiteRepository) UpdateHostMetricsFromLog(hostname string, cpu, ram, disk *int) error {
	// Eğer güncellenecek hiçbir metrik yoksa boşa işlem yapma
	if cpu == nil && ram == nil && disk == nil {
		return nil
	}

	// Squirrel ile UPDATE sorgusunu inşa etmeye başlıyoruz
	builder := squirrel.Update("hosts")

	if cpu != nil {
		builder = builder.Set("cpu_usage", *cpu)
	}
	if ram != nil {
		builder = builder.Set("ram_usage", *ram)
	}
	if disk != nil {
		builder = builder.Set("disk_usage", *disk)
	}

	// WHERE şartlarını OR mantığıyla ekliyoruz
	builder = builder.Where(squirrel.Or{
		squirrel.Eq{"hostname": hostname},
		squirrel.Like{"id": hostname + "-agent"},
	})

	// Sorguyu ve argümanları güvenli bir şekilde derle
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %v", err)
	}

	// Derlenen sorguyu veritabanında çalıştır
	_, err = r.db.Exec(query, args...)
	return err
}

// UpdateHostThreshold permanently saves threshold settings to SQLite
func (r *SQLiteRepository) UpdateHostThreshold(hostID string, cpu int, ram int, disk int, errLimit int) error {
	query := `
		UPDATE hosts 
		SET max_cpu_usage = ?, max_ram_usage = ?, max_disk_usage = ?, error_alert_limit = ? 
		WHERE id = ?`

	_, err := r.db.Exec(query, cpu, ram, disk, errLimit, hostID)
	return err
}

// RegisterHost registers a new agent or updates the last_seen timestamp if it exists
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

// IsBatchProcessed checks if the signature exists (Kept for interface compatibility)
func (r *SQLiteRepository) IsBatchProcessed(signature string) bool {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM processed_batches WHERE signature = ?", signature).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// MarkBatchProcessed attempts to insert the signature into the database.
// If it fails due to UNIQUE constraint, we know it's a duplicate (Race Condition check).
func (r *SQLiteRepository) MarkBatchProcessed(signature string) error {
	_, err := r.db.Exec("INSERT INTO processed_batches (signature) VALUES (?)", signature)
	return err
}

// Close safely closes the database connection
func (s *SQLiteRepository) Close() {
	s.db.Close()
}
