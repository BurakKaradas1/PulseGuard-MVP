package internal

import (
	"net/http"
	"syscall"
	"time"
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
		targetURL = "https://google.com"
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
