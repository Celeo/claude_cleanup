package main

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"empty", "", 5, ""},
		{"shorter than n", "hi", 5, "hi"},
		{"equal to n", "hello", 5, "hello"},
		{"longer than n gets ellipsis", "hello world", 8, "hello w…"},
		{"newlines collapse to spaces", "line one\nline two", 80, "line one line two"},
		{"leading and trailing whitespace trimmed", "  padded  ", 80, "padded"},
		{"trim happens before length check", "  hi  ", 4, "hi"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncate(tc.in, tc.n); got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
		})
	}
}

func TestDecodeProjectDir(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no leading dash returned as-is", "weird-name", "weird-name"},
		{"leading dash becomes slash, others become slashes too (lossy)", "-home-matt-code", "/home/matt/code"},
		// Documents the lossiness: a real underscore-vs-dash distinction is preserved,
		// but a real dash in the original cwd cannot be recovered.
		{"underscores in original path survive", "-home-matt-my_project", "/home/matt/my_project"},
		{"real dashes in original path are indistinguishable from separators", "-home-matt-my-project", "/home/matt/my/project"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := decodeProjectDir(tc.in); got != tc.want {
				t.Errorf("decodeProjectDir(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPickTitleFromReader(t *testing.T) {
	const (
		customLine     = `{"type":"custom-title","customTitle":"My Custom"}`
		aiLine         = `{"type":"ai-title","aiTitle":"AI-generated thing"}`
		aiLineSecond   = `{"type":"ai-title","aiTitle":"Should be ignored"}`
		lastPromptLine = `{"type":"last-prompt","lastPrompt":"the most recent thing the user asked"}`
		userLine       = `{"type":"user","message":{"role":"user","content":"first user prompt"}}`
		userMetaLine   = `{"type":"user","isMeta":true,"message":{"role":"user","content":"meta caveat"}}`
		userBlocks     = `{"type":"user","message":{"role":"user","content":[{"type":"text","text":"prompt from blocks"}]}}`
	)

	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"empty stream returns fallback", nil, "SESSION-ID"},
		{"custom beats everything", []string{customLine, aiLine, lastPromptLine, userLine}, "My Custom"},
		{"ai beats last-prompt and first user", []string{aiLine, lastPromptLine, userLine}, "AI-generated thing"},
		{"first ai-title wins (latest is ignored)", []string{aiLine, aiLineSecond}, "AI-generated thing"},
		{"last-prompt beats first user", []string{lastPromptLine, userLine}, "the most recent thing the user asked"},
		{"first user prompt is the fallback", []string{userLine}, "first user prompt"},
		{"meta user turns are skipped", []string{userMetaLine, userLine}, "first user prompt"},
		{"block-shaped user content is read", []string{userBlocks}, "prompt from blocks"},
		{"truly empty falls back to id", []string{`{"type":"file-history-snapshot"}`}, "SESSION-ID"},
		{"malformed lines are ignored", []string{"not json at all", userLine}, "first user prompt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(strings.Join(tc.lines, "\n"))
			if got := pickTitleFromReader(r, "SESSION-ID"); got != tc.want {
				t.Errorf("pickTitleFromReader = %q, want %q", got, tc.want)
			}
		})
	}
}
