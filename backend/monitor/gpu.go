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

	cmd := "nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total --format=csv,noheader"
	out, err := RunCommand(client, cmd)
	if err != nil {
		return gpus
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
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
