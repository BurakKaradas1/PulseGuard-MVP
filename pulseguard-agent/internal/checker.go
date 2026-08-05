package internal

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type LogLevel string

const (
	InfoLevel    LogLevel = "INFO"
	WarningLevel LogLevel = "WARNING"
	ErrorLevel   LogLevel = "ERROR"
)

type Event struct {
	Level   LogLevel
	Message string
	Passed  bool
}
type Checker interface {
	Check() Event
	Name() string
}

type SystemChecker struct{}

func (s *SystemChecker) Name() string { return "System Integrity" }
func (s *SystemChecker) Check() Event {
	return Event{Passed: true, Level: InfoLevel, Message: "System integrity is intact"}
}

type AnalysisChecker struct{} //Debugger, Analiz kontrolü

func (a *AnalysisChecker) Name() string { return "Analysis/Debugger Detection" }
func (a *AnalysisChecker) Check() Event {
	//Windows cekirdek kütüphanesini yükle
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	//Fonksiyonu cagır
	result, _, _ := isDebuggerPresent.Call()

	if result != 0 {
		return Event{Passed: false, Level: ErrorLevel, Message: "Debugger detected in the environment"} //Hata ayıklayıcı tespit edilmis
	}
	return Event{Passed: true, Level: InfoLevel, Message: "No debugger detected"}
}

type NetworkChecker struct {
	URL string //YAML'dan gelecek hedef adres
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
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to connect to C2"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Event{Passed: true, Level: WarningLevel, Message: "C2 connection successful"}
	}

	return Event{Passed: true, Level: InfoLevel, Message: "C2 connection successful"}
}

//------------------------------------------------------------

// CPU Checker
type CpuChecker struct {
	Threshold int //YAML'dan gelecek sinir
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
		return Event{Passed: false, Level: WarningLevel, Message: message}
	}

	message := fmt.Sprintf("CPU normal: %d%%", usage)
	return Event{Passed: true, Level: InfoLevel, Message: message}
}

//-------------------------------------------------------------

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
		return Event{Passed: false, Level: WarningLevel, Message: message}
	}

	message := fmt.Sprintf("RAM normal: %d%%", usedRam)
	return Event{Passed: true, Level: InfoLevel, Message: message}
}

//--------------------------------------------------------------

// Disk Checker
type DiskChecker struct {
	Threshold int
}

func (d *DiskChecker) Name() string { return "Disk Usage" }
func (d *DiskChecker) Check() Event {
	diskStat, err := disk.Usage("C:\\")
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to read Disk metric"}
	}

	usedSpace := int(diskStat.UsedPercent)

	if usedSpace > d.Threshold {
		message := fmt.Sprintf("Disk capacity exceeded threshold: %d%%", usedSpace)
		return Event{Passed: false, Level: WarningLevel, Message: message}
	}

	message := fmt.Sprintf("Disk normal: %d%%", usedSpace)
	return Event{Passed: true, Level: InfoLevel, Message: message}
}

//---------------------------------------------------------------

// Custom Command Checker
type CustomCommandChecker struct {
	CheckName string
	Command   string
}

func (c *CustomCommandChecker) Name() string { return "c.CheckName" }
func (c *CustomCommandChecker) Check() Event {
	var err error

	if runtime.GOOS == "windows" {
		err = exec.Command("cmd", "/C", c.Command).Run()
	} else {
		err = exec.Command("sh", "-c", c.Command).Run()
	}

	if err != nil {
		return Event{
			Passed:  false,
			Level:   ErrorLevel,
			Message: fmt.Sprintf("Command execution failed: %s", c.Command),
		}
	}

	return Event{
		Passed:  true,
		Level:   InfoLevel,
		Message: fmt.Sprintf("Command executed successfully: %s", c.Command),
	}
}

//---------------------------------------------------------------

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
		return Event{
			Passed:  false,
			Level:   ErrorLevel,
			Message: fmt.Sprintf("Failed to reach endpoint: %s", h.URL),
		}
	}
	defer resp.Body.Close()

	// 200-299 arası başarı, diğerleri hata kabul edilir
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Event{
			Passed:  false,
			Level:   WarningLevel,
			Message: fmt.Sprintf("Endpoint returned HTTP %d: %s", resp.StatusCode, h.URL),
		}
	}

	return Event{
		Passed:  true,
		Level:   InfoLevel,
		Message: fmt.Sprintf("Endpoint is healthy (HTTP 200): %s", h.URL),
	}
}
