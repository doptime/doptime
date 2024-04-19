package dlog

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func getMachineName() string {
	var memStr string
	host, _ := os.Hostname()

	if memOutput, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if totalMemBytes, err := strconv.ParseUint(strings.TrimSpace(string(memOutput)), 10, 64); err == nil {
			totalMemMB := totalMemBytes / 1024 / 1024 // convert bytes to MB
			totalMemGB := totalMemMB / 1024           // convert MB to GB
			if totalMemGB >= 1 {
				memStr = fmt.Sprintf("%dGB", totalMemGB)
			} else {
				memStr = fmt.Sprintf("%dMB", totalMemMB)
			}

		}
	}

	cpuCores := runtime.NumCPU()

	machineName := fmt.Sprintf("%s-Mem%s-CPU%d", host, memStr, cpuCores)

	return machineName
}
