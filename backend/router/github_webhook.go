package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// 发送消息到 NapCat
func sendNapcatMessage(msg string) {
	apiUrl := os.Getenv("NAPCAT_API_URL")
	groupIDsStr := os.Getenv("NAPCAT_GROUP_ID") // 目标群号，支持逗号分隔多个
	userIDsStr := os.Getenv("NAPCAT_USER_ID")   // 目标私聊QQ号，支持逗号分隔多个
	token := os.Getenv("NAPCAT_TOKEN")          // 认证用的 token (如有)

	if apiUrl == "" {
		fmt.Println("Napcat API URL not configured, skipping push.")
		return
	}

	// 抽出通用的 HTTP POST 逻辑以带上 Token Header
	doPost := func(url string, payload []byte) (*http.Response, error) {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		client := &http.Client{}
		return client.Do(req)
	}

	// 推送到指定的群聊
	if groupIDsStr != "" {
		groups := strings.Split(groupIDsStr, ",")
		for _, gStr := range groups {
			gStr = strings.TrimSpace(gStr)
			if gStr == "" {
				continue
			}

			groupId, err := strconv.ParseInt(gStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid NAPCAT_GROUP_ID member '%s', must be integer: %v\n", gStr, err)
				continue
			}

			payload := map[string]interface{}{
				"group_id": groupId,
				"message":  msg,
			}

			jsonValue, _ := json.Marshal(payload)
			resp, err := doPost(fmt.Sprintf("%s/send_group_msg", apiUrl), jsonValue)
			if err != nil {
				fmt.Printf("Failed to send message to NapCat group %d: %v\n", groupId, err)
				continue
			}
			resp.Body.Close()
		}
	}

	// 推送到指定的个人QQ（私聊）
	if userIDsStr != "" {
		users := strings.Split(userIDsStr, ",")
		for _, uStr := range users {
			uStr = strings.TrimSpace(uStr)
			if uStr == "" {
				continue
			}

			userId, err := strconv.ParseInt(uStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid NAPCAT_USER_ID member '%s', must be integer: %v\n", uStr, err)
				continue
			}

			payload := map[string]interface{}{
				"user_id": userId,
				"message": msg,
			}

			jsonValue, _ := json.Marshal(payload)
			resp, err := doPost(fmt.Sprintf("%s/send_msg", apiUrl), jsonValue)
			if err != nil {
				fmt.Printf("Failed to send private message to NapCat user %d: %v\n", userId, err)
				continue
			}
			resp.Body.Close()
		}
	}
}

// 处理来自 GitHub 的事件结构体 (简略版)
type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Pusher struct {
		Name string `json:"name"`
	} `json:"pusher"`
	Commits []struct {
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
	} `json:"commits"`
}

type GitHubPullRequestEvent struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

// GithubWebhookHandler 处理 GitHub 主动推送的 Webhook
func GithubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Github 通过该 Header 标识事件类型
	eventType := r.Header.Get("X-GitHub-Event")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch eventType {
	case "ping":
		var repo struct {
			Repository struct {
				Name string `json:"name"`
			} `json:"repository"`
		}
		if err := json.Unmarshal(body, &repo); err == nil {
			msg := fmt.Sprintf("✅ 成功接入 GitHub Webhook 监听！\n📦 仓库：%s\n设置监听成功", repo.Repository.Name)
			go sendNapcatMessage(msg)
		}

	case "push":
		var event GitHubPushEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "Invalid push payload", http.StatusBadRequest)
			return
		}

		branch := strings.TrimPrefix(event.Ref, "refs/heads/")
		commitCount := len(event.Commits)

		if commitCount > 0 {
			msg := fmt.Sprintf("GitHub 仓库有新的代码推送！\n📦 仓库：%s\n🌿 分支：%s\n👤 提交者：%s\n\n📌 包含 %d 个 Commit：",
				event.Repository.Name, branch, event.Pusher.Name, commitCount)

			for i, commit := range event.Commits {
				if i >= 5 {
					msg += "\n...以及更多未展示"
					break
				}
				// 取第一行作为简述
				shortMsg := strings.Split(commit.Message, "\n")[0]
				msg += fmt.Sprintf("\n- %s (%s)", shortMsg, commit.Author.Name)
			}
			go sendNapcatMessage(msg)
		}

	case "pull_request":
		var event GitHubPullRequestEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "Invalid PR payload", http.StatusBadRequest)
			return
		}

		// 只对常用的动作做推送
		if event.Action == "opened" || event.Action == "closed" || event.Action == "reopened" {
			actionMsg := "发起了"
			switch event.Action {
			case "closed":
				actionMsg = "关闭或合并了"
			case "reopened":
				actionMsg = "重新打开了"
			}

			msg := fmt.Sprintf("🔀 GitHub PR 更新！\n👤 %s %s Pull Request\n📦 仓库：%s\n📝 标题：%s\n🔗 链接：%s",
				event.PullRequest.User.Login, actionMsg, event.Repository.Name, event.PullRequest.Title, event.PullRequest.HTMLURL)
			go sendNapcatMessage(msg)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
