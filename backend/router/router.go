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

type ActivityCache struct {
	mu       sync.RWMutex
	state    monitor.ActivityWatchStatus
	cachedAt time.Time
}

type StatusCache struct {
	mu    sync.RWMutex
	state DashboardStatus
}

var globalCache *StatusCache
var globalActivityCache *ActivityCache

// StartPolling 启动后台定时轮询任务
func StartPolling(sshClient *ssh.Client, githubUsername string, ollamaURL string, awURL string, intervalSeconds int) {
	globalCache = &StatusCache{
		state: DashboardStatus{Status: "initializing"},
	}
	globalActivityCache = &ActivityCache{}

	var ghStatus monitor.GitHubStatus
	var ghMutex sync.RWMutex

	// 启动一个专门获取 Github 状态的 goroutine，限制 10 分钟拉取一次以避免 Rate limit exceeded
	go func() {
		for {
			gh := monitor.GetGitHubStatus(githubUsername)
			ghMutex.Lock()
			ghStatus = gh
			ghMutex.Unlock()
			time.Sleep(10 * time.Minute)
		}
	}()

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

		// Github 状态不依赖工作站，且受限频影响，从每10分钟更新一次的独立缓存拉取
		ghMutex.RLock()
		gh := ghStatus
		ghMutex.RUnlock()

		// ActivityWatch 状态从本地 Windows 主机拉取
		activity := monitor.GetActivityWatchStatus(awURL)

		globalActivityCache.mu.Lock()
		globalActivityCache.state = activity
		globalActivityCache.cachedAt = time.Now()
		globalActivityCache.mu.Unlock()

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
func SetupRouter(sshClient *ssh.Client) http.Handler {
	mux := http.NewServeMux()

	// 在启动时异步获取并缓存硬件信息，避免阻塞主线程，且只需获取一次
	var hwCache *monitor.HardwareInfo
	var hwMu sync.RWMutex
	go func() {
		info := monitor.GetHardwareInfo(sshClient)
		hwMu.Lock()
		hwCache = &info
		hwMu.Unlock()
	}()

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

	mux.HandleFunc("/api/activity", func(w http.ResponseWriter, r *http.Request) {
		if globalActivityCache == nil {
			http.Error(w, `{"error": "Activity backend is not fully initialized"}`, http.StatusServiceUnavailable)
			return
		}

		globalActivityCache.mu.RLock()
		data := globalActivityCache.state
		globalActivityCache.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// 配置 CORS 处理跨域
	mux.HandleFunc("/api/hardware", func(w http.ResponseWriter, r *http.Request) {
		hwMu.RLock()
		data := hwCache
		hwMu.RUnlock()

		if data == nil {
			http.Error(w, `{"error": "Hardware info is still loading"}`, http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// 增加针对 NapCat / OneBot v11 的 Webhook 机器人解析接口
	mux.HandleFunc("/api/bot", BotWebhookHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"}, // 允许 POST 处理机器人请求
	})

	return c.Handler(mux)
}
