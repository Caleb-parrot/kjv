package tui

import "github.com/charmbracelet/lipgloss"

// Palette taken from the DOS Bible reader on workspace 2:
// blue frames, cream body text, a darker title strip.
var (
	frame    = lipgloss.Color("#6b8cff")
	titleBar = lipgloss.Color("#3d5fbe")
	cream    = lipgloss.Color("#ffd75f")
	offWhite = lipgloss.Color("#fff1b8")
	selectBg = lipgloss.Color("#ffd75f")
	selectFg = lipgloss.Color("#1a1f33")
	muted    = lipgloss.Color("#7a88b8")
	numBlue  = lipgloss.Color("#8eb0ff")

	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(frame)

	activeFrameStyle = frameStyle.
				BorderForeground(lipgloss.Color("#a8c0ff"))

	titleStyle = lipgloss.NewStyle().
			Foreground(offWhite).
			Background(titleBar).
			Bold(true)

	indexItemStyle = lipgloss.NewStyle().
			Foreground(cream)

	indexMutedStyle = lipgloss.NewStyle().
			Foreground(muted)

	indexSelStyle = lipgloss.NewStyle().
			Foreground(selectFg).
			Background(selectBg).
			Bold(true)

	verseNumStyle = lipgloss.NewStyle().
			Foreground(numBlue).
			Bold(true)

	verseTextStyle = lipgloss.NewStyle().
			Foreground(cream)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted)

	statusStyle = lipgloss.NewStyle().
			Foreground(cream)

	searchStyle = lipgloss.NewStyle().
			Foreground(offWhite)

	badgeStyle = lipgloss.NewStyle().
			Foreground(offWhite).
			Background(lipgloss.Color("#c45c5c")).
			Bold(true).
			Padding(0, 1)
)
