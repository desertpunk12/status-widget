package widget

import (
	"fmt"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/process"
)

// updateSystemStats updates CPU and memory usage of this widget
func (w *Widget) updateSystemStats() {
	// Get current process
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return // Skip update if process lookup fails
	}

	// Memory: RSS (Resident Set Size) - matches Task Manager's Working Set better
	// RSS includes physical pages allocated, Working Set = RSS + shared
	memInfo, err := p.MemoryInfo()
	if err == nil {
		// RSS is in bytes, convert to MB
		w.memUsageMB = float64(memInfo.RSS) / 1024 / 1024
	}

	// CPU: Get process CPU percentage
	cpuPercent, err := p.CPUPercent()
	if err == nil {
		w.cpuUsage = cpuPercent / float64(runtime.NumCPU())
	}
}

// formatCPUText formats CPU usage into a display string
func (w *Widget) formatCPUText() string {
	return fmt.Sprintf("%.1f%%", w.cpuUsage)
}

// formatMEMText formats memory usage into a display string
func (w *Widget) formatMEMText() string {
	return fmt.Sprintf("%.1f MB", w.memUsageMB)
}
