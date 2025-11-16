package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/HrodWolfS/snipster/internal/model"
	"github.com/HrodWolfS/snipster/internal/snippets"
	"github.com/HrodWolfS/snipster/internal/version"
)

func ensureDataDir() (string, error) {
	// 1) Explicit override via env
	if dir := os.Getenv("SNIPSTER_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return dir, nil
	}
	// 2) Try user's HOME
	if home, err := os.UserHomeDir(); err == nil {
		root := filepath.Join(home, ".snipster", "snippets")
		if err := os.MkdirAll(root, 0o755); err == nil {
			return root, nil
		} else if !errors.Is(err, os.ErrPermission) {
			// return non-permission errors
			return "", err
		}
		// fallthrough to local
	}
	// 3) Fallback to local workspace directory
	root := filepath.Join(".snipster", "snippets")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	log.Printf("using local data dir: %s", root)
	return root, nil
}

func main() {
	// Lightweight handling of version flags before starting the TUI
	for _, a := range os.Args[1:] {
		if a == "--version" || a == "-version" || a == "-v" || a == "version" {
			prog := filepath.Base(os.Args[0])
			v := version.Version
			extra := ""
			if version.Commit != "" || version.Date != "" {
				extra = fmt.Sprintf(" (%s %s)", version.Commit, version.Date)
			}
			fmt.Printf("%s %s%s\n", prog, v, extra)
			return
		}
	}

	dataDir, err := ensureDataDir()
	if err != nil {
		log.Fatalf("failed to ensure data dir: %v", err)
	}

	repo := snippets.NewRepo(dataDir)
	all, err := repo.LoadAll()
	if err != nil {
		log.Printf("warning: failed to load snippets: %v", err)
	}

	m := model.New(appContext{repo: repo, dataDir: dataDir}, all)

	// Graceful shutdown on Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := tea.NewProgram(m, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

type appContext struct {
	repo    *snippets.Repo
	dataDir string
}

func (a appContext) Repo() *snippets.Repo { return a.repo }
func (a appContext) DataDir() string      { return a.dataDir }
