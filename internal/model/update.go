package model

import (
    "os"
    "os/exec"
    "strings"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/HrodWolfS/snipster/internal/snippets"
    "github.com/HrodWolfS/snipster/internal/ui"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.State == StateWelcome {
            // Any key moves to home; `/` also enables search mode instantly
            if msg.String() == "/" {
                m.State = StateHome
                m.SearchActive = true
                m.SearchInput.Focus()
                return m, nil
            }
            m.State = StateHome
            return m, nil
        }
        if m.State == StateHome {
            // When in explicit search mode, only handle search keys and ESC.
            if m.SearchActive {
                switch msg.String() {
                case "esc":
                    m.SearchInput.SetValue("")
                    m.SearchQuery = ""
                    m.SearchInput.Blur()
                    m.SearchActive = false
                    m.applyFilter("")
                    return m, nil
                default:
                    // Update search input and apply filter; allow list to update for navigation
                    var sc tea.Cmd
                    m.SearchInput, sc = m.SearchInput.Update(msg)
                    m.applyFilter(m.SearchInput.Value())
                    var lc tea.Cmd
                    m.List, lc = m.List.Update(msg)
                    return m, tea.Batch(sc, lc)
                }
            }
            switch msg.String() {
            case "/":
                m.SearchActive = true
                m.SearchInput.Focus()
                return m, nil
            case "right", "l":
                // Enter folder when focused item is a folder and no active search
                if strings.TrimSpace(m.SearchQuery) == "" {
                    idx := m.List.Index()
                    if idx >= 0 && idx < len(m.VisibleItems) {
                        it := m.VisibleItems[idx]
                        if it.Kind == SidebarItemFolder {
                            m.CurrentPath = it.Path
                            m.applyFilter("")
                            if len(m.VisibleItems) > 0 { m.List.Select(0) }
                            return m, nil
                        }
                    }
                }
                return m, nil
            case "left", "h":
                // Go up to parent folder when no active search
                if strings.TrimSpace(m.SearchQuery) == "" {
                    if m.CurrentPath != "" {
                        // parent is CurrentPath without last segment
                        p := m.CurrentPath
                        if i := strings.LastIndex(p, "/"); i >= 0 {
                            m.CurrentPath = p[:i]
                        } else {
                            m.CurrentPath = ""
                        }
                        m.applyFilter("")
                        if len(m.VisibleItems) > 0 { m.List.Select(0) }
                        return m, nil
                    }
                }
                return m, nil
            case "esc":
                // If not in search mode, ignore ESC here (other dialogs handle their own ESC)
                if strings.TrimSpace(m.SearchQuery) != "" {
                    m.SearchInput.SetValue("")
                    m.SearchQuery = ""
                    m.SearchInput.Blur()
                    m.SearchActive = false
                    m.applyFilter("")
                    return m, nil
                }
                return m, nil
            case "t":
                // Cycle border accent color
                if n := len(ui.BorderColors); n > 0 {
                    m.BorderIndex = (m.BorderIndex + 1) % n
                    ui.Theme.SetBorderColor(ui.BorderColors[m.BorderIndex])
                }
                m.Status = "Border color changed"
                return m, nil
            case "f":
                // Toggle fuzzy search
                m.Fuzzy = !m.Fuzzy
                m.applyFilter(m.SearchInput.Value())
                if m.Fuzzy { m.Status = "Fuzzy search ON" } else { m.Status = "Fuzzy search OFF" }
                return m, nil
            case "enter":
                if s, ok := m.currentSnippet(); ok {
                    return m, copyToClipboard(s.Content)
                }
            case "E":
                if s, ok := m.currentSnippet(); ok {
                    path := s.Path
                    return m, func() tea.Msg {
                        ed := os.Getenv("VISUAL")
                        if ed == "" { ed = os.Getenv("EDITOR") }
                        if ed == "" { ed = "nano" }
                        cmd := exec.Command(ed, path)
                        cmd.Stdin = os.Stdin
                        cmd.Stdout = os.Stdout
                        cmd.Stderr = os.Stderr
                        _ = cmd.Run()
                        all, _ := m.ctx.Repo().LoadAll()
                        return reloadedMsg(all)
                    }
                }
                return m, nil
            case "n":
                m.State = StateCreate
                m.initModalInputs()
                m.mTitle.Focus()
                return m, nil
            case "e":
                if s, ok := m.currentSnippet(); ok {
                    m.State = StateEdit
                    m.initModalInputs()
                    m.editing = &s
                    m.mTitle.SetValue(s.Title)
                    m.mCategory.SetValue(s.Category)
                    m.mTags.SetValue(strings.Join(s.Tags, ", "))
                    m.mLang.SetValue(s.Language)
                    m.mContent.SetValue(s.Content)
                    m.mTitle.Focus()
                }
                return m, nil
            case "d":
                if s, ok := m.currentSnippet(); ok {
                    m.State = StateConfirmDelete
                    m.editing = &s
                }
                return m, nil
            case "q":
                return m, tea.Quit
            }
        } else {
            // In modal states
            switch m.State {
            case StateCreate, StateEdit:
                switch msg.String() {
                case "esc":
                    m.State = StateHome
                    m.editing = nil
                    return m, nil
                case "enter":
                    if m.modalFocus != 4 {
                        return m.handleSubmit()
                    }
                    // If focus is on content textarea, do not submit here;
                    // fall through to component update so textarea gets Enter.
                case "ctrl+s":
                    // Save from any field, including textarea
                    return m.handleSubmit()
                case "tab":
                    // Prevent advancing past required fields when empty
                    if isCurrentRequiredEmpty(&m) {
                        setCurrentRequiredError(&m)
                        return m, nil
                    }
                    m.setModalFocus(m.modalFocus+1)
                    return m, nil
                case "shift+tab":
                    m.setModalFocus(m.modalFocus-1)
                    return m, nil
                }
            case StateConfirmDelete:
                switch msg.String() {
                case "y", "Y":
                    if m.editing != nil {
                        s := *m.editing
                        return m, func() tea.Msg {
                            _ = m.ctx.Repo().Delete(s)
                            all, _ := m.ctx.Repo().LoadAll()
                            return reloadedMsg(all)
                        }
                    }
                case "n", "N", "esc":
                    m.State = StateHome
                    m.editing = nil
                    return m, nil
                }
            }
        }

    case tea.WindowSizeMsg:
        // Store terminal size
        w, h := msg.Width, msg.Height
        m.Width, m.Height = w, h

        // Compute inner content sizes consistent with viewLayout to prevent overflow/cut
        const border = 1
        const padX = 2
        const padY = 1
        contentWidth := m.Width - 2*(border+padX)
        contentHeight := m.Height - 2*(border+padY)
        if contentWidth < 20 { contentWidth = 20 }
        if contentHeight < 10 { contentHeight = 10 }

        // Reserve generous header height to avoid wrap-induced layout shifts
        headerHeight := 3
        footerHeight := 1
        bodyHeight := contentHeight - headerHeight - footerHeight
        if bodyHeight < 3 { bodyHeight = 3 }

        // Pane sizing (content area inside borders/padding)
        gap := 1
        const paneExtraX = 4 // 2 border + 2 padding
        const paneExtraY = 2 // 2 border
        availableContentW := contentWidth - gap - 2*paneExtraX
        if availableContentW < 10 { availableContentW = 10 }
        sbContentW := availableContentW / 3
        if sbContentW < 10 { sbContentW = 10 }
        pvContentW := availableContentW - sbContentW
        if pvContentW < 10 { pvContentW = 10 }
        paneContentH := bodyHeight - paneExtraY
        if paneContentH < 1 { paneContentH = 1 }

        // Apply sizes to bubbles components (content area sizes)
        m.List.SetSize(sbContentW, paneContentH)
        m.Preview.Width = pvContentW
        m.Preview.Height = paneContentH

        // Search input width in header
        siw := contentWidth / 3
        if siw < 20 { siw = 20 }
        m.SearchInput.Width = siw
        return m, nil

    case statusMsg:
        m.Status = string(msg)
        return m, nil

    case reloadedMsg:
        m.Snippets = sortSnippets([]snippets.Snippet(msg))
        m.rebuildSidebar()
        m.applyFilter(m.SearchInput.Value())
        m.State = StateHome
        m.editing = nil
        m.Status = "reloaded"
        return m, nil
    }

    // Update sub-components depending on state
    var cmd tea.Cmd
    switch m.State {
    case StateHome:
        var lc tea.Cmd
        m.List, lc = m.List.Update(msg)
        // Keep preview synced with list index
        m.refreshPreview()
        var sc tea.Cmd
        m.SearchInput, sc = m.SearchInput.Update(msg)
        // Always reflect current input into query and apply filter
        m.applyFilter(m.SearchInput.Value())
        cmd = tea.Batch(lc, sc)
    case StateCreate, StateEdit:
        // Update only focused input plus textarea if focused
        switch m.modalFocus {
        case 0:
            m.mTitle, _ = m.mTitle.Update(msg)
        case 1:
            m.mCategory, _ = m.mCategory.Update(msg)
        case 2:
            m.mTags, _ = m.mTags.Update(msg)
        case 3:
            m.mLang, _ = m.mLang.Update(msg)
        case 4:
            var taCmd tea.Cmd
            m.mContent, taCmd = m.mContent.Update(msg)
            cmd = tea.Batch(cmd, taCmd)
        }
    case StateConfirmDelete:
        // no sub-components
    }

    return m, cmd
}

