package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Session struct {
	ID         string
	Path       string
	ProjectDir string
	Title      string
	ModifiedAt time.Time
}

// sessionItem wraps Session for the bubbles list (which expects Title() — the
// field name on Session would collide).
type sessionItem struct{ Session }

func (i sessionItem) Title() string { return i.Session.Title }
func (i sessionItem) Description() string {
	return i.Session.ProjectDir + " · " + i.Session.ModifiedAt.Format("2006-01-02 15:04")
}
func (i sessionItem) FilterValue() string { return i.Session.Title }

// LoadSessions walks ~/.claude/projects/*/*.jsonl and returns sessions sorted
// by modification time, newest first.
func LoadSessions() ([]Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	root := filepath.Join(home, ".claude", "projects")
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return loadSessionsFromRoot(root)
}

// loadSessionsFromRoot walks any root that uses the projects/<encoded>/<id>.jsonl
// layout (the real Claude Code projects dir, or the trash dir).
func loadSessionsFromRoot(root string) ([]Session, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, projectEntry := range entries {
		if !projectEntry.IsDir() {
			continue
		}
		projectDir := filepath.Join(root, projectEntry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			path := filepath.Join(projectDir, f.Name())
			info, err := f.Info()
			if err != nil {
				continue
			}
			s := Session{
				ID:         strings.TrimSuffix(f.Name(), ".jsonl"),
				Path:       path,
				ProjectDir: decodeProjectDir(projectEntry.Name()),
				ModifiedAt: info.ModTime(),
			}
			s.Title = pickTitle(path, s.ID)
			sessions = append(sessions, s)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModifiedAt.After(sessions[j].ModifiedAt)
	})
	return sessions, nil
}

// decodeProjectDir converts "-home-matt-code-claude-cleanup" back to
// "/home/matt/code/claude-cleanup" (best-effort; Claude Code replaces / with -
// but original dashes vs separators are ambiguous, so just swap leading - for /).
func decodeProjectDir(encoded string) string {
	if strings.HasPrefix(encoded, "-") {
		return "/" + strings.ReplaceAll(encoded[1:], "-", "/")
	}
	return encoded
}

// pickTitle scans the jsonl once and returns the best available title.
// Precedence: latest custom-title > ai-title > last-prompt > first user prompt.
func pickTitle(path, id string) string {
	f, err := os.Open(path)
	if err != nil {
		return id
	}
	defer f.Close()
	return pickTitleFromReader(f, id)
}

// pickTitleFromReader is the reader-shaped core of pickTitle. Split out so the
// precedence logic can be tested without touching the filesystem.
func pickTitleFromReader(r io.Reader, fallback string) string {
	var customTitle, aiTitle, lastPrompt, firstPrompt string
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		var head struct {
			Type        string `json:"type"`
			Subtype     string `json:"subtype"`
			CustomTitle string `json:"customTitle"`
			AITitle     string `json:"aiTitle"`
			LastPrompt  string `json:"lastPrompt"`
			IsMeta      bool   `json:"isMeta"`
			Message     struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &head); err != nil {
			continue
		}
		switch head.Type {
		case "custom-title":
			customTitle = head.CustomTitle
		case "ai-title":
			if aiTitle == "" {
				aiTitle = head.AITitle
			}
		case "last-prompt":
			lastPrompt = head.LastPrompt
		case "user":
			if firstPrompt == "" && !head.IsMeta && head.Message.Role == "user" {
				text := extractText(head.Message.Content)
				text = stripCommandTags(text)
				if text != "" {
					firstPrompt = text
				}
			}
		}
	}

	switch {
	case customTitle != "":
		return customTitle
	case aiTitle != "":
		return aiTitle
	case lastPrompt != "":
		return truncate(lastPrompt, 80)
	case firstPrompt != "":
		return truncate(firstPrompt, 80)
	}
	return fallback
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
