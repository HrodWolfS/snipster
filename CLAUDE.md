# Snipster - Guide Claude Code

Guide de développement pour Claude Code sur le projet Snipster.

## Vue d'ensemble

**Snipster** est un gestionnaire de snippets TUI (Terminal User Interface) construit avec Go et Bubble Tea. Il permet d'organiser, rechercher et gérer des snippets de code stockés localement en JSON.

**Binaire**: `snip` (nom court pour une utilisation rapide en CLI)
**Module Go**: `github.com/HrodWolfS/snipster`

## Architecture

### Structure des fichiers

```
snipster/
├── cmd/snip/                    # Point d'entrée unique
│   └── main.go                  # Bootstrap, gestion du dataDir
├── internal/
│   ├── model/                   # Architecture Bubble Tea (Elm)
│   │   ├── model.go            # État de l'application
│   │   ├── update.go           # Logique événements et navigation
│   │   └── view.go             # Rendu des différentes vues
│   ├── ui/                      # Composants UI réutilisables
│   │   ├── styles.go           # Thèmes et styles centralisés
│   │   ├── list.go             # Configuration liste sidebar
│   │   ├── input.go            # TextInput customisé
│   │   └── code.go             # Rendu code avec coloration
│   ├── snippets/                # Modèle de données
│   │   ├── snippet.go          # Structure Snippet
│   │   ├── loader.go           # Chargement depuis JSON
│   │   └── writer.go           # Écriture/sauvegarde
│   └── version/                 # Versioning (build info)
│       └── version.go
├── .github/workflows/
│   ├── ci.yml                   # Tests automatiques
│   └── release.yml              # Release avec GoReleaser v6
├── .goreleaser.yml              # Config GoReleaser v2
├── Makefile                     # Build avec version/commit/date
├── go.mod                       # Module: github.com/HrodWolfS/snipster
└── README.md
```

### Principes d'architecture

1. **Modularité**: Un fichier = une responsabilité (~50-200 lignes max)
2. **Bubble Tea (Elm)**: Model → Update → View pattern strict
3. **États explicites**: AppState enum pour gérer les modes (Welcome, Home, Create, Edit, ConfirmDelete)
4. **Contexte léger**: AppContext interface pour injection de dépendances

## Patterns de code

### 1. Gestion des états

```go
type AppState int

const (
    StateWelcome AppState = iota
    StateHome
    StateCreate
    StateEdit
    StateConfirmDelete
)
```

**Règle**: Toujours utiliser un switch sur `m.State` dans `Update()` et `View()`

### 2. Navigation hiérarchique

- **CurrentPath**: Chemin actuel dans l'arborescence (ex: "backend/express")
- **folderNode**: Arbre de dossiers avec enfants et snippets
- **SidebarItem**: Union type folder/snippet pour affichage liste

### 3. Recherche

- **SearchActive**: bool pour mode recherche actif
- **SearchQuery**: string du query courant
- **Fuzzy**: bool pour toggle recherche fuzzy
- **applyFilter()**: Reconstruit VisibleItems selon CurrentPath + query

### 4. Modals

- **États**: StateCreate, StateEdit, StateConfirmDelete
- **Champs**: mTitle, mCategory, mTags, mLang, mContent (tous textinput/textarea)
- **Focus**: modalFocus int (0-4) pour navigation Tab
- **Validation**: mErrTitle, mErrCategory, mErrContent pour feedback

### 5. Thèmes

```go
// ui/styles.go
var Theme = NewTheme()
var BorderColors = []lipgloss.Color{...}

// Changer couleur bordure avec touche 't'
Theme.SetBorderColor(BorderColors[m.BorderIndex % len(BorderColors)])
```

## Conventions de code

### Nommage

- **Fichiers**: snake_case (model.go, update.go, view.go)
- **Types**: PascalCase (Model, SidebarItem, AppState)
- **Fonctions privées**: camelCase (applyFilter, currentSnippet)
- **Fonctions publiques**: PascalCase (New, Init, Update, View)
- **Constantes**: UPPER_SNAKE_CASE pour enums (StateWelcome, SidebarItemFolder)

### Organisation des imports

```go
import (
    // Standard library
    "fmt"
    "strings"

    // Dépendances externes
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    // Modules internes
    "github.com/HrodWolfS/snipster/internal/snippets"
    "github.com/HrodWolfS/snipster/internal/ui"
)
```

### Gestion des erreurs

```go
// Logging pour warnings non-bloquants
log.Printf("warning: failed to load snippets: %v", err)

// Fatal pour erreurs critiques au démarrage
log.Fatalf("failed to ensure data dir: %v", err)

// Retour d'erreur pour opérations métier
if err := repo.Save(snippet); err != nil {
    m.Status = fmt.Sprintf("Error: %v", err)
    return m, nil
}
```

## Raccourcis clavier (état actuel)

### Navigation
- `↑↓jk`: Naviguer liste
- `→l`: Entrer dans dossier
- `←h`: Remonter au parent
- `/`: Activer recherche
- `f`: Toggle fuzzy
- `Esc`: Quitter recherche/modal

