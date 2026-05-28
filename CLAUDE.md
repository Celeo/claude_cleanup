# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A Go TUI for browsing and cleaning up Claude Code conversation history. Reads the JSONL session files Claude Code writes under `~/.claude/projects/<encoded-cwd>/<sessionId>.jsonl`.

## Commands

- Build: `go build ./...` (produces `./claude_cleanup`)
- Run: `go run .`
- Vet: `go vet ./...`

There are no tests yet. There is no lint config beyond `go vet`.

## Architecture

Single `main` package, split by responsibility — not by layer. The TUI is built on the **v2** Charm libraries (`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`), which differ from v1 in a few load-bearing ways (see "v2 gotchas" below).

- `main.go` — `rootModel` is a screen router. It owns a `screen` enum (`screenList` / `screenViewport` / `screenConfirm`) and dispatches `Update` to a per-screen handler. Child components (`list.Model`, `viewport.Model`) are embedded directly on the root rather than wrapped in their own `tea.Model` types, because they don't need independent message routing. `deleteSession` is the stub the confirm dialog calls — wire real deletion here.
- `sessions.go` — discovers sessions by walking `~/.claude/projects/*/*.jsonl`. Title precedence is **custom-title > ai-title > last-prompt > first user prompt > session ID**; these come from dedicated single-line JSONL records (`{"type":"custom-title",...}` etc.), not from message content. `sessionItem` wraps `Session` only because the `Title` field collides with the `Title()` method the bubbles list delegate requires.
- `conversation.go` — parses the JSONL again for display. Only `user` / `assistant` records with non-empty text survive; `isMeta`, `isSidechain`, `tool_use`, `tool_result`, and `thinking` blocks are dropped. `message.content` is polymorphic (string OR array of typed blocks) — `extractText` handles both. `stripCommandTags` removes the `<system-reminder>`, `<local-command-*>`, and `<command-*>` wrappers that Claude Code injects into user turns so the rendered conversation matches what the user actually saw.
- `styles.go` — all lipgloss styles live here.

## JSONL session format notes

These are the only record types currently consumed; the file contains many more (`mode`, `permission-mode`, `file-history-snapshot`, `attachment`, `agent-name`, etc.) that are intentionally ignored:

- `custom-title` / `ai-title` / `last-prompt` — title metadata records, one field each.
- `user` / `assistant` — the actual turns. `message.content` is either a string (typical for user) or an array of blocks `[{"type":"text","text":"..."}]` (typical for assistant).
- `isMeta: true` marks system-injected user turns (e.g. the initial caveat) — skip.
- `isSidechain: true` marks subagent turns — skip when rendering the main thread.

## v2 gotchas

- `Model.View()` returns `tea.View`, not `string`. Wrap with `tea.NewView(s)`.
- Alt-screen is **not** a `tea.NewProgram` option in v2. Set `v.AltScreen = true` on the returned `tea.View` instead. `tea.WithAltScreen()` does not exist.
- Key messages are `tea.KeyPressMsg`, not `tea.KeyMsg`.
- `viewport.New` takes functional options (`viewport.WithWidth`, `WithHeight`), not positional args.

## Version control

This repo is managed with [jj](https://github.com/jj-vcs/jj) on top of git. The git remote and branch names still work for CI and GitHub URLs, but day-to-day commands the user runs are `jj` (`jj st`, `jj new`, `jj describe`, etc.) — don't suggest raw `git commit` / `git rebase` workflows.

## Conventions

- Project directory names under `~/.claude/projects/` are the original cwd with `/` replaced by `-`. The decode in `decodeProjectDir` is best-effort (lossy: real dashes in the original path are indistinguishable from separators) — don't rely on it for anything beyond display.
