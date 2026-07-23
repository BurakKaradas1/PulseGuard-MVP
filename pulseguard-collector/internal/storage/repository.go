package storage

import "time"

type Event struct {
	ID      int    `json:"id"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Passed  bool   `json:"passed"`
}

type Host struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	LastSeen  time.Time `json:"last_seen"`
	CpuUsage  int       `json:"cpu_usage"`  // YENİ
	RamUsage  int       `json:"ram_usage"`  // YENİ
	DiskUsage int       `json:"disk_usage"` // YENİ
}

type Repository interface {
	//Ajanlardan gelen yeni olayları kaydeder
	SaveEvent(level, message string, passed bool) error

	//Dashboard için filreleme
	GetEvents(levelFilter string, timeRange string) ([]Event, error)

	GetHosts() ([]Host, error)

	UpdateHostThreshold(hostID string, cpu int, ram int, disk int, errLimit int) error

	RegisterHost(hostID, hostname, ip, os string) error

	GetHostByID(hostID string) (HostFullDetail, error)

	IsBatchProcessed(signature string) bool

	MarkBatchProcessed(signature string) error

	UpdateHostMetricsFromLog(hostname string, cpu, ram, disk *int) error
}
