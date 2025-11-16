package snippets

import (
    "encoding/json"
    "errors"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type Repo struct {
    root string
}

func NewRepo(root string) *Repo { return &Repo{root: root} }

func (r *Repo) Root() string { return r.root }

// LoadAll scans recursively for .json snippet files.
func (r *Repo) LoadAll() ([]Snippet, error) {
    var out []Snippet
    err := filepath.WalkDir(r.root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
            return nil
        }
        b, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        var s Snippet
        if err := json.Unmarshal(b, &s); err != nil {
            return err
        }
        s.Path = path
        // Best-effort parse timestamps if zero strings were used
        if s.CreatedAt.IsZero() {
            s.CreatedAt = time.Now().UTC()
        }
        if s.UpdatedAt.IsZero() {
            s.UpdatedAt = s.CreatedAt
        }
        out = append(out, s)
        return nil
    })
    if errors.Is(err, os.ErrNotExist) {
        return []Snippet{}, nil
    }
    return out, err
}

