package model

import (
    "fmt"
    "sort"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/atotto/clipboard"
    fuzzy "github.com/sahilm/fuzzy"

    "github.com/HrodWolfS/snipster/internal/snippets"
    "github.com/HrodWolfS/snipster/internal/ui"
)

type AppState int

const (
    StateWelcome AppState = iota
    StateHome
    StateCreate
    StateEdit
    StateConfirmDelete
)

type AppContext interface{
    Repo() *snippets.Repo
    DataDir() string
}

// Sidebar item kinds: folder vs snippet (file)
type SidebarItemKind int

const (
    SidebarItemFolder SidebarItemKind = iota
    SidebarItemSnippet
)

// SidebarItem models folders and snippet rows for the sidebar list.
type SidebarItem struct {
    Kind    SidebarItemKind
    Name    string              // e.g. "frontend", "react", or snippet title
    Path    string              // e.g. "frontend" or "frontend/react"
    Indent  int                 // 0=top folder, 1=subfolder, 2=snippet
    Snippet *snippets.Snippet   // nil for folders
}

// folderNode represents a folder in the category tree for sidebar navigation.
type folderNode struct {
    Name     string
    Path     string
    Children map[string]*folderNode
    Snippets []*snippets.Snippet
}

func (i SidebarItem) Title() string {
    indent := strings.Repeat("  ", i.Indent)
    switch i.Kind {
    case SidebarItemFolder:
        // folder icon + name + slash
        return ui.Theme.SidebarTitle.Render(indent + "📁 " + i.Name + "/")
    case SidebarItemSnippet:
        if i.Snippet == nil {
            return indent + "📄 " + i.Name
        }
        return indent + "📄 " + i.Snippet.Title
    default:
        return indent + i.Name
    }
}

func (i SidebarItem) Description() string {
    switch i.Kind {
    case SidebarItemFolder:
        return "Dossier"
    case SidebarItemSnippet:
        if i.Snippet == nil { return "" }
        return fmt.Sprintf("%s  [%s]", i.Snippet.Category, strings.Join(i.Snippet.Tags, ", "))
    default:
        return ""
    }
}

func (i SidebarItem) FilterValue() string {
    if i.Kind == SidebarItemSnippet && i.Snippet != nil {
        return i.Snippet.Title
    }
    return i.Path
}

type Model struct {
    ctx AppContext

    Snippets  []snippets.Snippet
    // Hierarchical sidebar and visible items after filtering
    SidebarItems []SidebarItem
    VisibleItems []SidebarItem
    // Folder navigation state
    CurrentPath string
    folderRoot  *folderNode

    SearchInput textinput.Model
    SearchQuery string
    SearchActive bool
    Fuzzy       bool
    List        list.Model
    Preview     viewport.Model

    State         AppState
    SelectedIndex int

    // Transient status
    Status string

    // Modal fields
    mTitle    textinput.Model
    mCategory textinput.Model
    mTags     textinput.Model
    mLang     textinput.Model
    mContent  textarea.Model

    // Editing target
    editing *snippets.Snippet

    // Modal focus index: 0=title,1=category,2=tags,3=lang,4=content
    modalFocus int

    // Modal field errors
    mErrTitle    string
    mErrCategory string
    mErrContent  string

    // Window size for centering/layout
    Width  int
    Height int

    // Border accent index for theme toggle
    BorderIndex int
}

func New(ctx AppContext, initial []snippets.Snippet) Model {
    search := ui.NewInput("Search (/, title/tags/category/content)")
    l := ui.NewList()
    vp := viewport.New(60, 20)
    m := Model{
        ctx:         ctx,
        Snippets:    sortSnippets(initial),
        SearchInput: search,
        List:        l,
        Preview:     vp,
        State:       StateWelcome,
        BorderIndex: 0,
        CurrentPath: "",
        Fuzzy:       false,
        SearchActive: false,
    }
    m.rebuildSidebar()
    m.applyFilter("")
    m.initModalInputs()
    return m
}

func sortSnippets(in []snippets.Snippet) []snippets.Snippet {
    out := append([]snippets.Snippet(nil), in...)
    sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title) })
    return out
}

func (m *Model) initModalInputs() {
    m.mTitle = ui.NewInput("Title")
    m.mCategory = ui.NewInput("category e.g. backend/express")
    m.mTags = ui.NewInput("tags comma-separated")
    m.mLang = ui.NewInput("language e.g. js, ts, go")
    ta := textarea.New()
    ta.Placeholder = "content..."
    ta.SetWidth(60)
    ta.SetHeight(10)
    m.mContent = ta
    m.modalFocus = 0
    m.setModalFocus(0)
    m.mErrTitle, m.mErrCategory, m.mErrContent = "", "", ""
}

func (m Model) Init() tea.Cmd { return nil }

