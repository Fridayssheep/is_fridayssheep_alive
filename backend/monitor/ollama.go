package monitor

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type OllamaStatus struct {
	Models []struct {
		Name    string `json:"name"`
		Model   string `json:"model"`
		Size    uint64 `json:"size"`
		Digest  string `json:"digest"`
		Details struct {
			ParentModel       string   `json:"parent_model"`
			Format            string   `json:"format"`
			Family            string   `json:"family"`
			Families          []string `json:"families"`
			ParameterSize     string   `json:"parameter_size"`
			QuantizationLevel string   `json:"quantization_level"`
		} `json:"details"`
		ExpiresAt time.Time `json:"expires_at"`
		SizeVRAM  uint64    `json:"size_vram"`
	} `json:"models"`
}

type OllamaRunningModel struct {
	Name     string `json:"name"`
	SizeVRAM uint64 `json:"size_vram"`
}

func GetOllamaStatus(client *ssh.Client, ollamaURL string) []OllamaRunningModel {
	var running []OllamaRunningModel
	if client == nil {
		return running
	}

	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	cmd := fmt.Sprintf("curl -s %s/api/ps", ollamaURL)
	out, err := RunCommand(client, cmd)
	if err != nil || out == "" {
		return running
	}

	var status OllamaStatus
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return running
	}

	for _, m := range status.Models {
		running = append(running, OllamaRunningModel{
			Name:     m.Name,
			SizeVRAM: m.SizeVRAM,
		})
	}

	return running
}
