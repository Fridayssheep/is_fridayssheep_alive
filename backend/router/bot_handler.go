package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type NapcatWebhookReq struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	RawMessage  string `json:"raw_message"`
	UserId      int64  `json:"user_id"`
}

type NapcatWebhookResp struct {
	Reply string `json:"reply"`
}

// BotWebhookHandler 处理来自 NapCat / OneBot v11 的 Webhook 请求
func BotWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NapcatWebhookReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 只处理普通消息事件
	if req.PostType != "message" {
		w.WriteHeader(http.StatusOK)
		return
	}

	replyMsg := ""
	msgText := strings.TrimSpace(req.RawMessage)

	if strings.HasPrefix(msgText, "/在干啥") || strings.HasPrefix(msgText, "/看看你的") {
		globalActivityCache.mu.RLock()
		act := globalActivityCache.state
		globalActivityCache.mu.RUnlock()

		if act.IsActive {
			replyMsg = fmt.Sprintf("Fridayssheep 的HOMOPC在线！\n💻 正在使用：%s\n📝 详细：%s", act.App, act.Title)
		} else {
			replyMsg = "Fridayssheep 似乎不在电脑前，可能在摸鱼或者睡觉..."
		}
	} else if strings.HasPrefix(msgText, "/状态") || strings.HasPrefix(msgText, "/status") {
		globalCache.mu.RLock()
		data := globalCache.state
		globalCache.mu.RUnlock()

		if data.Status == "ok" {
			replyMsg = fmt.Sprintf("💻 工作站状态 (更新时间: %s)：\nCPU: %.1f%%\n内存: %dMB / %dMB",
				time.Now().Format("15:04:05"), data.System.CPUPercent, data.System.MemUsed/(1024*1024), data.System.MemTotal/(1024*1024))

			if len(data.GPUs) > 0 {
				for i, gpu := range data.GPUs {
					replyMsg += fmt.Sprintf("\n🎮 GPU%d: %s, %s / %s", i, gpu.Utilization, gpu.MemoryUsed, gpu.MemoryTotal)
				}
			}

			if len(data.Ollama) > 0 {
				replyMsg += "\n\n正在运行的大模型(Ollama):"
				for _, model := range data.Ollama {
					replyMsg += fmt.Sprintf("\n- %s (占用 VRAM: %d MB)", model.Name, model.SizeVRAM/(1024*1024))
				}
			} else {
				replyMsg += "\n\nOllama 状态: 无运行中的模型，或未开启服务"
			}
		} else {
			replyMsg = "⚠️ 目标工作站当前离线或无法连接！"
		}
	} else if strings.HasPrefix(msgText, "/github") || strings.HasPrefix(msgText, "/代码") {
		globalCache.mu.RLock()
		data := globalCache.state
		globalCache.mu.RUnlock()

		gh := data.GitHub
		if gh.Error != "" {
			replyMsg = "⚠️ 获取 GitHub 状态失败：" + gh.Error
		} else if !gh.HasRecentActivity {
			replyMsg = "Fridayssheep 最近好像在 GitHub 上摸鱼，没有什么新动态..."
		} else {
			replyMsg = fmt.Sprintf("Fridayssheep 最近的 GitHub 动态：\n📝 操作：%s\n📦 仓库：%s\n⏱️ 时间：%s",
				gh.LastActivityType, gh.LastActivityRepo, gh.LastActivityTime)

			if len(gh.RecentCommits) > 0 {
				replyMsg += "\n💡 最新 Commit："
				for i, commit := range gh.RecentCommits {
					if i >= 3 {
						break
					}
					// 简单截断过长的 commit message
					msg := commit.Message
					if len(msg) > 100 {
						msg = msg[:97] + "..."
					}
					replyMsg += fmt.Sprintf("\n- %s", msg)
				}
			}
		}
	} else if strings.HasPrefix(msgText, "/help") || strings.HasPrefix(msgText, "/菜单") || strings.HasPrefix(msgText, "/帮助") {
		replyMsg = "可用指令：\n" +
			"1. /在干啥 或 /看看你的 - 查看这BYD当前是否还活着及正在使用的应用\n" +
			"2. /状态 或 /status - 查看工作站当前运行状态(CPU、内存、GPU)\n" +
			"3. /github 或 /代码 - 查看这BYD最近在 GitHub 的活动\n" +
			"详情可以访问 example.com 查看监控面板！"
	}

	// 如果没有匹配到指令，给 Napcat 返回 200 并退出，避免不必要的回复打扰群聊
	if replyMsg == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 返回符合 OneBot v11 快捷回复格式的 JSON
	resp := NapcatWebhookResp{Reply: replyMsg}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
