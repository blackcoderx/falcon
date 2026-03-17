package tui

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/llm"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// openModelPicker initializes and opens the model picker.
// Called when the user selects /model.
func (m Model) openModelPicker() Model {
	m.modelPickerActive = true
	m.modelPickerStep = 0
	m.modelPickerItems = llm.All()
	m.modelPickerIdx = 0

	// Pre-select current provider if found
	currentProvider := "" // read from viper or current client
	for i, p := range m.modelPickerItems {
		if p.ID() == currentProvider {
			m.modelPickerIdx = i
			break
		}
	}

	return m
}

// handleModelPickerKeys processes keyboard input for the model picker.
func (m Model) handleModelPickerKeys(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	if !m.modelPickerActive {
		return false, m, nil
	}

	switch m.modelPickerStep {
	case 0: // Provider selection step
		switch msg.String() {
		case "up", "shift+tab":
			if m.modelPickerIdx > 0 {
				m.modelPickerIdx--
			} else {
				m.modelPickerIdx = len(m.modelPickerItems) - 1
			}
			return true, m, m.spinner.Tick
		case "down", "tab":
			if m.modelPickerIdx < len(m.modelPickerItems)-1 {
				m.modelPickerIdx++
			} else {
				m.modelPickerIdx = 0
			}
			return true, m, m.spinner.Tick
		case "enter":
			// Advance to model name step
			m.modelPickerStep = 1
			ti := textinput.New()
			if len(m.modelPickerItems) > 0 {
				ti.Placeholder = m.modelPickerItems[m.modelPickerIdx].DefaultModel()
			}
			ti.Focus()
			ti.Width = 40
			m.modelPickerInput = ti
			return true, m, textinput.Blink
		case "esc":
			m.modelPickerActive = false
			return true, m, m.spinner.Tick
		}

	case 1: // Model name input step
		switch msg.String() {
		case "enter":
			m = m.applyModelSwitch()
			return true, m, m.spinner.Tick
		case "esc":
			// Go back to provider step
			m.modelPickerStep = 0
			return true, m, m.spinner.Tick
		default:
			// Pass key to the text input
			var cmd tea.Cmd
			m.modelPickerInput, cmd = m.modelPickerInput.Update(msg)
			return true, m, cmd
		}
	}

	return false, m, nil
}

// applyModelSwitch builds a new LLM client and swaps it into the agent.
func (m Model) applyModelSwitch() Model {
	if len(m.modelPickerItems) == 0 || m.modelPickerIdx >= len(m.modelPickerItems) {
		m.modelPickerActive = false
		return m
	}

	p := m.modelPickerItems[m.modelPickerIdx]
	modelName := m.modelPickerInput.Value()
	if modelName == "" {
		modelName = p.DefaultModel()
	}

	// Collect credentials from viper/env (reuse existing helper)
	values := collectProviderValues(p)

	client, err := p.BuildClient(values, modelName)
	if err != nil {
		// Show error in logs and close picker
		m.logs = append(m.logs, logEntry{
			Type:    "error",
			Content: fmt.Sprintf("Failed to switch model: %v", err),
		})
		m.modelPickerActive = false
		m.updateViewportContent()
		return m
	}

	// Swap the client
	m.agent.SwapLLMClient(client)
	m.modelName = client.GetModel()
	m.modelPickerActive = false

	m.logs = append(m.logs, logEntry{
		Type:    "response",
		Content: fmt.Sprintf("Switched to **%s** (%s)", m.modelName, p.DisplayName()),
	})
	m.updateViewportContent()
	return m
}

// renderModelPicker renders the model picker panel above the input.
func (m Model) renderModelPicker() string {
	if !m.modelPickerActive {
		return ""
	}

	var lines []string

	if m.modelPickerStep == 0 {
		// Provider list
		header := SlashItemStyle.Render("  Select provider (↑↓ navigate, enter select, esc cancel)")
		lines = append(lines, header)
		for i, p := range m.modelPickerItems {
			line := "  " + p.DisplayName()
			if i == m.modelPickerIdx {
				line = SlashItemSelectedStyle.Render(line)
			} else {
				line = SlashItemStyle.Render(line)
			}
			lines = append(lines, line)
		}
	} else {
		// Model name input
		selected := ""
		if len(m.modelPickerItems) > 0 && m.modelPickerIdx < len(m.modelPickerItems) {
			selected = m.modelPickerItems[m.modelPickerIdx].DisplayName()
		}
		lines = append(lines, SlashItemStyle.Render(fmt.Sprintf("  Provider: %s", SlashItemSelectedStyle.Render(selected))))
		lines = append(lines, SlashItemStyle.Render("  Model name (enter to confirm, esc to go back):"))
		lines = append(lines, "  "+m.modelPickerInput.View())
	}

	return SlashPanelStyle.Render(strings.Join(lines, "\n"))
}

// modelPickerHeight returns the rendered height of the model picker panel.
func (m Model) modelPickerHeight() int {
	if !m.modelPickerActive {
		return 0
	}
	if m.modelPickerStep == 0 {
		return len(m.modelPickerItems) + 2 // header + providers + padding
	}
	return 4 // provider line + prompt line + input line + padding
}
