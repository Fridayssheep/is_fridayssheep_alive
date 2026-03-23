package monitor

import (
	"frisheep-alive-backend/logger"
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

	// 使用 bash -c '...' 来确保命令能在远程的正确 PATH 下运行
	// 通过 tail -1 来规避可能混入 bashrc 的启动欢迎图案
	cpuCmd := "bash -c \"vmstat 1 2 | tail -1 | awk '{print 100-\\$15}'\""
	cpuOut, err := RunCommand(client, cpuCmd)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(cpuOut), "\n")
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		if cpuVal, err2 := strconv.ParseFloat(lastLine, 64); err2 == nil {
			status.CPUPercent = cpuVal
		}
	} else {
		logger.Warnf("Get CPU error: %v", err)
	}

	memCmd := "bash -c \"free -b | awk 'NR==2{print \\$2,\\$3}'\""
	memOut, err := RunCommand(client, memCmd)
	if err != nil {
		logger.Warnf("Error fetching memory: %v", err)
	} else {
		lines := strings.Split(strings.TrimSpace(memOut), "\n")
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		parts := strings.Fields(lastLine)
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
