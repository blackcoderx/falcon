package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#FF6B9D")
	secondaryColor = lipgloss.Color("#C792EA")
	accentColor    = lipgloss.Color("#89DDFF")
	bgColor        = lipgloss.Color("#1E1E2E")
	textColor      = lipgloss.Color("#CDD6F4")
	mutedColor     = lipgloss.Color("#6C7086")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(0, 2, 1, 2)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1).
			Width(80)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 2)
)

type model struct {
	input       string
	messages    []string
	thinking    bool
	err         error
	width       int
	height      int
	cursorBlink bool
}

func initialModel() model {
	return model{
		input:    "",
		messages: []string{},
		thinking: false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.input != "" {
				m.messages = append(m.messages, fmt.Sprintf("You: %s", m.input))
				m.input = ""
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	// Header
	header := titleStyle.Render("⚡ ZAP") + "\n" +
		subtitleStyle.Render("AI-powered API testing in your terminal")

	// Messages area
	messagesView := ""
	if len(m.messages) == 0 {
		messagesView = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Render("No messages yet. Start by typing a command...")
	} else {
		for _, msg := range m.messages {
			messagesView += lipgloss.NewStyle().
				Foreground(textColor).
				Render(msg) + "\n"
		}
	}

	// Input area
	inputView := inputPromptStyle.Render("→ ") + m.input
	if m.cursorBlink {
		inputView += "▋"
	}

	// Help
	help := helpStyle.Render("ctrl+c or q to quit • enter to send")

	// Container
	content := containerStyle.Render(
		header + "\n\n" +
			messagesView + "\n\n" +
			inputView,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content+"\n"+help,
	)
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
