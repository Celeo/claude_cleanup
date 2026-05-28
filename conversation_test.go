package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStripCommandTags(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"no tags returns trimmed input", "  plain text  ", "plain text"},
		{"strips system-reminder", "before <system-reminder>noise</system-reminder> after", "before  after"},
		{"strips multiline system-reminder", "hi\n<system-reminder>line one\nline two</system-reminder>\nbye", "hi\n\nbye"},
		{"strips local-command-caveat", "x<local-command-caveat>caveat</local-command-caveat>y", "xy"},
		{"strips command stdout wrapper", "<local-command-stdout>output</local-command-stdout>real content", "real content"},
		{"strips command-name/message/args", `<command-name>/compact</command-name><command-message>msg</command-message><command-args>a b</command-args>after`, "after"},
		{"strips multiple kinds at once", "hi <system-reminder>x</system-reminder> <command-name>/y</command-name> bye", "hi   bye"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripCommandTags(tc.in); got != tc.want {
				t.Errorf("stripCommandTags(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractText(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"empty raw", "", ""},
		{"plain string content", `"hello world"`, "hello world"},
		{"empty string content", `""`, ""},
		{"single text block", `[{"type":"text","text":"hi"}]`, "hi"},
		{"multiple text blocks joined", `[{"type":"text","text":"one"},{"type":"text","text":"two"}]`, "one\n\ntwo"},
		{"tool_use blocks skipped", `[{"type":"text","text":"keep"},{"type":"tool_use","name":"X"}]`, "keep"},
		{"tool_result blocks skipped", `[{"type":"tool_result","content":"junk"},{"type":"text","text":"keep"}]`, "keep"},
		{"thinking blocks skipped", `[{"type":"thinking","thinking":"hidden"},{"type":"text","text":"keep"}]`, "keep"},
		{"empty text fields skipped", `[{"type":"text","text":""},{"type":"text","text":"keep"}]`, "keep"},
		{"malformed array returns empty", `[not json`, ""},
		{"malformed string returns empty", `"unterminated`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractText(json.RawMessage(tc.raw))
			if got != tc.want {
				t.Errorf("extractText(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseConversation(t *testing.T) {
	lines := []string{
		`{"type":"file-history-snapshot"}`,
		`{"type":"user","message":{"role":"user","content":"hello"}}`,
		`{"type":"user","isMeta":true,"message":{"role":"user","content":"<caveat>"}}`,
		`{"type":"user","isSidechain":true,"message":{"role":"user","content":"subagent"}}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi back"},{"type":"tool_use","name":"X"}]}}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"thinking","thinking":"hmm"}]}}`,
		`{"type":"user","message":{"role":"user","content":"<system-reminder>noise</system-reminder>second prompt"}}`,
		`not json at all`,
	}
	r := strings.NewReader(strings.Join(lines, "\n"))
	msgs, err := parseConversation(r)
	if err != nil {
		t.Fatalf("parseConversation returned error: %v", err)
	}
	want := []convoMessage{
		{Role: "user", Text: "hello"},
		{Role: "assistant", Text: "hi back"},
		{Role: "user", Text: "second prompt"},
	}
	if len(msgs) != len(want) {
		t.Fatalf("got %d messages, want %d: %+v", len(msgs), len(want), msgs)
	}
	for i, w := range want {
		if msgs[i] != w {
			t.Errorf("msg[%d] = %+v, want %+v", i, msgs[i], w)
		}
	}
}

func TestParseConversationEmpty(t *testing.T) {
	msgs, err := parseConversation(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected no messages, got %+v", msgs)
	}
}

func TestParseConversationDropsEmptyAssistantTurns(t *testing.T) {
	// An assistant turn that's pure tool-use / thinking with no text should not
	// produce a convoMessage at all.
	line := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"X"}]}}`
	msgs, err := parseConversation(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected no messages, got %+v", msgs)
	}
}
