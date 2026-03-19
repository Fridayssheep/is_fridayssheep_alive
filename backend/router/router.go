package router

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"frisheep-alive-backend/monitor"

	"github.com/rs/cors"
	"golang.org/x/crypto/ssh"
)

type DashboardStatus struct {
	System     monitor.SysStatus            `json:"system"`
	GPUs       []monitor.GPUStatus          `json:"gpus"`
	Ollama     []monitor.OllamaRunningModel `json:"ollama"`
	GitHub     monitor.GitHubStatus         `json:"github"`
	UpdateTime string                       `json:"update_time"`
	Status     string                       `json:"status"` // 用于标识远端拉取状态，比如 "ok" 或是 "error"
}

type StatusCache struct {
	mu    sync.RWMutex
	state DashboardStatus
}

var globalCache *StatusCache

// StartPolling 启动后台定时轮询任务
func StartPolling(sshClient *ssh.Client, githubUsername string, ollamaURL string, intervalSeconds int) {
	globalCache = &StatusCache{
		state: DashboardStatus{Status: "initializing"},
	}

	updateTask := func() {
		// 检查 SSH 连接是否存活。通过执行一个简单的 echo 命令测试
		_, err := monitor.RunCommand(sshClient, "echo 1")

		var sys monitor.SysStatus
		var gpus []monitor.GPUStatus
		var ollamaStat []monitor.OllamaRunningModel

		statusMsg := "ok"

		if err != nil {
			// 连接断开、目标机器离线或关机
			statusMsg = "offline"
		} else {
			sys = monitor.GetSysStatus(sshClient)
			gpus = monitor.GetGPUStatus(sshClient)
			ollamaStat = monitor.GetOllamaStatus(sshClient, ollamaURL)

			if sys.MemTotal == 0 {
				statusMsg = "error_fetching_data"
			}
		}

		// Github 状态不依赖工作站，应该始终拉取
		gh := monitor.GetGitHubStatus(githubUsername)

		globalCache.mu.Lock()
		globalCache.state = DashboardStatus{
			System:     sys,
			GPUs:       gpus,
			Ollama:     ollamaStat,
			GitHub:     gh,
			UpdateTime: time.Now().Format(time.RFC3339),
			Status:     statusMsg,
		}
		globalCache.mu.Unlock()
	}

	// 启动一个 goroutine，先执行一次，然后按设定的间隔不断轮询
	go func() {
		updateTask()
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		for range ticker.C {
			updateTask()
		}
	}()
}

// SetupRouter 初始化挂载所有的路由接口并包装 CORS
func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if globalCache == nil {
			http.Error(w, `{"error": "Backend is not fully initialized"}`, http.StatusServiceUnavailable)
			return
		}

		// 使用读锁安全获取当前缓存状态
		globalCache.mu.RLock()
		data := globalCache.state
		globalCache.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// 配置 CORS 处理跨域
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	})

	return c.Handler(mux)
}
