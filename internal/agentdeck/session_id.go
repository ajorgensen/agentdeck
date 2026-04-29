package agentdeck

import (
	"os"
	"regexp"
	"time"
)

var (
	codexThreadIDPattern     = regexp.MustCompile(`"thread_id"\s*:\s*"([^"]+)"`)
	openCodeSessionIDPattern = regexp.MustCompile(`"sessionID"\s*:\s*"([^"]+)"`)
)

func waitForLogMatch(path string, pattern *regexp.Regexp, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for {
		if value := firstLogMatch(path, pattern); value != "" {
			return value
		}
		if time.Now().After(deadline) {
			return ""
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func firstLogMatch(path string, pattern *regexp.Regexp) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	matches := pattern.FindSubmatch(data)
	if len(matches) != 2 {
		return ""
	}
	return string(matches[1])
}
