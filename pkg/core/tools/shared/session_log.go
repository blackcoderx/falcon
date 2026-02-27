package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionLogTool writes and reads session audit records in .falcon/sessions/.
// Each session records which tools were used, when, and a user-provided summary.
type SessionLogTool struct {
	falconDir   string
	sessionFile string // set on "start"
	startTime   time.Time
}

func NewSessionLogTool(falconDir string) *SessionLogTool {
	return &SessionLogTool{falconDir: falconDir}
}

type SessionLogParams struct {
	// Action: "start", "end", "list", "read"
	Action  string `json:"action"`
	Summary string `json:"summary,omitempty"` // for "end": what was accomplished
	Session string `json:"session,omitempty"` // for "read": session ID (timestamp prefix)
}

type sessionRecord struct {
	SessionID string    `json:"session_id"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time,omitempty"`
	Summary   string    `json:"summary,omitempty"`
}

func (t *SessionLogTool) Name() string { return "session_log" }

func (t *SessionLogTool) Description() string {
	return "Manage session audit records in .falcon/sessions/. Actions: start (record session start), end (record completion summary), list (show recent sessions), read (inspect a specific session)"
}

func (t *SessionLogTool) Parameters() string {
	return `{
  "action":  "start|end|list|read",
  "summary": "What was tested and what was found (for end)",
  "session": "session_20260227_153000 (for read)"
}`
}

func (t *SessionLogTool) Execute(args string) (string, error) {
	var params SessionLogParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	sessionsDir := filepath.Join(t.falconDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sessions directory: %w", err)
	}

	switch params.Action {
	case "start":
		t.startTime = time.Now()
		sessionID := "session_" + t.startTime.Format("20060102_150405")
		t.sessionFile = filepath.Join(sessionsDir, sessionID+".json")

		rec := sessionRecord{
			SessionID: sessionID,
			StartTime: t.startTime.Format(time.RFC3339),
		}
		data, _ := json.MarshalIndent(rec, "", "  ")
		if err := os.WriteFile(t.sessionFile, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write session record: %w", err)
		}
		return fmt.Sprintf("Session started: %s", sessionID), nil

	case "end":
		if t.sessionFile == "" {
			// Try to find the most recent session file
			files, err := os.ReadDir(sessionsDir)
			if err != nil || len(files) == 0 {
				return "No active session to end.", nil
			}
			// Sort descending to get the latest
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() > files[j].Name()
			})
			t.sessionFile = filepath.Join(sessionsDir, files[0].Name())
		}

		data, err := os.ReadFile(t.sessionFile)
		if err != nil {
			return "", fmt.Errorf("failed to read session file: %w", err)
		}

		var rec sessionRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return "", fmt.Errorf("failed to parse session record: %w", err)
		}

		rec.EndTime = time.Now().Format(time.RFC3339)
		rec.Summary = params.Summary

		updated, _ := json.MarshalIndent(rec, "", "  ")
		if err := os.WriteFile(t.sessionFile, updated, 0644); err != nil {
			return "", fmt.Errorf("failed to update session record: %w", err)
		}

		t.sessionFile = ""
		return fmt.Sprintf("Session ended. Summary recorded: %s", params.Summary), nil

	case "list":
		files, err := os.ReadDir(sessionsDir)
		if err != nil {
			return "", fmt.Errorf("failed to read sessions directory: %w", err)
		}

		var sessionFiles []string
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				sessionFiles = append(sessionFiles, f.Name())
			}
		}

		if len(sessionFiles) == 0 {
			return "No sessions recorded yet.", nil
		}

		// Show latest 10
		sort.Sort(sort.Reverse(sort.StringSlice(sessionFiles)))
		if len(sessionFiles) > 10 {
			sessionFiles = sessionFiles[:10]
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Recent sessions (%d shown):\n\n", len(sessionFiles)))
		for _, name := range sessionFiles {
			data, _ := os.ReadFile(filepath.Join(sessionsDir, name))
			var rec sessionRecord
			if err := json.Unmarshal(data, &rec); err != nil {
				sb.WriteString("  - " + name + " (unreadable)\n")
				continue
			}
			status := "in-progress"
			if rec.EndTime != "" {
				status = "completed"
			}
			summary := rec.Summary
			if summary == "" {
				summary = "(no summary)"
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s — %s\n    %s\n", status, rec.SessionID, rec.StartTime, summary))
		}
		return sb.String(), nil

	case "read":
		if params.Session == "" {
			return "", fmt.Errorf("session ID is required for read")
		}
		name := params.Session
		if !strings.HasSuffix(name, ".json") {
			name += ".json"
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, name))
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("session not found: %s", params.Session)
			}
			return "", fmt.Errorf("failed to read session: %w", err)
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("unknown action '%s' (use: start, end, list, read)", params.Action)
	}
}
