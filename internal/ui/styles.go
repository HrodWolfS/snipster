package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme centralizes all UI styles for consistent reuse across views.
type ThemeStyles struct {
	// Base colors
	Bg        lipgloss.Color
	Accent    lipgloss.Color
	Accent2   lipgloss.Color
	SidebarBg lipgloss.Color
	PreviewBg lipgloss.Color
	Muted     lipgloss.Color
	Error     lipgloss.Color

	// Containers
	Frame  lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	Sidebar      lipgloss.Style
	SidebarTitle lipgloss.Style
	Preview      lipgloss.Style
	PreviewTitle lipgloss.Style

	Title       lipgloss.Style
	Status      lipgloss.Style
	ErrorText   lipgloss.Style
	ModalBorder lipgloss.Style

	// Code rendering
	CodeGutter  lipgloss.Style
	CodeText    lipgloss.Style
	CodeKeyword lipgloss.Style

	// Highlight style for search matches
	Match lipgloss.Style
}

func NewTheme() ThemeStyles {
	bg := lipgloss.Color("#0f1117")
	side := lipgloss.Color("#131621")
	prev := lipgloss.Color("#0b0d14")
	accent := lipgloss.Color("#d16ba5")  // magenta
	accent2 := lipgloss.Color("#4cc9f0") // cyan
	muted := lipgloss.Color("#6b7280")
	errc := lipgloss.Color("#ef4444")

	// Transparent background: do not set Background for the frame.
	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent2).
		Padding(1, 2)

	// Transparent background for sidebar container.
	sidebar := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent2)

	// Transparent background for preview container.
	preview := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent2)

	return ThemeStyles{
		Bg: bg, Accent: accent, Accent2: accent2, SidebarBg: side, PreviewBg: prev, Muted: muted, Error: errc,
		Frame:        frame,
		Header:       lipgloss.NewStyle().Bold(true).Foreground(accent),
		Footer:       lipgloss.NewStyle().Foreground(muted),
		Sidebar:      sidebar,
		SidebarTitle: lipgloss.NewStyle().Foreground(accent2).Bold(true),
		Preview:      preview,
		PreviewTitle: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Title:        lipgloss.NewStyle().Bold(true),
		Status:       lipgloss.NewStyle().Foreground(accent2),
		ErrorText:    lipgloss.NewStyle().Foreground(errc),
		ModalBorder:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		CodeGutter:   lipgloss.NewStyle().Foreground(muted),
		CodeText:     lipgloss.NewStyle(),
		CodeKeyword:  lipgloss.NewStyle().Foreground(accent2).Bold(true),
		Match:        lipgloss.NewStyle().Foreground(accent2).Underline(true),
	}
}

var Theme = NewTheme()

// Backward-compatible aliases
var (
	AppStyle    = Theme.Frame
	TitleStyle  = Theme.Title
	StatusStyle = Theme.Status
	ErrorStyle  = Theme.ErrorText
	ModalBorder = Theme.ModalBorder
)

// Color palette for border accent toggling
var BorderColors = []lipgloss.Color{
	lipgloss.Color("#5BCEFA"), // cyan
	lipgloss.Color("#F5A9B8"), // pink
	lipgloss.Color("#B5E853"), // green
	lipgloss.Color("#FFCC66"), // orange
}

// SetBorderColor updates only border foreground colors across frame and containers.
func (t *ThemeStyles) SetBorderColor(c lipgloss.Color) {
	t.Frame = t.Frame.BorderForeground(c)
	t.Sidebar = t.Sidebar.BorderForeground(c)
	t.Preview = t.Preview.BorderForeground(c)
}