### Actions
- `Enter`: Copier snippet au presse-papiers
- `n`: Nouveau snippet (modal)
- `e`: Éditer snippet (modal)
- `d`: Supprimer (confirmation)
- `E`: Ouvrir dans éditeur externe ($EDITOR)
- `t`: Changer thème bordures
- `q`: Quitter

### TODO (à implémenter)
- `?`: Help modal (PRIORITÉ 1)
- `y`: Copy path du fichier JSON (PRIORITÉ 2)

## Stockage

### Structure JSON

```json
{
  "id": "uuid-v4",
  "title": "Fetch users",
  "category": "backend/db",
  "language": "sql",
  "tags": ["users", "postgres"],
  "content": "SELECT * FROM users;",
  "created_at": "2025-11-16T12:34:56Z",
  "updated_at": "2025-11-16T12:34:56Z",
  "path": "/absolute/path/to/file.json"
}
```

### Emplacement

1. **Priorité 1**: `$SNIPSTER_DIR` (override explicite)
2. **Priorité 2**: `~/.snipster/snippets/` (défaut user)
3. **Fallback**: `./.snipster/snippets/` (local si $HOME inaccessible)

### Organisation

- Hiérarchie réflétée dans `category` (ex: "backend/express/middleware")
- Un fichier JSON par snippet
- Nom de fichier: slug du titre (ex: `fetch-users.json`)

## Développement

### Commandes utiles

```bash
# Dev local
go run ./cmd/snip

# Build avec metadata
make build

# Tests
go test ./...

# Lint
go vet ./...

# Install local
make user-install  # installe dans ~/bin/snip
```

### Versioning

- Format: `v1.0.0` (SemVer strict)
- Tags Git déclenchent release automatique
- Metadata injectée via ldflags:
  - `internal/version.Version`
  - `internal/version.Commit`
  - `internal/version.Date`

### Release

```bash
# 1. Commit final
git add .
git commit -m "chore: prepare v1.0.0 release"
git push

# 2. Tag avec message détaillé
git tag -a v1.0.0 -m "v1.0.0 - Description

Features:
- Feature 1
- Feature 2
"

# 3. Push tag (déclenche GoReleaser)
git push origin v1.0.0

# 4. Vérifier release
gh release view v1.0.0
```

## Ajout de features

### Exemple: Ajouter une touche Help modal

1. **Définir l'état** dans `model.go`:
   ```go
   const (
       StateWelcome AppState = iota
       StateHome
       StateHelp  // NOUVEAU
       StateCreate
       // ...
   )
   ```

2. **Gérer la touche** dans `update.go`:
   ```go
   case "?":
       if m.State == StateHome {
           m.State = StateHelp
           return m, nil
       }
   ```

3. **Créer la vue** dans `view.go`:
   ```go
   case StateHelp:
       return m.viewHelp()

   func (m Model) viewHelp() string {
       // Rendu du modal help
   }
   ```

4. **Tester manuellement**:
   ```bash
   go run ./cmd/snip
   # Appuyer sur '?' pour ouvrir help
   ```

## Bonnes pratiques

### DO ✅

- Suivre l'architecture Bubble Tea (Model/Update/View)
- Garder les fichiers courts (<300 lignes)
- Utiliser `ui.Theme` pour tous les styles
- Logger les warnings, fatal pour erreurs critiques
- Tester manuellement chaque feature dans le TUI
- Commit messages: `feat:`, `fix:`, `chore:`, `docs:`

### DON'T ❌

- Ne jamais modifier `Model` directement hors de `Update()`
- Ne pas mixer logique métier et rendu dans `View()`
- Ne pas dupliquer styles (centraliser dans `ui/styles.go`)
- Ne pas laisser de TODOs dans le code (utiliser Issues GitHub)
- Ne pas commit sans tester `go build ./cmd/snip`

## Dépendances critiques

- **Bubble Tea**: Framework TUI principal (https://github.com/charmbracelet/bubbletea)
- **Lipgloss**: Styles et layout (https://github.com/charmbracelet/lipgloss)
- **Bubbles**: Composants (list, textarea, textinput, viewport)
- **clipboard**: Copie presse-papiers cross-platform
- **fuzzy**: Recherche fuzzy (sahilm/fuzzy)

## Roadmap

### Implémenté ✅
- [x] Explorateur hiérarchique avec breadcrumbs
- [x] Recherche `/` + fuzzy toggle `f`
- [x] Highlight matches (liste + preview)
- [x] CRUD complet (modals + édition externe)
- [x] Copie presse-papiers (`Enter`)
- [x] Thèmes cyclables (`t`)

### En cours 🔄
- [ ] Help modal (`?`) - PRIORITÉ 1
- [ ] Copy path (`y`) - PRIORITÉ 2

### Futur 🔮
- [ ] Bookmarks/Favoris (`b`)
- [ ] Récents (Ctrl+R)
- [ ] Tags avancés (filtrage)
- [ ] Export/Import
- [ ] Templates
- [ ] Distribution Homebrew

## Support

- **Issues**: https://github.com/HrodWolfS/snipster/issues
- **Auteur**: @hrodwolf
- **License**: MIT
