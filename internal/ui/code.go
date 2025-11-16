package ui

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/HrodWolfS/snipster/internal/snippets"
)

// RenderCode renders a snippet with a small header and a styled code block
// including a gutter with line numbers and basic keyword coloring.
func RenderCode(s snippets.Snippet) string {
    return RenderCodeHighlighted(s, "")
}

// RenderCodeHighlighted applies a simple substring highlight for lines containing the query.
// Highlighting is applied before keyword coloring for simplicity.
func RenderCodeHighlighted(s snippets.Snippet, query string) string {
    header := strings.Join([]string{
        Theme.PreviewTitle.Render(s.Title),
        Theme.Status.Render(fmt.Sprintf("%s | %s | %s", s.Category, s.Language, strings.Join(s.Tags, ", "))),
        "",
    }, "\n")

    q := strings.ToLower(strings.TrimSpace(query))
    lines := strings.Split(s.Content, "\n")
    var b strings.Builder
    for i, ln := range lines {
        // Left gutter with 1-based line numbers and a subtle bar; add an arrow if the line matches.
        marker := "│"
        if q != "" && strings.Contains(strings.ToLower(ln), q) {
            marker = Theme.CodeKeyword.Render("▶")
        }
        gutter := Theme.CodeGutter.Render(fmt.Sprintf("%3d %s ", i+1, marker))
        // Apply simple contains highlight first, then keyword coloring.
        if q != "" {
            ln = highlightContains(ln, q)
        }
        code := highlightLine(ln, s.Language)
        b.WriteString(gutter)
        b.WriteString(code)
        if i < len(lines)-1 { b.WriteString("\n") }
    }
    return header + b.String()
}

var (
    reJS     = regexp.MustCompile(`\b(const|let|var|function|return|if|else|for|while|switch|case|break|await|async|new|class|try|catch|throw)\b`)
    reGo     = regexp.MustCompile(`\b(func|package|import|return|if|else|for|range|switch|case|break|go|defer|type|struct|interface|map|chan|var|const)\b`)
    reSQL    = regexp.MustCompile(`\b(SELECT|FROM|WHERE|AND|OR|INSERT|INTO|VALUES|UPDATE|SET|DELETE|JOIN|LEFT|RIGHT|ON|GROUP|BY|ORDER|LIMIT)\b`)
    reTS     = reJS
)

func highlightLine(line, lang string) string {
    styleKeyword := Theme.CodeKeyword
    hi := func(re *regexp.Regexp, s string) string {
        return re.ReplaceAllStringFunc(s, func(m string) string { return styleKeyword.Render(m) })
    }
    switch strings.ToLower(lang) {
    case "js", "javascript":
        return hi(reJS, line)
    case "ts", "typescript":
        return hi(reTS, line)
    case "go", "golang":
        return hi(reGo, line)
    case "sql":
        return hi(reSQL, strings.ToUpper(line))
    default:
        return line
    }
}

// highlightContains wraps occurrences of q (already lowercased) in a subtle accent color.
func highlightContains(line, q string) string {
    if q == "" { return line }
    lower := strings.ToLower(line)
    var out strings.Builder
    i := 0
    for {
        idx := strings.Index(lower[i:], q)
        if idx < 0 {
            out.WriteString(line[i:])
            break
        }
        idx += i
        out.WriteString(line[i:idx])
        out.WriteString(Theme.CodeKeyword.Render(line[idx:idx+len(q)]))
        i = idx + len(q)
        if i >= len(line) { break }
    }
    return out.String()
}
