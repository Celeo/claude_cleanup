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

type mode int

const (
	modeDelete mode = iota
	modeTrash
)

type rootModel struct {
	screen screen
	mode   mode

	list     list.Model
	viewport viewport.Model
	current  *Session // session being viewed / targeted

	width, height int
	statusMsg     string
	confirmYes    bool // selection in the confirm dialog (true = act)

	// fatalErr is set when an action (move/restore/reload) fails. When set,
	// the Update returns tea.Quit and main() prints the error to stderr and
	// exits non-zero — we never want a failed delete to look like a successful
	// one in the UI.
	fatalErr error
}

func newRootModel() (rootModel, error) {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	vp := viewport.New()

	m := rootModel{
		screen:   screenList,
		mode:     modeDelete,
		list:     l,
		viewport: vp,
	}
	if err := m.reloadList(); err != nil {
		return rootModel{}, err
	}
	return m, nil
}

// reloadList re-reads sessions for the current mode and updates the list's
// items, title, and title styling.
func (m *rootModel) reloadList() error {
	var (
		sessions []Session
		err      error
	)
	switch m.mode {
	case modeTrash:
		sessions, err = LoadTrash()
		m.list.Title = "Trash — Restore Deleted Conversations"
		m.list.Styles.Title = listTitleRestoreStyle
	default:
		sessions, err = LoadSessions()
		m.list.Title = "Claude Code Conversations"
		m.list.Styles.Title = listTitleDeleteStyle
	}
	if err != nil {
		return err
	}
	items := make([]list.Item, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, sessionItem{Session: s})
	}
	m.list.SetItems(items)
	return nil
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// leave one line under the list for the mode-aware key hint
		m.list.SetSize(msg.Width, msg.Height-1)
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - 4)
		if m.current != nil {
			m.viewport.SetContent(renderConversation(m.current.Path, msg.Width))
		}
		return m, nil

	case tea.KeyPressMsg:
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
		// Suppress single-letter hotkeys while the user is typing into the filter,
		// otherwise "t" / "q" / "enter" would steal keystrokes from the filter input.
		if m.list.FilterState() != list.Filtering {
			switch key.String() {
			case "q":
				return m, tea.Quit
			case "t":
				if m.mode == modeDelete {
					m.mode = modeTrash
				} else {
					m.mode = modeDelete
				}
				m.statusMsg = ""
				if err := m.reloadList(); err != nil {
					m.statusMsg = "failed to reload: " + err.Error()
				}
				return m, nil
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
		case "esc", "backspace":
			m.screen = screenList
			return m, nil
		case "d", "delete":
			if m.mode == modeDelete {
				m.screen = screenConfirm
				m.confirmYes = false
				return m, nil
			}
		case "r":
			if m.mode == modeTrash {
				m.screen = screenConfirm
				m.confirmYes = false
				return m, nil
			}
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
			var (
				err   error
				okMsg string
			)
			if m.mode == modeTrash {
				err = RestoreFromTrash(*m.current)
				okMsg = fmt.Sprintf("Restored %s", m.current.Title)
			} else {
				err = MoveToTrash(*m.current)
				okMsg = fmt.Sprintf("Moved %s to trash", m.current.Title)
			}
			if err != nil {
				m.fatalErr = err
				return m, tea.Quit
			}
			m.statusMsg = okMsg
			m.current = nil
			m.screen = screenList
			if reloadErr := m.reloadList(); reloadErr != nil {
				m.fatalErr = reloadErr
				return m, tea.Quit
			}
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
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m rootModel) listView() string {
	listView := m.list.View()
	hint := m.listFooterHint()
	footer := footerStyle.Render(hint)
	if m.statusMsg != "" {
		footer = footerStyle.Render(m.statusMsg + "  ·  " + hint)
	}
	return lipgloss.JoinVertical(lipgloss.Left, listView, footer)
}

func (m rootModel) listFooterHint() string {
	if m.mode == modeTrash {
		return "enter open · t back to active · / filter · q quit  [TRASH]"
	}
	return "enter open · t view trash · / filter · q quit  [DELETE]"
}

func (m rootModel) viewView() string {
	title := "untitled"
	if m.current != nil {
		title = m.current.Title
	}
	var header, footer string
	if m.mode == modeTrash {
		header = headerRestoreStyle.Width(m.width).Render("[TRASH] " + title)
		footer = footerStyle.Width(m.width).Render(
			"↑/↓ scroll · esc back · r restore · ctrl+c quit",
		)
	} else {
		header = headerDeleteStyle.Width(m.width).Render(title)
		footer = footerStyle.Width(m.width).Render(
			"↑/↓ scroll · esc back · d delete · ctrl+c quit",
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), footer)
}

func (m rootModel) confirmView() string {
	title := "untitled"
	if m.current != nil {
		title = m.current.Title
	}

	var (
		prompt     string
		actionText string
		boxStyle   lipgloss.Style
		titleStyle lipgloss.Style
		actStyle   lipgloss.Style
		actFocused lipgloss.Style
	)
	if m.mode == modeTrash {
		prompt = "Restore this conversation?"
		actionText = " Restore "
		boxStyle = confirmBoxRestoreStyle
		titleStyle = confirmTitleRestoreStyle
		actStyle = confirmRestoreActionStyle
		actFocused = confirmRestoreActionFocused
	} else {
		prompt = "Move this conversation to the trash?"
		actionText = " Delete "
		boxStyle = confirmBoxDeleteStyle
		titleStyle = confirmTitleDeleteStyle
		actStyle = confirmDeleteActionStyle
		actFocused = confirmDeleteActionFocused
	}

	yes := actStyle.Render(actionText)
	no := confirmCancelStyle.Render(" Cancel ")
	if m.confirmYes {
		yes = actFocused.Render(actionText)
	} else {
		no = confirmCancelFocused.Render(" Cancel ")
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(prompt),
		"",
		lipgloss.NewStyle().Faint(true).Render(truncate(title, 60)),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center, no, "  ", yes),
		"",
		footerStyle.Render("←/→ switch · enter confirm · esc cancel"),
	)

	box := boxStyle.Render(body)
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

func main() {
	if err := EnsureTrashDir(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create trash directory: %v\n", err)
		os.Exit(1)
	}
	m, err := newRootModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load sessions: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
	if rm, ok := final.(rootModel); ok && rm.fatalErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", rm.fatalErr)
		os.Exit(1)
	}
}
