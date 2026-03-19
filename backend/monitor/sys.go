package monitor

import (
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SysStatus struct {
	CPUPercent float64 `json:"cpu_percent"`
	MemTotal   uint64  `json:"mem_total"`
	MemUsed    uint64  `json:"mem_used"`
	MemPercent float64 `json:"mem_percent"`
}

func GetSysStatus(client *ssh.Client) SysStatus {
	var status SysStatus

	if client == nil {
		return status
	}

	// vmstat 1 2 runs for 1 sec and prints 2 lines. We take the last line, 15th field (id), CPU used = 100 - id
	cpuCmd := "vmstat 1 2 | tail -1 | awk '{print 100-$15}'"
	cpuOut, err := RunCommand(client, cpuCmd)
	if err == nil {
		if cpuVal, err2 := strconv.ParseFloat(strings.TrimSpace(cpuOut), 64); err2 == nil {
			status.CPUPercent = cpuVal
		}
	}

	// free -b outputs in bytes. Total and Used are the 2nd and 3rd fields of the 2nd line
	memCmd := "free -b | awk 'NR==2{print $2,$3}'"
	memOut, err := RunCommand(client, memCmd)
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(memOut))
		if len(parts) >= 2 {
			total, _ := strconv.ParseUint(parts[0], 10, 64)
			used, _ := strconv.ParseUint(parts[1], 10, 64)
			status.MemTotal = total
			status.MemUsed = used
			if total > 0 {
				status.MemPercent = float64(used) / float64(total) * 100.0
			}
		}
	}

	return status
}
