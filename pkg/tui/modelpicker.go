package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// openModelPicker initializes and opens the model picker.
// Called when the user selects /model.
// Only shows providers that have been configured in GlobalConfig.
func (m Model) openModelPicker() Model {
	gcfg, err := core.LoadGlobalConfig()
	if err != nil || gcfg == nil || len(gcfg.Providers) == 0 {
		m.logs = append(m.logs, logEntry{
			Type:    "system",
			Content: "No providers configured. Run 'falcon config' to add one.",
		})
		m.updateViewportContent()
		return m
	}

	// Build entries in provider registration order for a stable display.
	var entries []modelEntry
	for _, p := range llm.All() {
		entry, ok := gcfg.Providers[p.ID()]
		if !ok {
			continue
		}
		model := entry.Model
		if model == "" {
			model = p.DefaultModel()
		}
		entries = append(entries, modelEntry{
			ProviderID:  p.ID(),
			DisplayName: fmt.Sprintf("%s - %s", p.DisplayName(), model),
			Model:       model,
			Config:      entry.Config,
		})
	}

	if len(entries) == 0 {
		m.logs = append(m.logs, logEntry{
			Type:    "system",
			Content: "No providers configured. Run 'falcon config' to add one.",
		})
		m.updateViewportContent()
		return m
	}

	m.modelPickerActive = true
	m.modelPickerItems = entries
	m.modelPickerIdx = 0

	// Pre-select the currently active provider.
	currentProvider := gcfg.DefaultProvider
	for i, e := range entries {
		if e.ProviderID == currentProvider {
			m.modelPickerIdx = i
			break
		}
	}

	return m
}

// handleModelPickerKeys processes keyboard input for the model picker.
// Single-step: navigate the list with up/down, confirm with enter, cancel with esc.
func (m Model) handleModelPickerKeys(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	if !m.modelPickerActive {
		return false, m, nil
	}

	switch msg.String() {
	case "up", "shift+tab":
		if m.modelPickerIdx > 0 {
			m.modelPickerIdx--
		} else {
			m.modelPickerIdx = len(m.modelPickerItems) - 1
		}
		return true, m, nil
	case "down", "tab":
		if m.modelPickerIdx < len(m.modelPickerItems)-1 {
			m.modelPickerIdx++
		} else {
			m.modelPickerIdx = 0
		}
		return true, m, nil
	case "enter":
		m = m.applyModelSwitch()
		return true, m, m.spinner.Tick
	case "esc":
		m.modelPickerActive = false
		return true, m, nil
	}

	return false, m, nil
}

// applyModelSwitch builds a new LLM client from the selected modelEntry and swaps it into the agent.
func (m Model) applyModelSwitch() Model {
	if len(m.modelPickerItems) == 0 || m.modelPickerIdx >= len(m.modelPickerItems) {
		m.modelPickerActive = false
		return m
	}

	entry := m.modelPickerItems[m.modelPickerIdx]
	p, ok := llm.Get(entry.ProviderID)
	if !ok {
		m.logs = append(m.logs, logEntry{
			Type:    "error",
			Content: fmt.Sprintf("Unknown provider: %s", entry.ProviderID),
		})
		m.modelPickerActive = false
		m.updateViewportContent()
		return m
	}

	// Use config from the entry, with env-variable fallbacks for empty fields.
	values := make(map[string]string)
	for k, v := range entry.Config {
		values[k] = v
	}
	for _, f := range p.SetupFields() {
		if values[f.Key] == "" && f.EnvFallback != "" {
			values[f.Key] = os.Getenv(f.EnvFallback)
		}
	}

	client, err := p.BuildClient(values, entry.Model)
	if err != nil {
		m.logs = append(m.logs, logEntry{
			Type:    "error",
			Content: fmt.Sprintf("Failed to switch model: %v", err),
		})
		m.modelPickerActive = false
		m.updateViewportContent()
		return m
	}

	m.agent.SwapLLMClient(client)
	m.modelName = client.GetModel()
	m.modelPickerActive = false

	m.logs = append(m.logs, logEntry{
		Type:    "system",
		Content: fmt.Sprintf("Switched to %s (%s)", m.modelName, p.DisplayName()),
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
	header := SlashItemStyle.Render("  Switch model (↑↓ navigate, enter select, esc cancel)")
	lines = append(lines, header)

	for i, entry := range m.modelPickerItems {
		line := "  " + entry.DisplayName
		if i == m.modelPickerIdx {
			line = SlashItemSelectedStyle.Render(line)
		} else {
			line = SlashItemStyle.Render(line)
		}
		lines = append(lines, line)
	}

	return SlashPanelStyle.Render(strings.Join(lines, "\n"))
}

// modelPickerHeight returns the rendered height of the model picker panel.
func (m Model) modelPickerHeight() int {
	if !m.modelPickerActive {
		return 0
	}
	return len(m.modelPickerItems) + 2 // header + items + padding
}
