# claude_cleanup

[![CI](https://github.com/celeo/claude_cleanup/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/celeo/claude_cleanup/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/celeo/claude_cleanup)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-FF75B7)](https://github.com/charmbracelet/bubbletea)
[![Made with Charm](https://img.shields.io/badge/made%20with-Charm-pink)](https://charm.sh)
[![Go Report Card](https://goreportcard.com/badge/github.com/celeo/claude_cleanup)](https://goreportcard.com/report/github.com/celeo/claude_cleanup)

A terminal UI for browsing and cleaning up old [Claude Code](https://claude.com/claude-code) conversations.

Claude Code stores every session as a `.jsonl` file under `~/.claude/projects/`. Over time those pile up. `claude_cleanup` lists them, lets you scroll through any conversation in the terminal, and (eventually) lets you delete the ones you don't want anymore.

## Features

- Lists every Claude Code session on disk, newest first, with the same title Claude Code itself shows in `/recent` (custom title, then AI-generated title, then last prompt, then first prompt).
- Built-in fuzzy filter over titles — press `/` and start typing.
- Scrollable viewport of the conversation, with tool calls, internal thinking, and command noise stripped out so it reads like the chat you actually had.
- Confirmation prompt before deletion. (Deletion itself is currently stubbed — see [Status](#status).)

## Install

Requires Go 1.26 or newer.

```sh
git clone https://github.com/celeo/claude_cleanup
cd claude_cleanup
go build ./...
./claude_cleanup
```

Or run directly:

```sh
go run .
```

## Usage

Launch it and you get three screens:

| Screen | Keys |
| --- | --- |
| **List** | `↑`/`↓` move, `/` filter, `enter` open, `q` quit |
| **Conversation** | `↑`/`↓`/`pgup`/`pgdn` scroll, `esc` back, `d` delete |
| **Confirm** | `←`/`→` or `tab` toggle, `enter` confirm, `esc`/`n` cancel |

`ctrl+c` quits from anywhere.

## Status

Early. The TUI works end-to-end, but the deletion step is a stub — confirming a delete currently just clears the selection and returns you to the list. Wiring up the actual `os.Remove` is the next step.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — list + viewport components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — styling

## Contributing

Issues and PRs welcome.

## License

MIT — see [LICENSE](LICENSE).
