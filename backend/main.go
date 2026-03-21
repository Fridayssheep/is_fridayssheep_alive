package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"frisheep-alive-backend/router"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
)

var sshClient *ssh.Client

func initSSHClient() error {
	host := os.Getenv("SSH_HOST")
	port := os.Getenv("SSH_PORT")
	user := os.Getenv("SSH_USER")
	password := os.Getenv("SSH_PASSWORD")
	keyPath := os.Getenv("SSH_KEY_PATH")

	if host == "" || user == "" {
		return fmt.Errorf("SSH_HOST and SSH_USER environment variables are required")
	}
	if port == "" {
		port = "22"
	}

	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("unable to read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("unable to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("either SSH_PASSWORD or SSH_KEY_PATH must be provided")
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		// INSECURE: accept any host key in a trusted local env.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}

	sshClient = client
	return nil
}

func main() {
	_ = godotenv.Load("../.env", ".env")

	err := initSSHClient()
	if err != nil {
		log.Fatalf("SSH Initialization failed: %v", err)
	}
	fmt.Println("SSH connected successfully to", os.Getenv("SSH_HOST"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	githubUsername := os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		log.Fatal("GITHUB_USERNAME environment variable is not configured. Exiting.")
	}

	ollamaURL := os.Getenv("OLLAMA_API_URL")
	if ollamaURL == "" {
		log.Fatal("OLLAMA_API_URL environment variable is not configured. Exiting.")
	}

	awURL := os.Getenv("ACTIVITYWATCH_URL")
	if awURL == "" {
		log.Fatal("ACTIVITYWATCH_URL environment variable is not configured. Exiting.")
	}

	// 从环境变量加载刷新时间，默认 5 秒
	intervalStr := os.Getenv("REFRESH_INTERVAL")
	refreshInterval := 5
	if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
		refreshInterval = i
	}

	// 启动后台轮询数据（不再在每次请求时实时 SSH 去读取）
	router.StartPolling(sshClient, githubUsername, ollamaURL, awURL, refreshInterval)

	// 挂载路由（处理 GET 请求，直接返回内存缓存的值）
	handler := router.SetupRouter(sshClient)
	fmt.Printf("Starting remote monitoring backend on :%s, auto-refresh every %ds\n", port, refreshInterval)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