// Helpers
func (m *Model) currentSnippet() (snippets.Snippet, bool) {
    if len(m.VisibleItems) == 0 { return snippets.Snippet{}, false }
    idx := m.List.Index()
    if idx < 0 || idx >= len(m.VisibleItems) { return snippets.Snippet{}, false }
    it := m.VisibleItems[idx]
    if it.Kind != SidebarItemSnippet || it.Snippet == nil { return snippets.Snippet{}, false }
    return *it.Snippet, true
}

// applyFilter rebuilds the visible sidebar items based on the current query.
// - Empty query: show full hierarchical tree
// - Non-empty: flat list of matching snippets (no categories)
func (m *Model) applyFilter(q string) {
    m.SearchQuery = strings.TrimSpace(q)
    qq := strings.ToLower(m.SearchQuery)
    if qq == "" {
        // Show only current folder contents (folders + snippets).
        m.VisibleItems = m.itemsForFolder(m.CurrentPath)
    } else {
        var out []SidebarItem
        for i := range m.Snippets {
            s := m.Snippets[i]
            matches := false
            if m.Fuzzy {
                // Use fuzzy on title and category primarily
                if len(fuzzy.Find(qq, []string{s.Title})) > 0 || len(fuzzy.Find(qq, []string{s.Category})) > 0 {
                    matches = true
                }
            } else {
                if strings.Contains(strings.ToLower(s.Title), qq) ||
                    strings.Contains(strings.ToLower(s.Category), qq) ||
                    strings.Contains(strings.ToLower(s.Content), qq) ||
                    tagsContain(s.Tags, qq) {
                    matches = true
                }
            }
            if matches {
                ss := s // local copy for address stability
                // Highlight title for visibility in list
                displayTitle := ss.Title
                if m.Fuzzy {
                    // best-effort: highlight fuzzy matches on title
                    mm := fuzzy.Find(qq, []string{ss.Title})
                    if len(mm) > 0 {
                        displayTitle = highlightFuzzy(ss.Title, mm[0].MatchedIndexes)
                    }
                } else {
                    displayTitle = highlightContainsString(ss.Title, qq)
                }
                out = append(out, SidebarItem{
                    Kind: SidebarItemSnippet,
                    Name: displayTitle,
                    Path: ss.Category,
                    Indent: 0,
                    Snippet: &ss,
                })
            }
        }
        m.VisibleItems = out
    }

    // feed list items
    items := make([]list.Item, 0, len(m.VisibleItems))
    for i := range m.VisibleItems { items = append(items, m.VisibleItems[i]) }
    m.List.SetItems(items)
    if len(items) > 0 && (m.List.Index() < 0 || m.List.Index() >= len(items)) { m.List.Select(0) }
    m.refreshPreview()
}

// highlightContainsString wraps matches using the theme match style.
func highlightContainsString(s, q string) string {
    if q == "" { return s }
    lower := strings.ToLower(s)
    var b strings.Builder
    i := 0
    for {
        idx := strings.Index(lower[i:], q)
        if idx < 0 {
            b.WriteString(s[i:])
            break
        }
        idx += i
        b.WriteString(s[i:idx])
        b.WriteString(ui.Theme.Match.Render(s[idx:idx+len(q)]))
        i = idx + len(q)
        if i >= len(s) { break }
    }
    return b.String()
}

// highlightFuzzy highlights characters at the given indexes.
func highlightFuzzy(s string, idxs []int) string {
    if len(idxs) == 0 { return s }
    // Build a set of matched rune positions; note: assuming ASCII for simplicity
    mark := make(map[int]struct{}, len(idxs))
    for _, i := range idxs { mark[i] = struct{}{} }
    var b strings.Builder
    for i, r := range s {
        if _, ok := mark[i]; ok {
            b.WriteString(ui.Theme.Match.Render(string(r)))
        } else {
            b.WriteRune(r)
        }
    }
    return b.String()
}

func tagsContain(tags []string, q string) bool {
    for _, t := range tags { if strings.Contains(strings.ToLower(t), q) { return true } }
    return false
}

func (m *Model) refreshPreview() {
    s, ok := m.currentSnippet()
    if !ok {
        m.Preview.SetContent("No snippet")
        return
    }
    // Render with basic code styling and gutter
    m.Preview.SetContent(ui.RenderCodeHighlighted(s, m.SearchQuery))
}

// Clipboard helper
func copyToClipboard(content string) tea.Cmd {
    return func() tea.Msg {
        _ = clipboard.WriteAll(content)
        return statusMsg("copied to clipboard")
    }
}

// Messages
type statusMsg string

type reloadedMsg []snippets.Snippet

// CRUD messages
type createdMsg snippets.Snippet
type updatedMsg snippets.Snippet
type deletedMsg struct{ id string }

