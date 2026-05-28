package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
)

type convoMessage struct {
	Role string // "user" or "assistant"
	Text string
}

// LoadConversation parses a session jsonl and returns the human-readable
// message stream. Tool calls, tool results, thinking blocks, sidechains, and
// meta entries are skipped.
func LoadConversation(path string) ([]convoMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseConversation(f)
}

// parseConversation is the reader-shaped core of LoadConversation. Split out so
// the record-filtering logic can be tested without touching the filesystem.
func parseConversation(r io.Reader) ([]convoMessage, error) {
	var msgs []convoMessage
	scanner := bufio.NewScanner(r)
	// Session records can carry large inline tool outputs; bufio's default 64K
	// line limit trips on real files. 16MiB is comfortable headroom — anything
	// beyond that spills into a sibling overflow directory we don't read here.
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		var rec struct {
			Type        string `json:"type"`
			IsMeta      bool   `json:"isMeta"`
			IsSidechain bool   `json:"isSidechain"`
			Message     struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		if rec.IsMeta || rec.IsSidechain {
			continue
		}
		if rec.Type != "user" && rec.Type != "assistant" {
			continue
		}
		text := extractText(rec.Message.Content)
		text = stripCommandTags(text)
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		msgs = append(msgs, convoMessage{Role: rec.Message.Role, Text: text})
	}
	return msgs, scanner.Err()
}

// extractText pulls human text out of a message.content field that may be a
// plain string or an array of content blocks. Skips tool_use, tool_result,
// thinking, and image blocks.
func extractText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Sniff the first byte to dispatch between the two content shapes Claude Code
	// emits — a JSON string (typical for user turns) vs. an array of typed blocks
	// (typical for assistant turns). Cheaper and clearer than decoding into
	// interface{} and type-switching.
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
		return ""
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n\n")
}

var (
	systemReminderRe = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)
	caveatRe         = regexp.MustCompile(`(?s)<local-command-caveat>.*?</local-command-caveat>`)
	cmdStdoutRe      = regexp.MustCompile(`(?s)<local-command-stdout>.*?</local-command-stdout>`)
	cmdNameRe        = regexp.MustCompile(`(?s)<command-name>.*?</command-name>`)
	cmdMessageRe     = regexp.MustCompile(`(?s)<command-message>.*?</command-message>`)
	cmdArgsRe        = regexp.MustCompile(`(?s)<command-args>.*?</command-args>`)
)

// stripCommandTags removes the noisy local-command/system-reminder wrappers
// from message content so the rendered conversation reads like the chat that
// the user actually saw.
func stripCommandTags(s string) string {
	s = systemReminderRe.ReplaceAllString(s, "")
	s = caveatRe.ReplaceAllString(s, "")
	s = cmdStdoutRe.ReplaceAllString(s, "")
	s = cmdNameRe.ReplaceAllString(s, "")
	s = cmdMessageRe.ReplaceAllString(s, "")
	s = cmdArgsRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
