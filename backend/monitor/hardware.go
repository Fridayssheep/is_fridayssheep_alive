package monitor

import (
	"strings"

	"golang.org/x/crypto/ssh"
)

type HardwareInfo struct {
	OS       string   `json:"os"`
	CPU      string   `json:"cpu"`
	MemTotal string   `json:"mem_total"`
	GPUs     []string `json:"gpus"`
}

func getCleanOutput(client *ssh.Client, cmd string) (string, error) {
	// 使用 ====START==== 和 ====END==== 包围实际输出，来完全无视.bashrc或.bash_logout的噪音
	wrappedCmd := `echo "====START===="; ` + cmd + `; echo "====END===="`
	out, err := RunCommand(client, wrappedCmd)
	if err != nil {
		return "", err
	}

	startIdx := strings.LastIndex(out, "====START====")
	endIdx := strings.LastIndex(out, "====END====")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		cleanPart := out[startIdx+13 : endIdx]
		return strings.TrimSpace(cleanPart), nil
	}
	return strings.TrimSpace(out), nil
}

func GetHardwareInfo(client *ssh.Client) HardwareInfo {
	var info HardwareInfo

	// Get OS
	if out, err := getCleanOutput(client, "cat /etc/os-release | grep PRETTY_NAME | cut -d '=' -f 2 | tr -d '\"'"); err == nil {
		lines := strings.Split(out, "\n")
		info.OS = strings.TrimSpace(lines[len(lines)-1])
	}

	// Get CPU
	if out, err := getCleanOutput(client, "lscpu | grep 'Model name' | awk -F ':' '{print $2}' | xargs"); err == nil {
		lines := strings.Split(out, "\n")
		info.CPU = strings.TrimSpace(lines[len(lines)-1])
	}

	// Get Mem
	if out, err := getCleanOutput(client, "awk '/MemTotal/ {printf \"%.1f GB\", $2/1024/1024}' /proc/meminfo"); err == nil {
		lines := strings.Split(out, "\n")
		info.MemTotal = strings.TrimSpace(lines[len(lines)-1])
	}

	// Get GPUs (Support Nvidia, Intel, AMD by using lspci)
	if out, err := getCleanOutput(client, "lspci | grep -iE 'vga|3d|display' | awk -F ': ' '{print $2}'"); err == nil && out != "" {
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				info.GPUs = append(info.GPUs, line)
			}
		}
	}

	return info
}
