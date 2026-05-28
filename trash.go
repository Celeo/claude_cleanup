package main

import (
	"errors"
	"os"
	"path/filepath"
)

// Trashed sessions live under ~/.local/share/claude_cleanup/trash/, mirroring
// the <encoded-project>/<sessionId>.jsonl layout that Claude Code uses under
// ~/.claude/projects/. A sibling overflow directory (same UUID, no extension)
// is moved alongside the transcript when present.

func trashRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "claude_cleanup", "trash"), nil
}

func claudeProjectsRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// MoveToTrash relocates a session's transcript (and any sibling overflow dir)
// into the trash, preserving the encoded-project subdirectory.
func MoveToTrash(s Session) error {
	root, err := trashRoot()
	if err != nil {
		return err
	}
	return moveSession(s, root)
}

// RestoreFromTrash moves a trashed session back to ~/.claude/projects/.
func RestoreFromTrash(s Session) error {
	root, err := claudeProjectsRoot()
	if err != nil {
		return err
	}
	return moveSession(s, root)
}

func moveSession(s Session, destRoot string) error {
	encoded := filepath.Base(filepath.Dir(s.Path))
	destDir := filepath.Join(destRoot, encoded)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	jsonlDest := filepath.Join(destDir, s.ID+".jsonl")
	if _, err := os.Stat(jsonlDest); err == nil {
		return errors.New("destination already exists: " + jsonlDest)
	}
	if err := os.Rename(s.Path, jsonlDest); err != nil {
		return err
	}
	sibling := filepath.Join(filepath.Dir(s.Path), s.ID)
	if info, err := os.Stat(sibling); err == nil && info.IsDir() {
		_ = os.Rename(sibling, filepath.Join(destDir, s.ID))
	}
	return nil
}

// LoadTrash walks the trash dir and returns sessions in the same shape as
// LoadSessions. An absent trash dir is not an error — it just means nothing's
// been deleted yet.
func LoadTrash() ([]Session, error) {
	root, err := trashRoot()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return loadSessionsFromRoot(root)
}
