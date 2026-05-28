Here's the review. I'll skip nits and focus on things that meaningfully matter.

## Correctness & failure points

1. **`trash.go:64` — silent failure on sibling-dir rename.** After the `.jsonl` is moved, if the sibling overflow dir rename fails, the error is dropped and the session ends up half-trashed (transcript in trash, overflow still in projects). The user gets a "Moved to trash" message either way. At minimum, propagate the error; ideally roll the `.jsonl` rename back if the sibling fails.
2. **`trash.go:59` — `os.Rename` doesn't cross filesystems.** `~/.claude` and `~/.local/share` are usually on the same mount, but not always (separate `/home`, encrypted home, etc.). On `EXDEV` the user gets a confusing "invalid cross-device link" message and the session stays put. Worth either documenting the assumption or falling back to copy-then-remove.
3. **`sessions.go:44` — `LoadSessions` errors on missing `~/.claude/projects/`.** `LoadTrash` has a not-exist guard (`trash.go:77`); `LoadSessions` doesn't, so a clean machine with no Claude Code history crashes at startup. Mirror the guard.
4. **`main.go:142` — `enter` opens a viewport for the selected item, but doesn't guard against a nil selection on an empty list.** `list.SelectedItem()` returns nil on an empty list, the type assertion fails, the branch falls through to `m.list.Update(msg)` — so it's actually safe today. Worth a comment noting why we don't need an `else`, or just leaving it.
5. **`main.go:215` — status message reads `"Moved to trash: <title>"` vs `"Restored <title>"`** — the colon-vs-no-colon asymmetry is jarring. Trivial alignment fix.
6. **`main.go:165` — `q` and `h` as back-from-viewport.** `h` is fine for vim users but undocumented; `q` is documented as quit on the list screen and "back" here without being shown anywhere. Either add to the viewport footer hint or drop them — currently they're discoverable only by accident.
7. **`conversation.go:39` — bad JSON lines are silently swallowed.** Probably the right behavior (real session files do contain odd records) but if a file gets corrupted you'll see "(no displayable messages)" with no clue why. Not worth changing now, just flagging.

## Charm/Bubble Tea usage

- All v2 idioms look correct (`tea.View` return, `AltScreen` field, `tea.KeyPressMsg`, viewport functional options). No issues.
- **`main.go:48` — `viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))`** can just be `viewport.New()`. 0 is the default.
- **No mouse-wheel scrolling in the viewport.** Bubbletea has `tea.WithMouseCellMotion()`/`WithMouseAllMotion()` as `tea.NewProgram` options. A long conversation is the main thing the user scrolls; mouse support is a one-line win.
- **`main.go:102` — viewport height is `msg.Height - 4`** but header is 1 line and footer is 1 line, so 2 lines are unaccounted for. Probably intentional padding, but inconsistent with the list using `Height-1`. If you want symmetric: `Height - 2`, or rely on lipgloss to size the rendered output.
- **`reloadList` rewrites `m.list.Title` and `m.list.Styles.Title` every time mode toggles.** That's fine, but the title style assignment is a bit hidden in a load function. Pulling the mode-dependent strings/styles into a `modePalette` struct with `title`, `listTitleStyle`, `headerStyle`, `confirmBoxStyle`, etc. would centralize the dual paths now scattered across `reloadList`, `viewView`, and `confirmView`.

## Go style / small things

- **`styles.go:14` — `bodyStyle` is defined and unused.** `renderConversation` rebuilds an inline style instead. Either delete or use it.
- **`main.go:131-141` — toggle could be `m.mode = 1 - m.mode` or a small method, but the current if/else reads fine.** Skip.
- **`sessions.go:120` — `json.Unmarshal` of the whole anonymous struct on every line of every file.** Fine functionally; if profiles ever show this hot, the standard trick is to first decode just `{"type": "..."}` and skip the rest. Not worth doing preemptively.

## Easy testability extractions (no fs needed)

These are pure or `io.Reader`-shaped and would give you real coverage of the logic that matters without any tmp dirs:

- `truncate(s, n)` — `sessions.go:156`. Pure. Trivial cases: short, exactly n, longer, embedded newlines.
- `decodeProjectDir(encoded)` — `sessions.go:88`. Pure. Document the lossy round-trip in a test.
- `stripCommandTags(s)` — `conversation.go:101`. Pure. Tag soup → clean string.
- `extractText(raw)` — `conversation.go:62`. Pure. Plain-string content, block-array content, empty input, malformed.
- `listFooterHint()` — `main.go:255`. Trivially pure on `mode`.
- **`pickTitle` → split into `pickTitleFromReader(r io.Reader, fallback string) string`** at `sessions.go:97`. Keep the path-based `pickTitle` as a one-line wrapper that opens the file. Now you can test the precedence ladder (custom > ai > last > first > id) with `strings.NewReader`. This is the single highest-value extraction — that precedence is the title's whole reason for existing.
- **`LoadConversation` → split into `parseConversation(r io.Reader) ([]convoMessage, error)`** at `conversation.go:19`. Same shape as above. Lets you test sidechain/meta skipping, polymorphic content, tool-use filtering without touching disk.

What I'd *skip*: any "extract for tests" refactor of `loadSessionsFromRoot` or `moveSession`. Both are fundamentally filesystem operations — the only thing worth testing is the encoded-dir/sibling-dir handling, and you said you don't want a fake fs for that. Leave them.

## Suggested order

If you do anything, I'd queue it as:
1. Fix `trash.go` sibling-rename error handling (real bug).
2. Add the `LoadSessions` missing-dir guard (real bug, easy).
3. Extract `pickTitleFromReader` and `parseConversation`; write tests for those + the four pure helpers above. That's where most of the actual behavior lives.
4. Everything else is polish.
