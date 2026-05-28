# claude_cleanup

[![CI](https://github.com/celeo/claude_cleanup/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/celeo/claude_cleanup/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/celeo/claude_cleanup)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-FF75B7)](https://github.com/charmbracelet/bubbletea)
[![Made with Charm](https://img.shields.io/badge/made%20with-Charm-pink)](https://charm.sh)
[![Go Report Card](https://goreportcard.com/badge/github.com/celeo/claude_cleanup)](https://goreportcard.com/report/github.com/celeo/claude_cleanup)

A terminal UI for browsing and cleaning up old [Claude Code](https://claude.com/claude-code) conversations.

Claude Code stores every session as a `.jsonl` file under `~/.claude/projects/`. Over time those pile up. `claude_cleanup` lists them, lets you scroll through any conversation in the terminal, and lets you move the ones you don't want into a trash directory (with restore).

## Features

- Lists every Claude Code session on disk, newest first, with the same title Claude Code itself shows in `/recent` (custom title, then AI-generated title, then last prompt, then first prompt).
- Built-in fuzzy filter over titles — press `/` and start typing.
- Scrollable viewport of the conversation, with tool calls, internal thinking, and command noise stripped out so it reads like the chat you actually had.
- Two modes you can toggle with `t`: a red **delete** mode that moves a conversation into the trash, and a green **trash** mode that lists trashed conversations and lets you restore any of them.
- Confirmation prompt before either action.
- Non-destructive: "deletion" is really a rename into `~/.local/share/claude_cleanup/trash/`, preserving the original project layout so a restore puts the file (and any sibling overflow directory) back exactly where Claude Code expects it.

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
| **List** | `↑`/`↓` move, `/` filter, `enter` open, `t` toggle delete/trash mode, `q` quit |
| **Conversation** (delete mode) | `↑`/`↓`/`pgup`/`pgdn` scroll, `esc` back, `d` delete |
| **Conversation** (trash mode) | `↑`/`↓`/`pgup`/`pgdn` scroll, `esc` back, `r` restore |
| **Confirm** | `←`/`→` or `tab` toggle, `enter` confirm, `esc`/`n` cancel |

`ctrl+c` quits from anywhere. The list opens in delete mode (red accents); press `t` to switch to trash mode (green accents) and back.

## Where deleted conversations go

Deletes are non-destructive renames into:

```
~/.local/share/claude_cleanup/trash/<encoded-project>/<sessionId>.jsonl
```

The directory layout mirrors `~/.claude/projects/`, so a restore is a plain `rename` back to the original location. If the session has a sibling overflow directory (`<sessionId>/`, used by Claude Code to spill large tool outputs), it moves with the transcript.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — list + viewport components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — styling

## Contributing

Issues and PRs welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the (short) checklist.

## License

MIT — see [LICENSE](LICENSE).

## AI Tooling

This tool was entirely written through Claude, my first time doing so. Besides this little bit of the README and copy-pasting the LICENSE, I didn't write any of this program.
