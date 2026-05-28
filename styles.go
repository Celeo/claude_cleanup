package main

import "charm.land/lipgloss/v2"

var (
	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DD3FC"))

	assistantStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA"))

	footerStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	// Mode-themed accents. Delete mode is red; trash/restore mode is green.
	deleteAccent  = lipgloss.Color("#F87171")
	deleteBgFocus = lipgloss.Color("#EF4444")
	deleteBg      = lipgloss.Color("#DC2626")

	restoreAccent  = lipgloss.Color("#34D399")
	restoreBgFocus = lipgloss.Color("#10B981")
	restoreBg      = lipgloss.Color("#059669")

	headerDeleteStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(deleteAccent).
				Padding(0, 1)

	headerRestoreStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(restoreAccent).
				Padding(0, 1)

	listTitleDeleteStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(deleteBg).
				Padding(0, 1)

	listTitleRestoreStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(restoreBg).
				Padding(0, 1)

	// Confirm dialog — delete (red) variants.
	confirmBoxDeleteStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(deleteAccent).
				Padding(1, 3).
				Align(lipgloss.Center)

	confirmTitleDeleteStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(deleteAccent)

	confirmDeleteActionStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(deleteBg).
					Padding(0, 2)

	confirmDeleteActionFocused = confirmDeleteActionStyle.
					Background(deleteBgFocus).
					Underline(true)

	// Confirm dialog — restore (green) variants.
	confirmBoxRestoreStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(restoreAccent).
				Padding(1, 3).
				Align(lipgloss.Center)

	confirmTitleRestoreStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(restoreAccent)

	confirmRestoreActionStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(restoreBg).
					Padding(0, 2)

	confirmRestoreActionFocused = confirmRestoreActionStyle.
					Background(restoreBgFocus).
					Underline(true)

	// Shared "Cancel" (neutral) button for both confirm dialogs.
	confirmCancelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Padding(0, 2)

	confirmCancelFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4B5563")).
				Padding(0, 2)
)
