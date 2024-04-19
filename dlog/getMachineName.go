package dlog

import (
	"fmt"
	"os"
	"runtime"
)

func getMachineName() string {
	host, _ := os.Hostname()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)

	totalMemMB := mem.TotalAlloc / 1024 / 1024 // convert bytes to MB
	totalMemGB := totalMemMB / 1024            // convert MB to GB

	var memStr string
	if totalMemGB >= 1 {
		memStr = fmt.Sprintf("%dGB", totalMemGB)
	} else {
		memStr = fmt.Sprintf("%dMB", totalMemMB)
	}

	cpuCores := runtime.NumCPU()

	machineName := fmt.Sprintf("%s-Mem%s-CPU%d", host, memStr, cpuCores)

	return machineName
}
