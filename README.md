# Snipster ✂️

[![CI](https://github.com/HrodWolfS/snipster/workflows/CI/badge.svg)](https://github.com/HrodWolfS/snipster/actions/workflows/ci.yml)
[![Release](https://github.com/HrodWolfS/snipster/workflows/Release/badge.svg)](https://github.com/HrodWolfS/snipster/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/HrodWolfS/snipster)](https://goreportcard.com/report/github.com/HrodWolfS/snipster)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hrodwolf/snipster)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/HrodWolfS/snipster)](https://github.com/HrodWolfS/snipster/releases/latest)

> Un gestionnaire de snippets local, élégant et rapide pour le terminal, construit avec Go et Bubble Tea.

Snipster est un TUI pour organiser, rechercher et prévisualiser vos snippets de code stockés en JSON sur le disque. Il propose un explorateur de dossiers à gauche, un aperçu du code à droite, une recherche instantanée avec mode « / », et un CRUD simple via modals.

---

## ✨ Fonctionnalités

- 🎨 Interface TUI nette – Mise en page en cadre principal avec header, body en 2 colonnes (explorateur/aperçu) et footer.
- 📂 Explorateur dossiers/fichiers – Icônes 📁/📄, navigation par dossiers (gauche/droite), breadcrumbs dans le header.
- 🔎 Recherche rapide – Touche `/` pour activer, « contains » par défaut + bascule fuzzy (`f`), highlight des matches dans la liste et l’aperçu, gouttière « ▶ » sur lignes correspondantes.
- 🧠 Aperçu code – Header (titre/catégorie/langage/tags) et coloration simple par regex (js/ts/go/sql).
- ✏️ CRUD via modals – `n` créer, `e` éditer, `d` supprimer (confirmation), `Ctrl+S` sauvegarder, `Enter` dans contenu ajoute une ligne (pas de submit).
- 📋 Presse‑papiers – `Enter` copie le contenu du snippet courant.
- 🖊️ Édition externe – `E` ouvre le JSON dans `$VISUAL`/`$EDITOR` (sinon `nano`), puis reload.
- 🧵 Thème et bordures – Fond transparent, bordures visibles; `t` cycle la couleur (cyan/rose/vert/orange).
- 🖥️ Écran d’accueil – ASCII « SNIPSTER » (orange), centré, avec cadre.

---

## 📦 Installation

> Remplacez `hrodwolf/snipster` par l’URL finale de votre dépôt si besoin.

### Via `go install` (recommandé)

```bash
go install github.com/HrodWolfS/snipster/cmd/snip@latest
```

Le binaire `snip` sera installé dans `$GOPATH/bin` (souvent `~/go/bin`).

### Installation manuelle

```bash
# Cloner le dépôt
git clone https://github.com/HrodWolfS/snipster.git
cd snipster

# Compiler le binaire court
go build -o snip ./cmd/snip

# Installer dans /usr/local/bin (optionnel)
sudo mv snip /usr/local/bin/

# Ou installer dans ~/bin
mkdir -p ~/bin
mv snip ~/bin/
export PATH="$HOME/bin:$PATH"  # Ajouter à ~/.bashrc ou ~/.zshrc
```

### Avec Makefile (optionnel)

```bash
make build           # construit bin/snip avec version/commit/date
sudo make install    # installe dans /usr/local/bin/snip
make user-install    # installe dans ~/bin/snip (sans sudo)
make version         # affiche la version du binaire
```

### Vérifier l'installation

```bash
snip --version
```

---

## 🚀 Utilisation

### Démarrage rapide

```bash
# Lancer avec le stockage par défaut (~/.snipster/snippets)
snip

# Lancer en pointant un répertoire de snippets
SNIPSTER_DIR="$HOME/mes-snippets" snip
```

### Raccourcis

| Touche          | Action                          |
| --------------- | ------------------------------- |
| `↑` `↓` `j` `k` | Naviguer dans la liste          |
| `→` `l`         | Entrer dans un dossier          |
| `←` `h`         | Remonter au dossier parent      |
| `/`             | Activer la barre de recherche   |
| `f`             | Basculer recherche fuzzy        |
| `Esc`           | Quitter/vider la recherche      |
| `Enter`         | Copier le contenu du snippet    |
| `n`             | Nouveau snippet (modal)         |
| `e`             | Éditer (modal)                  |
| `d`             | Supprimer (confirmation)        |
| `E`             | Ouvrir dans l’éditeur externe   |
| `t`             | Changer la couleur des bordures |
| `q`             | Quitter                         |

---

## 🗃️ Stockage & Format

- Racine: `~/.snipster/snippets/` (ou via `SNIPSTER_DIR`).
- Fallback sandbox: `./.snipster/snippets/` si `$HOME` n’est pas accessible.
- Un fichier JSON par snippet.

Exemple de fichier JSON:

```json
{
  "id": "d8c2b8a1-3c9a-4d2b-9f2a-1e5c4f6b7a8c",
  "title": "Fetch users",
  "category": "backend/db",
  "language": "sql",
  "tags": ["users", "postgres"],
  "content": "SELECT * FROM users;",
  "created_at": "2025-11-16T12:34:56Z",
  "updated_at": "2025-11-16T12:34:56Z",
  "path": "/Users/you/.snipster/snippets/backend/db/fetch-users.json"
}
```

---

## 🎨 Aperçu (ASCII)

```
┌──────────────────────────────────────────────────────────────────────────┐
│                                SNIPSTER                                  │
│                        Press any key to continue…                        │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│ /backend/db                                                              │
│ ┌──────────────┬───────────────────────────────────────────────────────┐ │
│ │ 📁 queries/  │  -- Fetch users (sql)                                 │ │
│ │ 📄 users.json│  SELECT * FROM users WHERE ...                        │ │
│ │ 📄 auth.json │▶ SELECT id, email FROM auth ...                       │ │
│ └──────────────┴───────────────────────────────────────────────────────┘ │
│  Search: use                         • t border • / search • q quit      │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 🛠️ Développement

### Prérequis

- Go 1.22 ou supérieur
- Git

### Cloner et compiler

```bash
git clone https://github.com/HrodWolfS/snipster.git
cd snipster
go mod download
go build -o snip ./cmd/snip
```

### Lancer en mode développement

```bash
go run ./cmd/snip
```

### Dépendances

- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)
- [atotto/clipboard](https://github.com/atotto/clipboard)
- [sahilm/fuzzy](https://github.com/sahilm/fuzzy)

### Structure du projet

```
snipster/
├── cmd/snip/                  # Point d’entrée
│   └── main.go
├── internal/model/            # État, update, view
│   ├── model.go
│   ├── update.go
│   └── view.go
├── internal/ui/               # Styles et composants UI
│   ├── styles.go
│   ├── list.go
│   ├── input.go
│   └── code.go
├── internal/snippets/         # Chargement/écriture des snippets
│   ├── snippet.go
│   ├── loader.go
│   └── writer.go
├── go.mod
├── go.sum
└── README.md
```

---

## 🤝 Contribution

Les contributions sont les bienvenues !

1. Fork le projet
2. Crée une branche (`git checkout -b feature/ma-feature`)
3. Commit (`git commit -m 'feat: add my feature'`)
4. Push (`git push origin feature/ma-feature`)
5. Ouvre une Pull Request

---

## 🐛 Bugs & Suggestions

Ouvre une issue: https://github.com/HrodWolfS/snipster/issues

---

## 📝 Roadmap (extrait)

- [x] Explorateur dossiers/fichiers avec breadcrumbs
- [x] Recherche `/` (contains) + fuzzy toggle `f`
- [x] Highlight des matches (liste + preview)
- [x] CRUD via modals + édition externe `E`
- [x] Copie au presse‑papiers (`Enter`)
- [x] Thème: cycle couleur de bordures `t`
- [ ] Export / import de snippets
- [ ] Templates de snippets
- [ ] Tags avancés (filtrage, nuage de tags)
- [ ] Synchronisation (iCloud/Dropbox)
- [ ] Partage (gist) / intégrations
- [ ] Distribution Homebrew (tap)

---

## 📜 Licence

MIT. Voir le fichier [LICENSE](LICENSE).

---

## 👤 Auteur

**hrodwolf** — https://github.com/hrodwolf

⭐ Si ce projet vous plaît, laissez une étoile sur GitHub !
