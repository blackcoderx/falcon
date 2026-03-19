package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
)

// SlashCommand represents a command available via "/" prefix
type SlashCommand struct {
	Name        string // "model", flow filename, or request filename
	Description string
	Kind        string // "builtin" | "flow" | "request"
}

// SlashState tracks current slash command panel state
type SlashState struct {
	Active      bool
	Query       string
	Suggestions []SlashCommand
	Selected    int
	FlowContent string         // loaded file content (after selection, cleared after enter)
	TaggedFile  string         // filename of the tagged file (shown as chip until message is sent)
	cachedAll   []SlashCommand // cached full list (builtins + files), populated once
}

// listBuiltinCommands returns the list of built-in slash commands.
func listBuiltinCommands() []SlashCommand {
	return []SlashCommand{
		{Name: "model", Description: "Switch LLM provider/model", Kind: "builtin"},
		{Name: "env", Description: "Switch active environment", Kind: "builtin"},
	}
}

// listFlowFiles reads .falcon/flows/ for *.yaml/*.yml files.
func listFlowFiles(falconDir string) []SlashCommand {
	dir := filepath.Join(falconDir, "flows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var cmds []SlashCommand
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			cmds = append(cmds, SlashCommand{
				Name:        name,
				Description: "Flow",
				Kind:        "flow",
			})
		}
	}
	return cmds
}

// listRequestFiles reads .falcon/requests/ for *.yaml/*.yml files.
func listRequestFiles(falconDir string) []SlashCommand {
	dir := filepath.Join(falconDir, "requests")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var cmds []SlashCommand
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			cmds = append(cmds, SlashCommand{
				Name:        name,
				Description: "Request",
				Kind:        "request",
			})
		}
	}
	return cmds
}

// filterCommands returns commands whose names contain the query string (case-insensitive).
func filterCommands(all []SlashCommand, query string) []SlashCommand {
	if query == "" {
		return all
	}
	lq := strings.ToLower(query)
	var result []SlashCommand
	for _, cmd := range all {
		if strings.Contains(strings.ToLower(cmd.Name), lq) {
			result = append(result, cmd)
		}
	}
	return result
}

// refreshSlashCache builds the full unfiltered list of slash commands (builtins + flows + requests).
// This is called once per panel activation to avoid repeated I/O on every keystroke.
func refreshSlashCache(falconDir string) []SlashCommand {
	all := listBuiltinCommands()
	all = append(all, listFlowFiles(falconDir)...)
	all = append(all, listRequestFiles(falconDir)...)
	return all
}

// updateSlashState rebuilds the suggestion list filtered by the given query.
// It only reads the filesystem on first activation (cache miss); subsequent
// keystrokes filter the in-memory cache, keeping Update() non-blocking.
func (m Model) updateSlashState(query string) Model {
	if !m.slashState.Active || m.slashState.cachedAll == nil {
		// Build cache on first activation
		m.slashState.cachedAll = refreshSlashCache(core.FalconFolderName)
	}

	filtered := filterCommands(m.slashState.cachedAll, query)

	m.slashState.Active = true
	m.slashState.Query = query
	m.slashState.Suggestions = filtered

	// Clamp selected index to valid range
	if m.slashState.Selected >= len(filtered) {
		m.slashState.Selected = max(0, len(filtered)-1)
	}

	return m
}

// acceptSlashCommand processes the currently selected slash command.
func (m Model) acceptSlashCommand() (Model, tea.Cmd) {
	if len(m.slashState.Suggestions) == 0 {
		return m, nil
	}

	selected := m.slashState.Suggestions[m.slashState.Selected]

	switch selected.Kind {
	case "builtin":
		if selected.Name == "model" {
			m = m.openModelPicker()
			m.slashState = SlashState{}
			m.textinput.SetValue("")
		} else if selected.Name == "env" {
			m = m.openEnvPicker()
			m.slashState = SlashState{}
			m.textinput.SetValue("")
		}

	case "flow", "request":
		// Determine file path
		var subDir string
		if selected.Kind == "flow" {
			subDir = "flows"
		} else {
			subDir = "requests"
		}
		filePath := filepath.Join(core.FalconFolderName, subDir, selected.Name)

		content, err := os.ReadFile(filePath)
		if err == nil {
			m.slashState.FlowContent = string(content)
			m.slashState.TaggedFile = selected.Name
		}

		// Extract remainder text after the slash command in the input
		currentInput := m.textinput.Value()
		prefix := "/" + selected.Name
		remainder := ""
		if strings.HasPrefix(currentInput, prefix) {
			rest := currentInput[len(prefix):]
			remainder = strings.TrimLeft(rest, " ")
		}

		m.textinput.SetValue(remainder)
		m.textinput.CursorEnd()
		m.slashState.Active = false
		m.slashState.Query = ""
		m.slashState.Suggestions = nil
	}

	return m, nil
}

// handleSlashKeys intercepts key presses when the slash panel is active.
// Returns (handled bool, updated model, command).
// When a key is consumed, a non-nil cmd is always returned so that Update()
// takes the early-return path and does not forward the key to textinput.
func (m Model) handleSlashKeys(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	switch msg.String() {
	case "up", "shift+tab":
		if len(m.slashState.Suggestions) > 0 {
			if m.slashState.Selected == 0 {
				m.slashState.Selected = len(m.slashState.Suggestions) - 1
			} else {
				m.slashState.Selected--
			}
		}
		return true, m, m.spinner.Tick

	case "down", "tab":
		if len(m.slashState.Suggestions) > 0 {
			if m.slashState.Selected >= len(m.slashState.Suggestions)-1 {
				m.slashState.Selected = 0
			} else {
				m.slashState.Selected++
			}
		}
		return true, m, m.spinner.Tick

	case "enter":
		updated, cmd := m.acceptSlashCommand()
		if cmd == nil {
			cmd = m.spinner.Tick
		}
		return true, updated, cmd

	case "esc":
		// Close the panel but leave the typed text in the input intact.
		// cachedAll is cleared via SlashState{} so the next '/' re-reads the filesystem.
		m.slashState = SlashState{}
		return true, m, m.spinner.Tick

	default:
		return false, m, nil
	}
}

// slashPanelHeight returns the height the slash panel will occupy.
func (m Model) slashPanelHeight() int {
	if !m.slashState.Active || len(m.slashState.Suggestions) == 0 {
		return 0
	}
	count := len(m.slashState.Suggestions)
	if count > 6 {
		count = 6
	}
	return count + 1 // +1 for margin/padding (SlashPanelStyle has no borders)
}

// renderSlashPanel renders the slash command suggestion panel.
func (m Model) renderSlashPanel() string {
	suggestions := m.slashState.Suggestions
	if len(suggestions) == 0 {
		return ""
	}

	// Determine the window of items to show (max 6)
	start := 0
	end := len(suggestions)
	if end > 6 {
		end = 6
		// If selected is beyond the initial window, adjust
		if m.slashState.Selected >= 6 {
			start = m.slashState.Selected - 5
			end = m.slashState.Selected + 1
		}
	}

	visible := suggestions[start:end]

	var lines []string
	for i, cmd := range visible {
		actualIdx := start + i
		var line string
		if actualIdx == m.slashState.Selected {
			line = SlashItemSelectedStyle.Render("  /"+cmd.Name) + "  " + SlashItemKindStyle.Render(cmd.Description)
		} else {
			line = SlashItemStyle.Render("  /"+cmd.Name) + "  " + SlashItemKindStyle.Render(cmd.Description)
		}
		lines = append(lines, line)
	}

	panel := strings.Join(lines, "\n")
	return SlashPanelStyle.Render(panel)
}