func (m *Model) handleSubmit() (tea.Model, tea.Cmd) {
    // Build snippet from modal inputs
    tags := splitTags(m.mTags.Value())
    s := snippets.Snippet{
        Title:    strings.TrimSpace(m.mTitle.Value()),
        Category: strings.TrimSpace(m.mCategory.Value()),
        Language: strings.TrimSpace(m.mLang.Value()),
        Tags:     tags,
        Content:  m.mContent.Value(),
    }
    if m.State == StateEdit && m.editing != nil {
        s.ID = m.editing.ID
        s.Path = m.editing.Path
        s.CreatedAt = m.editing.CreatedAt
    }

    if !m.validateModal() {
        m.Status = "Please fix validation errors"
        return m, nil
    }

    return m, func() tea.Msg {
        var err error
        if m.State == StateCreate {
            s, err = m.ctx.Repo().Create(s)
        } else {
            s, err = m.ctx.Repo().Update(s)
        }
        if err != nil {
            return statusMsg("error: " + err.Error())
        }
        all, _ := m.ctx.Repo().LoadAll()
        return reloadedMsg(all)
    }
}

func splitTags(v string) []string {
    v = strings.TrimSpace(v)
    if v == "" { return nil }
    parts := strings.Split(v, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" { out = append(out, p) }
    }
    return out
}

// Helpers to enforce required fields on Tab navigation
func isCurrentRequiredEmpty(m *Model) bool {
    switch m.modalFocus {
    case 0: // title
        return strings.TrimSpace(m.mTitle.Value()) == ""
    case 1: // category
        return strings.TrimSpace(m.mCategory.Value()) == ""
    case 4: // content
        return strings.TrimSpace(m.mContent.Value()) == ""
    default:
        return false
    }
}

func setCurrentRequiredError(m *Model) {
    switch m.modalFocus {
    case 0:
        m.mErrTitle = "Title is required"
    case 1:
        m.mErrCategory = "Category is required"
    case 4:
        m.mErrContent = "Content is required"
    }
    m.Status = "Please fill required field"
}
