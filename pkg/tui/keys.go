package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg processes keyboard input and returns the updated model and command.
// This centralizes all key handling logic for the TUI.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Handle confirmation mode first (takes priority)
	if m.confirmationMode {
		return m.handleConfirmationKeys(msg)
	}

	// Handle slash command panel navigation (takes priority over regular keys)
	if m.slashState.Active {
		if handled, updatedModel, cmd := m.handleSlashKeys(msg); handled {
			return updatedModel, cmd
		}
	}

	// Handle model picker keys
	if m.modelPickerActive {
		if handled, updatedModel, cmd := m.handleModelPickerKeys(msg); handled {
			if cmd == nil {
				cmd = m.spinner.Tick
			}
			return updatedModel, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c":
		// Cancel any pending confirmation when quitting
		if m.confirmManager != nil {
			m.confirmManager.Cancel()
		}
		return m, tea.Quit

	case "esc":
		// If agent is running, cancel it instead of quitting
		if m.thinking && m.cancelAgent != nil {
			m.cancelAgent()
			m.thinking = false
			m.status = "idle"
			m.streamingBuffer = ""
			m.cancelAgent = nil
			// Remove any trailing streaming entry
			if len(m.logs) > 0 && m.logs[len(m.logs)-1].Type == "streaming" {
				m.logs = m.logs[:len(m.logs)-1]
			}
			m.logs = append(m.logs, logEntry{Type: "interrupted", Content: ""})
			m.updateViewportContent()
			return m, nil
		}
		// If not thinking, quit the application
		if m.confirmManager != nil {
			m.confirmManager.Cancel()
		}
		return m, tea.Quit

	case "ctrl+l":
		return m.handleClearScreen()

	case "ctrl+y":
		return m.handleCopyLastResponse()

	case "ctrl+u":
		return m.handleClearInput()

	case "shift+up":
		return m.handleHistoryUp()

	case "shift+down":
		return m.handleHistoryDown()

	case "enter":
		return m.handleEnter()

	case "up", "down", "pgup", "pgdown", "home", "end":
		return m.handleViewportScroll(msg)

	default:
		return m, nil
	}
}

// handleClearScreen clears all logs and resets the streaming buffer.
func (m Model) handleClearScreen() (Model, tea.Cmd) {
	m.logs = []logEntry{}
	m.streamingBuffer = ""
	m.updateViewportContent()
	return m, nil
}

// handleCopyLastResponse copies the last agent response to clipboard.
func (m Model) handleCopyLastResponse() (Model, tea.Cmd) {
	var lastResponse string
	for i := len(m.logs) - 1; i >= 0; i-- {
		if m.logs[i].Type == "response" {
			lastResponse = m.logs[i].Content
			break
		}
	}
	if lastResponse != "" {
		_ = clipboard.WriteAll(lastResponse)
	}
	return m, nil
}

// handleClearInput clears the current input and resets history navigation.
func (m Model) handleClearInput() (Model, tea.Cmd) {
	m.textinput.SetValue("")
	m.historyIdx = -1
	return m, nil
}

// handleHistoryUp navigates backwards through input history.
func (m Model) handleHistoryUp() (Model, tea.Cmd) {
	if m.thinking || len(m.inputHistory) == 0 {
		return m, nil
	}

	if m.historyIdx == -1 {
		// Save current input before navigating
		m.savedInput = m.textinput.Value()
		m.historyIdx = len(m.inputHistory) - 1
	} else if m.historyIdx > 0 {
		m.historyIdx--
	}

	m.textinput.SetValue(m.inputHistory[m.historyIdx])
	m.textinput.CursorEnd()
	return m, nil
}

// handleHistoryDown navigates forwards through input history.
func (m Model) handleHistoryDown() (Model, tea.Cmd) {
	if m.thinking || m.historyIdx == -1 {
		return m, nil
	}

	if m.historyIdx < len(m.inputHistory)-1 {
		m.historyIdx++
		m.textinput.SetValue(m.inputHistory[m.historyIdx])
	} else {
		// Return to saved input
		m.historyIdx = -1
		m.textinput.SetValue(m.savedInput)
	}

	m.textinput.CursorEnd()
	return m, nil
}

// handleEnter processes the enter key to send a message.
func (m Model) handleEnter() (Model, tea.Cmd) {
	if m.thinking {
		return m, nil
	}

	userInput := strings.TrimSpace(m.textinput.Value())
	if userInput == "" && m.slashState.FlowContent == "" {
		return m, nil
	}

	// Build the enriched input for the agent
	agentInput := userInput
	displayInput := userInput
	if m.slashState.FlowContent != "" {
		agentInput = fmt.Sprintf("File context:\n```yaml\n%s\n```\n\nTask: %s", m.slashState.FlowContent, userInput)
		m.slashState = SlashState{} // clear after use
	}

	if displayInput == "" {
		displayInput = "(context attached)"
	}

	// Add separator if there are previous logs
	if len(m.logs) > 0 {
		m.logs = append(m.logs, logEntry{Type: "separator", Content: ""})
	}
	m.logs = append(m.logs, logEntry{Type: "user", Content: displayInput})

	// Add to history (only real text, not the placeholder)
	if userInput != "" {
		m.inputHistory = append(m.inputHistory, userInput)
	}
	m.historyIdx = -1
	m.savedInput = ""

	m.textinput.SetValue("")
	m.thinking = true
	m.status = "thinking"
	m.streamingBuffer = ""
	m.updateViewportContent()

	return m, tea.Batch(
		m.spinner.Tick,
		runAgentAsync(m.agent, agentInput),
	)
}

// handleViewportScroll passes scroll events to the viewport.
func (m Model) handleViewportScroll(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleConfirmationKeys processes keyboard input during file write confirmation.
func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Approve the file change
		if m.confirmManager != nil {
			m.confirmManager.SendResponse(true)
		}
		m.confirmationMode = false
		m.logs = append(m.logs, logEntry{Type: "user", Content: "Approved file change"})
		m.pendingConfirmation = nil
		m.updateViewportContent()
		return m, nil

	case "n", "N":
		// Reject the file change
		if m.confirmManager != nil {
			m.confirmManager.SendResponse(false)
		}
		m.confirmationMode = false
		m.logs = append(m.logs, logEntry{Type: "error", Content: "Rejected file change"})
		m.pendingConfirmation = nil
		m.updateViewportContent()
		return m, nil

	case "esc", "ctrl+c":
		// Cancel = reject, but also quit on ctrl+c
		if m.confirmManager != nil {
			m.confirmManager.SendResponse(false)
		}
		m.confirmationMode = false
		m.pendingConfirmation = nil
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.logs = append(m.logs, logEntry{Type: "error", Content: "Rejected file change"})
		m.updateViewportContent()
		return m, nil

	case "pgup", "pgdown", "home", "end":
		// Allow scrolling in confirmation mode to view the diff
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	// Ignore other keys in confirmation mode
	return m, nil
}
