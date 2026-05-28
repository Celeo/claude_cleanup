# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A Go TUI for browsing and cleaning up Claude Code conversation history. Reads the JSONL session files Claude Code writes under `~/.claude/projects/<encoded-cwd>/<sessionId>.jsonl`.

## Commands

- Build: `go build ./...` (produces `./claude_cleanup`)
- Run: `go run .`
- Vet: `go vet ./...`
- Test: `go test ./...`

Tests cover the pure helpers (`truncate`, `decodeProjectDir`, `stripCommandTags`, `extractText`) and the reader-shaped cores of session/conversation parsing (`pickTitleFromReader`, `parseConversation`). The filesystem-touching wrappers (`LoadSessions`, `LoadTrash`, `moveSession`) are intentionally untested — keeping them thin lets the tested cores carry the behavioral coverage without needing a faked filesystem. There is no lint config beyond `go vet`.

## Architecture

Single `main` package, split by responsibility — not by layer. The TUI is built on the **v2** Charm libraries (`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`), which differ from v1 in a few load-bearing ways (see "v2 gotchas" below).

- `main.go` — `rootModel` is a screen router. It owns a `screen` enum (`screenList` / `screenViewport` / `screenConfirm`) and a `mode` enum (`modeDelete` / `modeTrash`) and dispatches `Update` to a per-screen handler. Child components (`list.Model`, `viewport.Model`) are embedded directly on the root rather than wrapped in their own `tea.Model` types, because they don't need independent message routing. `reloadList` rebuilds list items + title + title style based on the current mode; it's called on construction, on `t` toggle, and after each successful trash/restore. The confirm dialog's `enter` handler dispatches to `MoveToTrash` or `RestoreFromTrash` based on `m.mode`. Action / reload errors set `rootModel.fatalErr` and return `tea.Quit`; `main()` prints the error to stderr and exits non-zero on the way out — we deliberately don't show those in the footer, since a hidden error easily reads as "nothing happened."
- `sessions.go` — discovers sessions by walking a root that uses the `<encoded-project>/<sessionId>.jsonl` layout. `LoadSessions` walks `~/.claude/projects/`; `loadSessionsFromRoot` is the shared walker that `LoadTrash` also uses. Title precedence is **custom-title > ai-title > last-prompt > first user prompt > session ID**; these come from dedicated single-line JSONL records (`{"type":"custom-title",...}` etc.), not from message content. `sessionItem` wraps `Session` only because the `Title` field collides with the `Title()` method the bubbles list delegate requires.
- `conversation.go` — parses the JSONL again for display. Only `user` / `assistant` records with non-empty text survive; `isMeta`, `isSidechain`, `tool_use`, `tool_result`, and `thinking` blocks are dropped. `message.content` is polymorphic (string OR array of typed blocks) — `extractText` handles both. `stripCommandTags` removes the `<system-reminder>`, `<local-command-*>`, and `<command-*>` wrappers that Claude Code injects into user turns so the rendered conversation matches what the user actually saw.
- `trash.go` — owns the move-to-trash / restore-from-trash flow. The trash lives at `~/.local/share/claude_cleanup/trash/`, mirroring Claude Code's `<encoded-project>/<sessionId>.jsonl` layout so restore is a plain rename back. `EnsureTrashDir` is called from `main()` before the TUI starts so the root exists (and any permission problem surfaces immediately). `moveSession` is the shared core: it pre-checks both destinations (refuses to overwrite either), renames the `.jsonl`, then renames the sibling overflow directory (`<sessionId>/`, when present); if the sibling rename fails, it rolls the transcript rename back so a session never ends up half-moved.
- `styles.go` — all lipgloss styles live here. Two mode-themed palettes (`deleteAccent` red, `restoreAccent` green) drive the list title bar, viewport header, confirm dialog border/title/action button, and footer hints. The Cancel button stays neutral; only the affirmative button changes color per mode.

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

- Run `gofmt -w .` as a fit-and-finish step after you're done responding to a request that touched `.go` files. Don't bother reformatting between individual edits — just once at the end so the working tree stays formatted before the user runs `gofmt -l .` themselves.
- Project directory names under `~/.claude/projects/` are the original cwd with `/` replaced by `-`. The decode in `decodeProjectDir` is best-effort (lossy: real dashes in the original path are indistinguishable from separators) — don't rely on it for anything beyond display.
- A session is a `.jsonl` transcript and an optional sibling directory named after the session UUID (no extension) that holds spilled tool outputs. Anything that moves, deletes, or copies a session has to handle both — see `moveSession` in `trash.go` for the pattern.
- Deletion is non-destructive: it's a rename into the trash dir, never `os.Remove`. The trash dir's layout matches `~/.claude/projects/` precisely so restore is a symmetric rename back.
- `moveSession` uses `os.Rename`, which assumes `~/.claude/` and `~/.local/share/` are on the same filesystem. That's true on a standard single-mount home, but unusual setups (separate `/home` or `/var` partitions, some encrypted-home configurations) can fail with `EXDEV`. We don't currently have a copy-then-remove fallback — if a user reports it, that's the fix.
