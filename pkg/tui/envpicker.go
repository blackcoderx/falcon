package tui

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/storage"
	tea "github.com/charmbracelet/bubbletea"
)

// openEnvPicker initializes and opens the environment picker.
// Called when the user selects /env.
// Only shows environments that exist in .falcon/environments/.
func (m Model) openEnvPicker() Model {
	envs, err := storage.ListEnvironments(core.FalconFolderName)
	if err != nil || len(envs) == 0 {
		m.logs = append(m.logs, logEntry{
			Type:    "system",
			Content: "No environments found in .falcon/environments/. Create one by asking the agent or adding a YAML file there.",
		})
		m.updateViewportContent()
		return m
	}

	m.envPickerActive = true
	m.envPickerItems = envs
	m.envPickerIdx = 0

	// Pre-select the currently active environment.
	for i, name := range envs {
		if name == m.currentEnv {
			m.envPickerIdx = i
			break
		}
	}

	return m
}

// handleEnvPickerKeys processes keyboard input for the environment picker.
// Single-step: navigate with up/down, confirm with enter, cancel with esc.
func (m Model) handleEnvPickerKeys(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	if !m.envPickerActive {
		return false, m, nil
	}

	switch msg.String() {
	case "up", "shift+tab":
		if m.envPickerIdx > 0 {
			m.envPickerIdx--
		} else {
			m.envPickerIdx = len(m.envPickerItems) - 1
		}
		return true, m, nil
	case "down", "tab":
		if m.envPickerIdx < len(m.envPickerItems)-1 {
			m.envPickerIdx++
		} else {
			m.envPickerIdx = 0
		}
		return true, m, nil
	case "enter":
		m = m.applyEnvSwitch()
		return true, m, m.spinner.Tick
	case "esc":
		m.envPickerActive = false
		return true, m, nil
	}

	return false, m, nil
}

// applyEnvSwitch activates the selected environment via the shared PersistenceManager.
func (m Model) applyEnvSwitch() Model {
	if len(m.envPickerItems) == 0 || m.envPickerIdx >= len(m.envPickerItems) {
		m.envPickerActive = false
		return m
	}

	selected := m.envPickerItems[m.envPickerIdx]

	if m.persistManager == nil {
		m.logs = append(m.logs, logEntry{
			Type:    "error",
			Content: "Environment manager not available.",
		})
		m.envPickerActive = false
		m.updateViewportContent()
		return m
	}

	if err := m.persistManager.SetEnvironment(selected); err != nil {
		m.logs = append(m.logs, logEntry{
			Type:    "error",
			Content: fmt.Sprintf("Failed to switch environment: %v", err),
		})
		m.envPickerActive = false
		m.updateViewportContent()
		return m
	}

	m.currentEnv = selected
	m.envPickerActive = false

	m.logs = append(m.logs, logEntry{
		Type:    "system",
		Content: fmt.Sprintf("Environment switched to '%s'", selected),
	})
	m.updateViewportContent()
	return m
}

// renderEnvPicker renders the environment picker panel above the input.
func (m Model) renderEnvPicker() string {
	if !m.envPickerActive {
		return ""
	}

	var lines []string
	header := SlashItemStyle.Render("  Switch environment (↑↓ navigate, enter select, esc cancel)")
	lines = append(lines, header)

	for i, name := range m.envPickerItems {
		label := "  " + name
		if name == m.currentEnv {
			label += "  (active)"
		}
		if i == m.envPickerIdx {
			lines = append(lines, SlashItemSelectedStyle.Render(label))
		} else {
			lines = append(lines, SlashItemStyle.Render(label))
		}
	}

	return SlashPanelStyle.Render(strings.Join(lines, "\n"))
}

// envPickerHeight returns the rendered height of the environment picker panel.
func (m Model) envPickerHeight() int {
	if !m.envPickerActive {
		return 0
	}
	return len(m.envPickerItems) + 2 // header + items + padding
}
