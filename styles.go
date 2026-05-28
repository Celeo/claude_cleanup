package main

import "charm.land/lipgloss/v2"

var (
	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DD3FC"))

	assistantStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA"))

	bodyStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F472B6")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F87171")).
			Padding(1, 3).
			Align(lipgloss.Center)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F87171"))

	confirmDangerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#DC2626")).
				Padding(0, 2)

	confirmSafeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Padding(0, 2)

	confirmDangerFocused = confirmDangerStyle.
				Background(lipgloss.Color("#EF4444")).
				Underline(true)

	confirmSafeFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#10B981")).
				Padding(0, 2)
)
