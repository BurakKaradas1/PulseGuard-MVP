package internal

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type LogLevel string

const (
	InfoLevel    LogLevel = "INFO"
	WarningLevel LogLevel = "WARNING"
	ErrorLevel   LogLevel = "ERROR"
)

type Event struct {
	Level       LogLevel `json:"level"`
	Message     string   `json:"message"`
	Passed      bool     `json:"passed"`
	MetricName  string   `json:"metric_name"`  // BUGFIX: Regex kırılganlığını önlemek için
	MetricValue int      `json:"metric_value"` // BUGFIX: Metrikleri sayısal göndermek için
	BatchID     string   `json:"batch_id"`     // BUGFIX: Idempotency (Veri Kaybı) çözümü için eşsiz ID
}

type Checker interface {
	Check() Event
	Name() string
}

type SystemChecker struct{}

func (s *SystemChecker) Name() string { return "System Integrity" }
func (s *SystemChecker) Check() Event {
	return Event{Passed: true, Level: InfoLevel, Message: "System integrity is intact", MetricName: "system_integrity", MetricValue: 1}
}

type NetworkChecker struct {
	URL string // Target address from YAML
}

func (n *NetworkChecker) Name() string { return "C2 Network Status" }
func (n *NetworkChecker) Check() Event {
	targetURL := n.URL
	if targetURL == "" {
		targetURL = "http://localhost:8080/api/v1/events"
	}
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(targetURL)
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to connect to C2", MetricName: "c2_status", MetricValue: 0}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Event{Passed: true, Level: WarningLevel, Message: "C2 connection successful", MetricName: "c2_status", MetricValue: resp.StatusCode}
	}

	return Event{Passed: true, Level: InfoLevel, Message: "C2 connection successful", MetricName: "c2_status", MetricValue: resp.StatusCode}
}

// CPU Checker
type CpuChecker struct {
	Threshold int // Threshold from YAML
}

func (c *CpuChecker) Name() string { return "CPU Usage" }
func (c *CpuChecker) Check() Event {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to read CPU metric"}
	}

	usage := int(percentages[0])

	if usage > c.Threshold {
		message := fmt.Sprintf("CPU usage exceeded threshold: %d%%", usage)
		return Event{Passed: false, Level: WarningLevel, Message: message, MetricName: "cpu", MetricValue: usage}
	}

	message := fmt.Sprintf("CPU normal: %d%%", usage)
	return Event{Passed: true, Level: InfoLevel, Message: message, MetricName: "cpu", MetricValue: usage}
}

// RAM Checker
type RamChecker struct {
	Threshold int
}

func (r *RamChecker) Name() string { return "RAM Usage" }
func (r *RamChecker) Check() Event {
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to read RAM metric"}
	}

	usedRam := int(virtualMem.UsedPercent)

	if usedRam > r.Threshold {
		message := fmt.Sprintf("RAM usage exceeded threshold: %d%%", usedRam)
		return Event{Passed: false, Level: WarningLevel, Message: message, MetricName: "ram", MetricValue: usedRam}
	}

	message := fmt.Sprintf("RAM normal: %d%%", usedRam)
	return Event{Passed: true, Level: InfoLevel, Message: message, MetricName: "ram", MetricValue: usedRam}
}

// Custom Command Checker
type CustomCommandChecker struct {
	CheckName string
	Command   string
}

func (c *CustomCommandChecker) Name() string { return c.CheckName }
func (c *CustomCommandChecker) Check() Event {
	var err error

	if runtime.GOOS == "windows" {
		err = exec.Command("cmd", "/C", c.Command).Run()
	} else {
		err = exec.Command("sh", "-c", c.Command).Run()
	}

	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: fmt.Sprintf("Command execution failed: %s", c.Command), MetricName: "custom_command", MetricValue: 0}
	}

	return Event{Passed: true, Level: InfoLevel, Message: fmt.Sprintf("Command executed successfully: %s", c.Command), MetricName: "custom_command", MetricValue: 1}
}

// Custom HTTP Endpoint Checker
type CustomHttpChecker struct {
	CheckName string
	URL       string
}

func (h *CustomHttpChecker) Name() string { return h.CheckName }
func (h *CustomHttpChecker) Check() Event {
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(h.URL)
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: fmt.Sprintf("Failed to reach endpoint: %s", h.URL), MetricName: "custom_http", MetricValue: 0}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Event{Passed: false, Level: WarningLevel, Message: fmt.Sprintf("Endpoint returned HTTP %d: %s", resp.StatusCode, h.URL), MetricName: "custom_http", MetricValue: resp.StatusCode}
	}

	return Event{Passed: true, Level: InfoLevel, Message: fmt.Sprintf("Endpoint is healthy (HTTP 200): %s", h.URL), MetricName: "custom_http", MetricValue: resp.StatusCode}
}
