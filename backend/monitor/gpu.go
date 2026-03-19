package monitor

import (
	"strings"

	"golang.org/x/crypto/ssh"
)

type GPUStatus struct {
	Utilization string `json:"utilization"`
	MemoryUsed  string `json:"memory_used"`
	MemoryTotal string `json:"memory_total"`
}

func GetGPUStatus(client *ssh.Client) []GPUStatus {
	var gpus []GPUStatus
	if client == nil {
		return gpus
	}

	cmd := "bash -c 'PATH=$PATH:/usr/bin:/usr/sbin:/sbin:/bin nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total --format=csv,noheader 2>/dev/null'"
	out, err := RunCommand(client, cmd)
	if err != nil {
		return gpus
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 如果这一行包含猫图案或者各种无关字符，忽略。正常输出是类似 `15 %, 4096 MiB, 24576 MiB`
		if line == "" || !strings.Contains(line, "MiB") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			gpus = append(gpus, GPUStatus{
				Utilization: strings.TrimSpace(parts[0]),
				MemoryUsed:  strings.TrimSpace(parts[1]),
				MemoryTotal: strings.TrimSpace(parts[2]),
			})
		}
	}

	return gpus
}
