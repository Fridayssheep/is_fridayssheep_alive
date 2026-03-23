package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GitHubEvent struct {
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"` // ISO8601
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
}

type GitHubStatus struct {
	HasRecentActivity bool         `json:"has_recent_activity"`
	LastActivityType  string       `json:"last_activity_type"`
	LastActivityRepo  string       `json:"last_activity_repo"`
	LastActivityTime  string       `json:"last_activity_time"`
	RecentCommits     []CommitInfo `json:"recent_commits"`
	Error             string       `json:"error,omitempty"`
}

type CommitInfo struct {
	Message  string `json:"message"`
	Author   string `json:"author"`
	URL      string `json:"url"`
	Date     string `json:"date"`
	SHA      string `json:"sha"`
	ShortSHA string `json:"short_sha"`
}

var lastFetch time.Time
var cachedStatus GitHubStatus

func GetGitHubStatus(username string) GitHubStatus {
	if username == "" {
		return GitHubStatus{Error: "Username not configured"}
	}

	// Simple caching to avoid rate limits (e.g., fetch at most once per minute)
	if time.Since(lastFetch) < 1*time.Minute && cachedStatus.Error == "" {
		return cachedStatus
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events/public", username)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GitHubStatus{Error: err.Error()}
	}

	// GitHub API requires a user-agent
	req.Header.Set("User-Agent", "is-frisheep-alive-monitor")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return GitHubStatus{Error: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && strings.Contains(resp.Header.Get("X-RateLimit-Remaining"), "0") {
		return GitHubStatus{Error: "Rate limit exceeded"}
	}

	if resp.StatusCode != http.StatusOK {
		return GitHubStatus{Error: fmt.Sprintf("Received status: %d", resp.StatusCode)}
	}

	var events []GitHubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return GitHubStatus{Error: err.Error()}
	}

	lastFetch = time.Now()

	status := GitHubStatus{HasRecentActivity: false}
	if len(events) > 0 {
		// Find the latest push event, or just any event
		for _, ev := range events {
			if ev.Type == "PushEvent" {
				status.HasRecentActivity = true
				status.LastActivityType = ev.Type
				status.LastActivityRepo = ev.Repo.Name
				status.LastActivityTime = ev.CreatedAt
				break
			}
		}

		// If no push event, fallback to the latest any event
		if !status.HasRecentActivity {
			status.HasRecentActivity = true
			status.LastActivityType = events[0].Type
			status.LastActivityRepo = events[0].Repo.Name
			status.LastActivityTime = events[0].CreatedAt
		}

		// Fetch recent commits for the active repo
		if status.LastActivityRepo != "" {
			commitsURL := fmt.Sprintf("https://api.github.com/repos/%s/commits?per_page=5", status.LastActivityRepo)
			commitsReq, cerr := http.NewRequest("GET", commitsURL, nil)
			if cerr == nil {
				commitsReq.Header.Set("User-Agent", "is-frisheep-alive-monitor")
				commitsReq.Header.Set("Accept", "application/vnd.github.v3+json")

				if commitsResp, cerr := client.Do(commitsReq); cerr == nil {
					defer commitsResp.Body.Close()
					if commitsResp.StatusCode == http.StatusOK {
						var rc []struct {
							SHA     string `json:"sha"`
							HtmlUrl string `json:"html_url"`
							Commit  struct {
								Message string `json:"message"`
								Author  struct {
									Name string `json:"name"`
									Date string `json:"date"`
								} `json:"author"`
							} `json:"commit"`
						}
						if json.NewDecoder(commitsResp.Body).Decode(&rc) == nil {
							for _, c := range rc {
								shortSHA := c.SHA
								if len(shortSHA) > 7 {
									shortSHA = shortSHA[:7]
								}
								status.RecentCommits = append(status.RecentCommits, CommitInfo{
									Message:  strings.Split(c.Commit.Message, "\n")[0], // Get only the first line of the commit message
									Author:   c.Commit.Author.Name,
									URL:      c.HtmlUrl,
									Date:     c.Commit.Author.Date,
									SHA:      c.SHA,
									ShortSHA: shortSHA,
								})
							}
						}
					}
				}
			}
		}
	}

	cachedStatus = status
	return cachedStatus
}
