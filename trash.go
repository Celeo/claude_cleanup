package main

import (
	"errors"
	"fmt"
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

// EnsureTrashDir creates the trash root directory tree if it doesn't already
// exist. Called at startup so deletes can't fail just because the parent
// directories were never made, and so permission errors surface immediately
// rather than on the first delete attempt.
func EnsureTrashDir() error {
	root, err := trashRoot()
	if err != nil {
		return err
	}
	return os.MkdirAll(root, 0o755)
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

	sibling := filepath.Join(filepath.Dir(s.Path), s.ID)
	siblingDest := filepath.Join(destDir, s.ID)
	hasSibling := false
	if info, err := os.Stat(sibling); err == nil && info.IsDir() {
		hasSibling = true
		if _, err := os.Stat(siblingDest); err == nil {
			return errors.New("destination already exists: " + siblingDest)
		}
	}

	if err := os.Rename(s.Path, jsonlDest); err != nil {
		return err
	}
	if hasSibling {
		if err := os.Rename(sibling, siblingDest); err != nil {
			// Roll the transcript rename back so we don't leave a half-moved session.
			if rbErr := os.Rename(jsonlDest, s.Path); rbErr != nil {
				return fmt.Errorf("sibling rename failed: %w; transcript rollback also failed: %v", err, rbErr)
			}
			return fmt.Errorf("sibling rename failed: %w", err)
		}
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
