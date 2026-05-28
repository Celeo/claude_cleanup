package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type screen int

const (
	screenList screen = iota
	screenViewport
	screenConfirm
)

type rootModel struct {
	screen screen

	list     list.Model
	viewport viewport.Model
	current  *Session // session being viewed / targeted for deletion

	width, height int
	statusMsg     string // ephemeral feedback (e.g. "deletion stubbed")
	confirmYes    bool   // selection in the confirm dialog (true = delete)
}

func newRootModel() (rootModel, error) {
	sessions, err := LoadSessions()
	if err != nil {
		return rootModel{}, err
	}

	items := make([]list.Item, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, sessionItem{Session: s})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Claude Code Conversations"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))

	return rootModel{
		screen:   screenList,
		list:     l,
		viewport: vp,
	}, nil
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		// leave room for a header + footer line around the viewport
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - 4)
		if m.current != nil {
			m.viewport.SetContent(renderConversation(m.current.Path, msg.Width))
		}
		return m, nil

	case tea.KeyPressMsg:
		// Global escape hatch.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenList:
		return m.updateList(msg)
	case screenViewport:
		return m.updateViewport(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	}
	return m, nil
}

func (m rootModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		// Don't intercept keys while the list's filter input is active.
		if m.list.FilterState() != list.Filtering {
			switch key.String() {
			case "q":
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(sessionItem); ok {
					s := item.Session
					m.current = &s
					m.viewport.SetContent(renderConversation(s.Path, m.width))
					m.viewport.GotoTop()
					m.screen = screenViewport
					m.statusMsg = ""
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m rootModel) updateViewport(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc", "backspace", "q", "h":
			m.screen = screenList
			return m, nil
		case "d", "delete":
			m.screen = screenConfirm
			m.confirmYes = false // default to safe option
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m rootModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "left", "right", "tab", "h", "l":
		m.confirmYes = !m.confirmYes
	case "y", "Y":
		m.confirmYes = true
	case "n", "N", "esc":
		m.screen = screenViewport
		return m, nil
	case "enter":
		if m.confirmYes && m.current != nil {
			deleteSession(*m.current)
			m.statusMsg = fmt.Sprintf("Stubbed deletion of %s", m.current.Title)
			m.current = nil
			m.screen = screenList
			return m, nil
		}
		m.screen = screenViewport
		return m, nil
	}
	return m, nil
}

func (m rootModel) View() tea.View {
	var s string
	switch m.screen {
	case screenViewport:
		s = m.viewView()
	case screenConfirm:
		s = m.confirmView()
	default:
		s = m.listView()
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m rootModel) listView() string {
	v := m.list.View()
	if m.statusMsg != "" {
		v = footerStyle.Render(m.statusMsg) + "\n" + v
	}
	return v
}

func (m rootModel) viewView() string {
	title := "untitled"
	if m.current != nil {
		title = m.current.Title
	}
	header := headerStyle.Width(m.width).Render(title)
	footer := footerStyle.Width(m.width).Render(
		"↑/↓ scroll · esc back · d delete · ctrl+c quit",
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), footer)
}

func (m rootModel) confirmView() string {
	title := "untitled"
	if m.current != nil {
		title = m.current.Title
	}

	yes := confirmDangerStyle.Render(" Delete ")
	no := confirmSafeStyle.Render(" Cancel ")
	if m.confirmYes {
		yes = confirmDangerFocused.Render(" Delete ")
	} else {
		no = confirmSafeFocused.Render(" Cancel ")
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		confirmTitleStyle.Render("Delete this conversation?"),
		"",
		lipgloss.NewStyle().Faint(true).Render(truncate(title, 60)),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center, no, "  ", yes),
		"",
		footerStyle.Render("←/→ switch · enter confirm · esc cancel"),
	)

	box := confirmBoxStyle.Render(body)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// renderConversation reads + formats a session file as a single string for
// display in the viewport. Wraps body text to the available width.
func renderConversation(path string, width int) string {
	msgs, err := LoadConversation(path)
	if err != nil {
		return fmt.Sprintf("error reading conversation: %v", err)
	}
	if len(msgs) == 0 {
		return "(no displayable messages in this conversation)"
	}

	wrap := width - 4
	if wrap < 20 {
		wrap = 20
	}
	body := lipgloss.NewStyle().Width(wrap).PaddingLeft(2)

	var b strings.Builder
	for i, msg := range msgs {
		var label string
		if msg.Role == "user" {
			label = userStyle.Render("▌ You")
		} else {
			label = assistantStyle.Render("▌ Claude")
		}
		b.WriteString(label)
		b.WriteByte('\n')
		b.WriteString(body.Render(msg.Text))
		if i < len(msgs)-1 {
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

// deleteSession is the stubbed deletion target. Wired up to a confirmation in
// the TUI but does not actually delete anything yet.
func deleteSession(s Session) {
	_ = s
}

func main() {
	m, err := newRootModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load sessions: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
