package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/HrodWolfS/snipster/internal/ui"
)

// ASCII title for the welcome screen. Keep spacing exactly as-is.
const asciiTitle = `
███████╗███╗   ██╗██╗██████╗ ███████╗████████╗███████╗██████╗ 
██╔════╝████╗  ██║██║██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗
███████╗██╔██╗ ██║██║██████╔╝███████╗   ██║   █████╗  ██████╔╝
╚════██║██║╚██╗██║██║██╔═══╝ ╚════██║   ██║   ██╔══╝  ██╔══██╗
███████║██║ ╚████║██║██║     ███████║   ██║   ███████╗██║  ██║
╚══════╝╚═╝  ╚═══╝╚═╝╚═╝     ╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝

            C L I   S N I P P E T   M A N A G E R
`

func (m Model) View() string {
	switch m.State {
	case StateWelcome:
		return m.viewWelcome()
	case StateCreate:
		return m.viewLayout() + "\n" + m.viewCreateEdit("Create")
	case StateEdit:
		return m.viewLayout() + "\n" + m.viewCreateEdit("Edit")
	case StateConfirmDelete:
		return m.viewLayout() + "\n" + m.viewConfirmDelete()
	default:
		return m.viewLayout()
	}
}

// Welcome screen with centered ASCII art and call-to-action
func (m Model) viewWelcome() string {
	// Style the ASCII art in orange and supporting text using theme accents.
	titleStyle := ui.Theme.Title.Foreground(lipgloss.Color("#FFA657"))
	asciiTitleStyled := titleStyle.Render(asciiTitle)
	subtitleStyled := ui.Theme.Status.Render("SNIPSTER")
	ctaStyled := ui.Theme.Footer.Render("Press any key to open your library")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		asciiTitleStyled,
		subtitleStyled,
		ctaStyled,
	)

	rendered := ui.Theme.Frame.Render(content)
	// Center the framed content within the available window size.
	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		rendered,
	)
}

func (m Model) viewLayout() string {
	// Compute inner content area to prevent overflow beyond the frame.
	// Frame has Rounded border (1 each side) and Padding(1,2).
	const border = 1
	const padX = 2
	const padY = 1
	contentWidth := m.Width - 2*(border+padX)
	contentHeight := m.Height - 2*(border+padY)
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Header/footer heights
	headerHeight := 3
	footerHeight := 1
	bodyHeight := contentHeight - headerHeight - footerHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	// Body horizontal split: sidebar | gap | preview
	gap := 1
	// Each pane has Border(rounded) and Padding(0,1) -> extra width 2 (border) + 2 (padding) = 4
	const paneExtraX = 4
	// Vertical extra height: 2 for borders (no vertical padding)
	const paneExtraY = 2
	availableContentW := contentWidth - gap - 2*paneExtraX
	if availableContentW < 10 {
		availableContentW = 10
	}
	sbContentW := availableContentW / 3
	if sbContentW < 10 {
		sbContentW = 10
	}
	pvContentW := availableContentW - sbContentW
	if pvContentW < 10 {
		pvContentW = 10
	}

	paneContentH := bodyHeight - paneExtraY
	if paneContentH < 1 {
		paneContentH = 1
	}

	// Header: app title, search, status (no explicit sizing, relies on contentWidth)
	// Breadcrumbs path
	bc := m.CurrentPath
	if bc == "" {
		bc = "/"
	}
	// Show search field when search mode is active, or when focused, or when query is non-empty
	var searchView string
	if m.SearchActive || m.SearchInput.Focused() || strings.TrimSpace(m.SearchQuery) != "" {
		searchView = m.SearchInput.View()
	} else {
		// subtle hint when not searching
		searchView = ui.Theme.Footer.Render("/ to search")
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.Theme.Header.Render("Snipster"),
		"  ", ui.Theme.Status.Render(bc),
		"  ", searchView,
	)
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		left,
		"  ", ui.Theme.Status.Render(m.headerStatus()),
	)

	// Body: sized sidebar and preview contained within the frame.
	// Preview must have a single continuous border: put header+code (viewport content)
	// inside the same bordered container.
	sidebarContent := m.List.View()
	previewInner := m.Preview.View() // already includes header+code text (no borders)
	sidebarView := ui.Theme.Sidebar.
		Width(sbContentW).Height(paneContentH).
		Render(sidebarContent)
	previewView := ui.Theme.Preview.
		Width(pvContentW).Height(paneContentH).
		Render(previewInner)
	gapStr := strings.Repeat(" ", gap)
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, gapStr, previewView)

	// Footer: key help
	help := ui.Theme.Footer.Render("/ focus  j/k,↑/↓ navigate  enter copy  n new  e edit  d delete  q quit")

	inner := lipgloss.JoinVertical(lipgloss.Left, head, body, help)
	return ui.Theme.Frame.Render(inner)
}

func (m Model) viewCreateEdit(action string) string {
	// Inline error lines under fields if present
	titleLine := "Title*: " + m.mTitle.View()
	if m.mErrTitle != "" {
		titleLine += "\n" + ui.ErrorStyle.Render(m.mErrTitle)
	}

	catLine := "Category*: " + m.mCategory.View()
	if m.mErrCategory != "" {
		catLine += "\n" + ui.ErrorStyle.Render(m.mErrCategory)
	}

	contentHeader := "Content*:"
	contentBlock := m.mContent.View()
	if m.mErrContent != "" {
		contentBlock += "\n" + ui.ErrorStyle.Render(m.mErrContent)
	}

	form := strings.Join([]string{
		ui.TitleStyle.Render(fmt.Sprintf("%s Snippet", action)),
		"",
		titleLine,
		catLine,
		"Tags: " + m.mTags.View(),
		"Language: " + m.mLang.View(),
		contentHeader,
		contentBlock,
		ui.StatusStyle.Render("ctrl+s: save, esc: cancel (enter in content adds newline)"),
	}, "\n")
	return ui.ModalBorder.Render(form)
}

func (m Model) viewConfirmDelete() string {
	name := ""
	if m.editing != nil {
		name = m.editing.Title
	}
	msg := ui.TitleStyle.Render("Delete Snippet?") +
		"\n\n" + name +
		"\n\n" + ui.StatusStyle.Render("y: yes, n/esc: cancel")
	return ui.ModalBorder.Render(msg)
}

func (m Model) headerStatus() string {
	// Count only snippet rows in the current visible items
	count := 0
	for _, it := range m.VisibleItems {
		if it.Kind == SidebarItemSnippet && it.Snippet != nil {
			count++
		}
	}
	parts := []string{fmt.Sprintf("%d snippets", count)}
	if m.Fuzzy {
		parts = append(parts, "[fuzzy]")
	}
	if q := strings.TrimSpace(m.SearchInput.Value()); q != "" {
		parts = append(parts, "filter: "+q)
	}
	if m.Status != "" {
		parts = append(parts, m.Status)
	}
	return strings.Join(parts, "  ·  ")
}
