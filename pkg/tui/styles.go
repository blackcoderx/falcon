package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Minimal color palette
var (
	DimColor     = lipgloss.Color("#6c6c6c")
	TextColor    = lipgloss.Color("#e0e0e0")
	AccentColor  = lipgloss.Color("#7aa2f7")
	ErrorColor   = lipgloss.Color("#f7768e")
	ToolColor    = lipgloss.Color("#9ece6a")
	MutedColor   = lipgloss.Color("#545454")
	SuccessColor = lipgloss.Color("#73daca")

	// OpenCode-style colors
	UserMessageBg = lipgloss.Color("#2a2a2a") // Gray background for user messages
	InputAreaBg   = lipgloss.Color("#2a2a2a") // Matches user messages
	FooterBg      = lipgloss.Color("#1a1a1a") // Darker footer
	ModelBadgeBg  = lipgloss.Color("#565f89") // Model name badge
)

// Log entry styles
var (
	UserStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	ThinkingStyle = lipgloss.NewStyle().
			Foreground(DimColor).
			Italic(true)

	ToolStyle = lipgloss.NewStyle().
			Foreground(ToolColor)

	ObservationStyle = lipgloss.NewStyle().
				Foreground(DimColor)

	ResponseStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	PromptStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	HelpStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	// Status line styles
	StatusIdleStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	StatusActiveStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true)

	StatusToolStyle = lipgloss.NewStyle().
			Foreground(ToolColor).
			Bold(true)

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Shortcut key style
	ShortcutKeyStyle = lipgloss.NewStyle().
				Foreground(AccentColor)

	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(DimColor)
)

// OpenCode-style message block styles
var (
	// User message: blue left border + gray background
	UserMessageStyle = lipgloss.NewStyle().
				Background(UserMessageBg).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(AccentColor).
				BorderLeft(true).
				BorderTop(false).
				BorderRight(false).
				BorderBottom(false).
				PaddingLeft(1).
				PaddingRight(1).
				MarginTop(1).
				MarginBottom(1)

	// Tool calls: dimmed with circle prefix
	ToolCallStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	// Agent messages: plain text
	AgentMessageStyle = lipgloss.NewStyle().
				Foreground(TextColor)

	// Input area: matches user message style
	InputAreaStyle = lipgloss.NewStyle().
			Background(InputAreaBg).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(AccentColor).
			BorderLeft(true).
			BorderTop(false).
			BorderRight(false).
			BorderBottom(false).
			PaddingLeft(1)

	// Footer bar
	FooterStyle = lipgloss.NewStyle().
			Background(FooterBg).
			Foreground(DimColor).
			Padding(0, 1)

	// Model badge
	ModelBadgeStyle = lipgloss.NewStyle().
			Background(ModelBadgeBg).
			Foreground(TextColor).
			Padding(0, 1)
)

// Log prefixes (Claude Code style - kept for compatibility)
const (
	UserPrefix        = "> "
	ThinkingPrefix    = "  thinking "
	ToolPrefix        = "  tool "
	ObservationPrefix = "  result "
	ErrorPrefix       = "  error "
	Separator         = "───"

	// OpenCode-style prefix
	ToolCallPrefix = "○ " // Circle prefix for tool calls
)
