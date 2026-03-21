package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ActivityWatchStatus struct {
	App      string `json:"app"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
}

type AWEvent struct {
	Data struct {
		App   string `json:"app"`
		Title string `json:"title"`
	} `json:"data"`
	Timestamp string  `json:"timestamp"`
	Duration  float64 `json:"duration"`
}

// GetActivityWatchStatus calls the local ActivityWatch API to get the current window
func GetActivityWatchStatus(baseURL string) ActivityWatchStatus {
	client := http.Client{Timeout: 2 * time.Second}
	var fallback ActivityWatchStatus

	// 1. Get buckets
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/0/buckets", baseURL), nil)
	if err != nil {
		return fallback
	}

	resp, err := client.Do(req)
	if err != nil {
		return fallback // ActivityWatch not running or unreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fallback
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallback
	}

	var buckets map[string]interface{}
	if err := json.Unmarshal(body, &buckets); err != nil {
		return fallback
	}

	var windowBucket string
	for k := range buckets {
		if strings.HasPrefix(k, "aw-watcher-window_") {
			windowBucket = k
			break
		}
	}

	if windowBucket == "" {
		return fallback
	}

	// Get latest event in the window bucket
	eventURL := fmt.Sprintf("%s/api/0/buckets/%s/events?limit=1", baseURL, windowBucket)
	reqQueue, err := http.NewRequest("GET", eventURL, nil)
	if err != nil {
		return fallback
	}

	respQueue, err := client.Do(reqQueue)
	if err != nil {
		return fallback
	}
	defer respQueue.Body.Close()

	if respQueue.StatusCode != 200 {
		return fallback
	}

	bodyQueue, err := io.ReadAll(respQueue.Body)
	if err != nil {
		return fallback
	}

	var events []AWEvent
	if err := json.Unmarshal(bodyQueue, &events); err != nil {
		return fallback
	}

	if len(events) > 0 {
		ev := events[0]

		// Check if it's recent enough to be considered "active"
		// ActivityWatch timestamps are UTC like "2023-10-10T12:00:00Z"
		evTime, err := time.Parse(time.RFC3339Nano, ev.Timestamp)
		isActive := true
		if err == nil {
			// Compute the end time of the event (start time + duration)
			endTime := evTime.Add(time.Duration(ev.Duration) * time.Second)

			// If the event ended more than 10 minutes ago, consider the user AFK/offline
			if time.Since(endTime) > 10*time.Minute {
				isActive = false
			}
		}

		return ActivityWatchStatus{
			App:      ev.Data.App,
			Title:    ev.Data.Title,
			IsActive: isActive,
		}
	}

	return fallback
}
