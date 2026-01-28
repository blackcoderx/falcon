package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Minimal color palette
var (
	// USER: Adjust these colors to change the theme
	DimColor     = lipgloss.Color("#6c6c6c")
	TextColor    = lipgloss.Color("#e0e0e0")
	AccentColor  = lipgloss.Color("#7aa2f7") // The blue cursor/spinner color
	ErrorColor   = lipgloss.Color("#f7768e")
	ToolColor    = lipgloss.Color("#9ece6a")
	MutedColor   = lipgloss.Color("#545454")
	SuccessColor = lipgloss.Color("#73daca")
	WarningColor = lipgloss.Color("#e0af68") // Yellow/orange for warnings

	// OpenCode-style colors
	UserMessageBg = lipgloss.Color("#2a2a2a") // Gray background for user messages
	InputAreaBg   = lipgloss.Color("#2a2a2a") // Matches user messages
	FooterBg      = lipgloss.Color("#1a1a1a") // Darker footer
	ModelBadgeBg  = lipgloss.Color("#565f89") // Model name badge

	// Compact tool call colors
	ToolNameColor = lipgloss.Color("#cf8a6b") // Warm orange for tool names
	ToolArgsColor = lipgloss.Color("#6c6c6c") // Dim for arguments
	ToolUseColor  = lipgloss.Color("#545454") // Very muted for usage fraction

	// Response card
	ResponseCardBg    = lipgloss.Color("#1e1e2e") // Slightly elevated background
	ResponseCardBorder = lipgloss.Color("#3b3b5c") // Subtle border
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
				Foreground(AccentColor)

	StatusToolStyle = lipgloss.NewStyle().
			Foreground(ToolColor)

	// Status label style (for "thinking", "streaming", "tool calling")
	StatusLabelStyle = lipgloss.NewStyle().
				Foreground(DimColor)

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Shortcut key style
	ShortcutKeyStyle = lipgloss.NewStyle().
				Foreground(AccentColor)

	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(DimColor)

	// Footer specific styles (OpenCode style)
	FooterAppNameStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				PaddingRight(1)

	FooterModelStyle = lipgloss.NewStyle().
				Foreground(DimColor).
				PaddingRight(1)

	FooterInfoStyle = lipgloss.NewStyle().
			Foreground(DimColor)
)

// OpenCode-style message block styles
var (
	// User message: blue left border + gray background + vertical spacing
	UserMessageStyle = lipgloss.NewStyle().
				Background(UserMessageBg).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(AccentColor).
				BorderLeft(true).
				BorderTop(false).
				BorderRight(true).
				BorderBottom(false).
				Padding(1, 2).
				MarginLeft(ContentPadLeft).
				MarginTop(1).
				MarginBottom(1)

	// Compact tool call styles
	ToolNameCompactStyle = lipgloss.NewStyle().
				Foreground(ToolNameColor)

	ToolArgsCompactStyle = lipgloss.NewStyle().
				Foreground(ToolArgsColor)

	ToolUsageCompactStyle = lipgloss.NewStyle().
				Foreground(ToolUseColor)

	ToolDurationStyle = lipgloss.NewStyle().
				Foreground(MutedColor)

	// Tool calls: dimmed with circle prefix (legacy, kept for compatibility)
	ToolCallStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	// Agent messages: plain text with left margin + top spacing
	AgentMessageStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				MarginLeft(ContentPadLeft).
				MarginTop(1)

	// Response card: subtle box for tool output/responses
	ResponseCardStyle = lipgloss.NewStyle().
				Background(ResponseCardBg).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ResponseCardBorder).
				Padding(1, 2).
				MarginLeft(2)

	// Input area: matches user message style exactly (same borders, padding, margin)
	InputAreaStyle = lipgloss.NewStyle().
			Background(InputAreaBg).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(AccentColor).
			BorderLeft(true).
			BorderTop(false).
			BorderRight(true).
			BorderBottom(false).
			Padding(1, 2).
			MarginLeft(ContentPadLeft)

	// Footer bar style
	FooterStyle = lipgloss.NewStyle().
			Background(FooterBg).
			Foreground(DimColor).
			PaddingLeft(2)

	// Model badge
	ModelBadgeStyle = lipgloss.NewStyle().
			Background(ModelBadgeBg).
			Foreground(TextColor).
			Padding(0, 1)
)

// Content layout constants
const (
	ContentPadLeft  = 2 // Left padding for viewport content
	ContentPadRight = 2 // Right padding for viewport content
)

// Log prefixes
const (
	ThinkingPrefix    = "  thinking "
	ToolPrefix        = "  tool "
	ObservationPrefix = "  result "
	ErrorPrefix       = "  error "
	Separator         = "───"
	ToolCallPrefix    = "○ " // Circle prefix for tool calls (legacy)
)

// Pulse animation colors for status circle (dim blue → bright blue → dim blue)
var PulseColors = []lipgloss.Color{
	"#2a2f4e",
	"#3b4570",
	"#4c5a92",
	"#5d70b4",
	"#6e86d6",
	"#7aa2f7", // peak brightness (accent color)
	"#6e86d6",
	"#5d70b4",
	"#4c5a92",
	"#3b4570",
}

// Tool usage display styles
var (
	// Normal usage (green)
	ToolUsageNormalStyle = lipgloss.NewStyle().
				Foreground(ToolColor)

	// Warning usage (70-89% - yellow)
	ToolUsageWarningStyle = lipgloss.NewStyle().
				Foreground(WarningColor)

	// Critical usage (90%+ - red)
	ToolUsageCriticalStyle = lipgloss.NewStyle().
				Foreground(ErrorColor)

	// Tool name in usage display
	ToolUsageNameStyle = lipgloss.NewStyle().
				Foreground(DimColor)

	// Total usage style
	TotalUsageStyle = lipgloss.NewStyle().
			Foreground(AccentColor)
)

// Diff colors for file write confirmation
var (
	DiffAddColor    = lipgloss.Color("#73daca") // Green - added lines
	DiffRemoveColor = lipgloss.Color("#f7768e") // Red - removed lines
	DiffHunkColor   = lipgloss.Color("#7aa2f7") // Blue - hunk headers @@
	DiffHeaderColor = lipgloss.Color("#e0af68") // Yellow - file headers ---/+++
)

// Diff styles
var (
	DiffAddStyle = lipgloss.NewStyle().
			Foreground(DiffAddColor)

	DiffRemoveStyle = lipgloss.NewStyle().
			Foreground(DiffRemoveColor)

	DiffHunkStyle = lipgloss.NewStyle().
			Foreground(DiffHunkColor)

	DiffHeaderStyle = lipgloss.NewStyle().
			Foreground(DiffHeaderColor).
			Bold(true)

	DiffContextStyle = lipgloss.NewStyle().
				Foreground(DimColor)
)

// Confirmation dialog styles
var (
	ConfirmHeaderStyle = lipgloss.NewStyle().
				Foreground(WarningColor).
				Bold(true)

	ConfirmPathStyle = lipgloss.NewStyle().
				Foreground(AccentColor)

	ConfirmFooterStyle = lipgloss.NewStyle().
				Background(FooterBg).
				Padding(0, 1)

	ConfirmApproveStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true)

	ConfirmRejectStyle = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Bold(true)
)
