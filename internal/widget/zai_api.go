package widget

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// ZaiLimit represents a single usage limit from Z.ai API
type ZaiLimit struct {
	Type          string  `json:"type"`
	Usage         int     `json:"usage"`
	Remaining     int     `json:"remaining"`
	Percentage    float64 `json:"percentage"`
	NextResetTime int64   `json:"nextResetTime"`
}

// ZaiData contains the limits data from Z.ai API
type ZaiData struct {
	Limits []ZaiLimit `json:"limits"`
}

// ZaiResponse represents the full response from Z.ai API
type ZaiResponse struct {
	Code    int     `json:"code"`
	Msg     string  `json:"msg"`
	Success bool    `json:"success"`
	Data    ZaiData `json:"data"`
}

// getApiKey reads the Z.ai API key from ~/.keys file
func getApiKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	file, err := os.Open(home + "/.keys")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := scanner.Text()
		key = strings.TrimSpace(key)
		if strings.HasPrefix(key, "ZAI=") {
			key = strings.TrimPrefix(key, "ZAI=")
			key, _ = strings.CutPrefix(key, "\" ")
			key, _ = strings.CutSuffix(key, "\"")
			return key, nil
		}
	}

	return "", fmt.Errorf("ZAI= key not found")
}

// updateZaiAPI fetches and updates Z.ai API quota status
func (w *Widget) updateZaiAPI() {
	if w.apiKey == "" {
		w.messages = []string{"> ERROR", "> No API key found", "> Check ~/.keys file"}
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.z.ai/api/monitor/usage/quota/limit", nil)
	if err != nil {
		w.messages = []string{"> ERROR", "> Failed to create request"}
		return
	}

	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	resp, err := client.Do(req)
	if err != nil {
		w.messages = []string{"> ERROR", "> Network error", "> Check connection"}
		return
	}
	defer resp.Body.Close()

	var zai ZaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&zai); err != nil {
		w.messages = []string{"> ERROR", "> JSON parse error"}
		return
	}

	// Handle authentication failure
	if !zai.Success || zai.Code == 1000 {
		w.messages = []string{"> ERROR", "> Auth failed", zai.Msg}
		return
	}

	// Find TOKENS_LIMIT and update messages
	for _, limit := range zai.Data.Limits {
		if limit.Type == "TOKENS_LIMIT" {
			w.zaiData = &zai
			w.updateMessages(&limit)
			return
		}
	}

	w.messages = []string{"> ERROR", "> No token limit found"}
}

// updateMessages formats and updates messages based on API data
func (w *Widget) updateMessages(limit *ZaiLimit) {
	resetTime := time.UnixMilli(limit.NextResetTime)
	availPct := 100 - int(limit.Percentage)

	// Choose emoji icon based on availability
	icon := "🟢"
	switch {
	case availPct < 60:
		icon = "🟡"
	case availPct < 40:
		icon = "🟠"
	case availPct < 20:
		icon = "🔴"
	case availPct == 0:
		icon = "⬛"
	}

	// Format messages with emoji icons
	timeStr := resetTime.Format("03:04PM")
	usageStr := fmt.Sprintf("%.2f", limit.Percentage)

	w.messages = []string{
		"> Z.AI API STATUS",
		"> Available: " + fmt.Sprintf("%d%% %s", availPct, icon),
		"> Usage: " + usageStr + "%",
		"> Reset at: " + timeStr + " ⏳",
		">",
		"> Tokens remaining:",
		fmt.Sprintf("> %d tokens", limit.Remaining),
	}
}
