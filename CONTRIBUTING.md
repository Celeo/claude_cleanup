# Contributing

Thanks for your interest in contributing! This is a small project, so the bar is light — just keep the basics green.

## Before you open a PR

Run all three from the repo root. None of them should print anything:

```sh
gofmt -l .
go vet ./...
go test ./...
```

- **`gofmt -l .`** lists files that need formatting. The expected output is empty. If anything shows up, run `gofmt -w .` to fix it in place.
- **`go vet ./...`** catches common mistakes the compiler doesn't.
- **`go test ./...`** runs the test suite. New code that's easy to test (pure functions, `io.Reader`-shaped parsers) should come with tests; see `sessions_test.go` and `conversation_test.go` for the style.

That's it — no linters beyond `go vet`, no required commit message format, no CLA.

## Project notes

Architecture, JSONL format details, and Bubble Tea v2 gotchas live in [`CLAUDE.md`](CLAUDE.md). It's written for an AI assistant, but it's the most concise tour of the codebase and worth reading before a non-trivial change.
