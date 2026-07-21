package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
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
		last_seen DATETIME
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host_id TEXT,
		level TEXT,
		message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(host_id) REFERENCES hosts(id)
	);

	CREATE TABLE IF NOT EXISTS thresholds (
		level TEXT PRIMARY KEY,
		enabled BOOLEAN
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
	rows, err := r.db.Query("SELECT id, hostname, last_seen FROM hosts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		err := rows.Scan(&h.ID, &h.Hostname, &h.LastSeen)
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

func (r *SQLiteRepository) GetEvents(levelFilter string) ([]Event, error) {
	query := "SELECT level, message FROM events"
	var args []interface{}

	if levelFilter != "" {
		query += " WHERE level = ?"
		args = append(args, levelFilter)
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

func (s *SQLiteRepository) Close() {
	s.db.Close()
}
