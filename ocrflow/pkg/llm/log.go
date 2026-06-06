package llm

import (
	"strings"
	"time"
)

func logPrefix(logLabel string) string {
	ts := time.Now().Format("15:04:05")
	logLabel = strings.TrimSpace(logLabel)
	if logLabel == "" {
		return " " + ts
	}
	return " " + ts + " " + logLabel
}
