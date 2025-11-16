package ui

import (
    "github.com/charmbracelet/bubbles/textinput"
)

func NewInput(placeholder string) textinput.Model {
    ti := textinput.New()
    ti.Placeholder = placeholder
    ti.CharLimit = 256
    ti.Prompt = "Search: "
    return ti
}
