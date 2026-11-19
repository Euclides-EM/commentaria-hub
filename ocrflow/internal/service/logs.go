package service

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

type Logs struct {
	systemdUnit  string
	defaultLines int
	maxLines     int
	run          commandRunner
}

func NewLogsService(systemdUnit string, defaultLines, maxLines int) *Logs {
	if strings.TrimSpace(systemdUnit) == "" {
		systemdUnit = "commentaria-hub-api"
	}
	if defaultLines <= 0 {
		defaultLines = 200
	}
	if maxLines <= 0 {
		maxLines = 2000
	}
	if defaultLines > maxLines {
		defaultLines = maxLines
	}
	return &Logs{
		systemdUnit:  systemdUnit,
		defaultLines: defaultLines,
		maxLines:     maxLines,
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		},
	}
}

func (l *Logs) Tail(ctx context.Context, requestedLines int) (*common.LogTail, error) {
	linesToRead := l.normalizeLineCount(requestedLines)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	output, err := l.run(
		ctx,
		"journalctl",
		"-u", l.systemdUnit,
		"-n", strconv.Itoa(linesToRead),
		"--no-pager",
	)
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			return nil, fmt.Errorf("read journalctl logs for %s: %s: %w", l.systemdUnit, msg, err)
		}
		return nil, fmt.Errorf("read journalctl logs for %s: %w", l.systemdUnit, err)
	}

	trimmed := strings.TrimRight(string(output), "\n")
	var lines []string
	if trimmed != "" {
		lines = strings.Split(trimmed, "\n")
	}

	return &common.LogTail{
		Count:   len(lines),
		Lines:   lines,
		Service: l.systemdUnit,
	}, nil
}

func (l *Logs) normalizeLineCount(requestedLines int) int {
	if requestedLines <= 0 {
		return l.defaultLines
	}
	if requestedLines > l.maxLines {
		return l.maxLines
	}
	return requestedLines
}
