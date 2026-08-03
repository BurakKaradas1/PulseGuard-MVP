//go:build linux || darwin

package internal

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

type AnalysisChecker struct{}

func (a *AnalysisChecker) Name() string { return "Analysis/Debugger Detection" }
func (a *AnalysisChecker) Check() Event {
	return Event{Passed: true, Level: InfoLevel, Message: "No debugger detected (Unix bypass)", MetricName: "debugger_detected", MetricValue: 0}
}

type DiskChecker struct {
	Threshold int
}

func (d *DiskChecker) Name() string { return "Disk Usage" }
func (d *DiskChecker) Check() Event {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return Event{Passed: false, Level: ErrorLevel, Message: "Failed to read Disk metric"}
	}

	usedSpace := int(diskStat.UsedPercent)
	if usedSpace > d.Threshold {
		message := fmt.Sprintf("Disk capacity exceeded threshold: %d%%", usedSpace)
		return Event{Passed: false, Level: WarningLevel, Message: message, MetricName: "disk", MetricValue: usedSpace}
	}

	message := fmt.Sprintf("Disk normal: %d%%", usedSpace)
	return Event{Passed: true, Level: InfoLevel, Message: message, MetricName: "disk", MetricValue: usedSpace}
}