// Focus helpers for modal
func (m *Model) setModalFocus(idx int) {
    if idx < 0 { idx = 0 }
    if idx > 4 { idx = 4 }
    m.modalFocus = idx
    // Blur all
    m.mTitle.Blur(); m.mCategory.Blur(); m.mTags.Blur(); m.mLang.Blur(); m.mContent.Blur()
    switch idx {
    case 0: m.mTitle.Focus()
    case 1: m.mCategory.Focus()
    case 2: m.mTags.Focus()
    case 3: m.mLang.Focus()
    case 4: m.mContent.Focus()
    }
}

// Build hierarchical sidebar items from m.Snippets as a folder tree.
func (m *Model) rebuildSidebar() {
    root := &folderNode{Children: map[string]*folderNode{}}

    // Build tree from categories
    for i := range m.Snippets {
        s := &m.Snippets[i]
        cat := strings.TrimSpace(s.Category)
        if cat == "" { cat = "uncategorized" }
        parts := strings.Split(cat, "/")
        cur := root
        var pathParts []string
        for _, p := range parts {
            if p == "" { continue }
            pathParts = append(pathParts, p)
            if cur.Children == nil { cur.Children = map[string]*folderNode{} }
            if cur.Children[p] == nil {
                cur.Children[p] = &folderNode{Name: p, Path: strings.Join(pathParts, "/"), Children: map[string]*folderNode{}}
            }
            cur = cur.Children[p]
        }
        cur.Snippets = append(cur.Snippets, s)
    }

    // Flatten tree to sidebar items
    var items []SidebarItem
    var flattenFolder func(node *folderNode, indent int)
    flattenFolder = func(node *folderNode, indent int) {
        if node.Name != "" {
            items = append(items, SidebarItem{
                Kind:  SidebarItemFolder,
                Name:  node.Name,
                Path:  node.Path,
                Indent: indent,
            })
        }

        // sort children by name for deterministic order
        keys := make([]string, 0, len(node.Children))
        for k := range node.Children { keys = append(keys, k) }
        sort.Strings(keys)
        for _, k := range keys {
            flattenFolder(node.Children[k], indent+1)
        }

        // sort snippets by title
        sort.Slice(node.Snippets, func(i, j int) bool {
            return strings.ToLower(node.Snippets[i].Title) < strings.ToLower(node.Snippets[j].Title)
        })
        for _, sp := range node.Snippets {
            items = append(items, SidebarItem{
                Kind:    SidebarItemSnippet,
                Name:    sp.Title,
                Path:    sp.Category,
                Indent:  indent + 1,
                Snippet: sp,
            })
        }
    }

    // Flatten each top-level folder (children of root) in sorted order
    topKeys := make([]string, 0, len(root.Children))
    for k := range root.Children { topKeys = append(topKeys, k) }
    sort.Strings(topKeys)
    for _, k := range topKeys {
        flattenFolder(root.Children[k], 0)
    }

    m.SidebarItems = items
    m.folderRoot = root
}

// itemsForFolder returns immediate child folders and snippets of the folder at path.
func (m *Model) itemsForFolder(path string) []SidebarItem {
    node := m.findFolder(path)
    if node == nil { return nil }
    var out []SidebarItem
    // children folders sorted
    keys := make([]string, 0, len(node.Children))
    for k := range node.Children { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys {
        child := node.Children[k]
        out = append(out, SidebarItem{Kind: SidebarItemFolder, Name: child.Name, Path: child.Path, Indent: 0})
    }
    // snippets sorted by title
    sort.Slice(node.Snippets, func(i, j int) bool { return strings.ToLower(node.Snippets[i].Title) < strings.ToLower(node.Snippets[j].Title) })
    for _, sp := range node.Snippets {
        out = append(out, SidebarItem{Kind: SidebarItemSnippet, Name: sp.Title, Path: sp.Category, Indent: 0, Snippet: sp})
    }
    return out
}

func (m *Model) findFolder(path string) *folderNode {
    if m.folderRoot == nil { return nil }
    if path == "" { return m.folderRoot }
    cur := m.folderRoot
    parts := strings.Split(path, "/")
    for _, p := range parts {
        if p == "" { continue }
        next := cur.Children[p]
        if next == nil { return cur }
        cur = next
    }
    return cur
}

// Validate current modal inputs, set error messages and focus first invalid.
func (m *Model) validateModal() bool {
    m.mErrTitle, m.mErrCategory, m.mErrContent = "", "", ""
    title := strings.TrimSpace(m.mTitle.Value())
    cat := strings.TrimSpace(m.mCategory.Value())
    content := strings.TrimSpace(m.mContent.Value())
    valid := true
    if title == "" {
        m.mErrTitle = "Title is required"
        valid = false
    }
    if cat == "" {
        m.mErrCategory = "Category is required"
        valid = false
    }
    if content == "" {
        m.mErrContent = "Content is required"
        valid = false
    }
    if !valid {
        // Focus first invalid
        switch {
        case m.mErrTitle != "":
            m.setModalFocus(0)
        case m.mErrCategory != "":
            m.setModalFocus(1)
        case m.mErrContent != "":
            m.setModalFocus(4)
        }
    }
    return valid
}
