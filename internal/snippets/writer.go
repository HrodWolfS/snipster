package snippets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Create writes a new snippet JSON file based on category and id.
func (r *Repo) Create(s Snippet) (Snippet, error) {
	if s.ID == "" {
		s.ID = Slugify(s.Title)
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	dir := filepath.Join(r.root, filepath.FromSlash(s.Category))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return s, err
	}
	path := filepath.Join(dir, s.ID+".json")
	if _, err := os.Stat(path); err == nil {
		return s, fmt.Errorf("snippet exists: %s", path)
	}
	if err := writeJSON(path, s); err != nil {
		return s, err
	}
	s.Path = path
	return s, nil
}

// Update overwrites an existing snippet JSON file.
func (r *Repo) Update(s Snippet) (Snippet, error) {
	if s.ID == "" {
		s.ID = Slugify(s.Title)
	}
	s.UpdatedAt = time.Now().UTC()
	// Respect existing path if provided; otherwise compute from category/id
	path := s.Path
	if path == "" {
		dir := filepath.Join(r.root, filepath.FromSlash(s.Category))
		path = filepath.Join(dir, s.ID+".json")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return s, err
	}
	if err := writeJSON(path, s); err != nil {
		return s, err
	}
	s.Path = path
	return s, nil
}

// Delete removes the snippet JSON file.
func (r *Repo) Delete(s Snippet) error {
	path := s.Path
	if path == "" {
		dir := filepath.Join(r.root, filepath.FromSlash(s.Category))
		path = filepath.Join(dir, s.ID+".json")
	}
	return os.Remove(path)
}

func writeJSON(path string, s Snippet) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(in string) string {
	x := strings.ToLower(strings.TrimSpace(in))
	x = slugRe.ReplaceAllString(x, "-")
	x = strings.Trim(x, "-")
	if x == "" {
		x = fmt.Sprintf("snippet-%d", time.Now().Unix())
	}
	return x
}
