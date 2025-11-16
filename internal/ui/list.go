package ui

import (
    "github.com/charmbracelet/bubbles/list"
)

func NewList() list.Model {
    // Customize delegate to strengthen selection contrast in sidebar
    del := list.NewDefaultDelegate()
    del.Styles.SelectedTitle = del.Styles.SelectedTitle.
        Foreground(Theme.Accent2).Bold(true)
    del.Styles.SelectedDesc = del.Styles.SelectedDesc.
        Foreground(Theme.Accent2)

    l := list.New([]list.Item{}, del, 30, 10)
    l.SetShowStatusBar(false)
    l.SetShowHelp(false)
    l.SetShowFilter(false)
    l.SetShowPagination(false)
    l.DisableQuitKeybindings()
    return l
}
