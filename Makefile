BINARY := snip
MODULE := github.com/HrodWolfS/snipster

# Auto-fill metadata (override with: make build VERSION=v0.1.0)
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE    ?= $(shell date -u +%FT%TZ)

LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) \
           -X $(MODULE)/internal/version.Commit=$(COMMIT) \
           -X $(MODULE)/internal/version.Date=$(DATE)

BIN_DIR := bin
PREFIX  ?= /usr/local
BUILD   := go build -ldflags "$(LDFLAGS)"

.PHONY: all build install user-install run clean version ensure-bin

all: build

ensure-bin:
	mkdir -p $(BIN_DIR)

build: ensure-bin
	$(BUILD) -o $(BIN_DIR)/$(BINARY) ./cmd/$(BINARY)

# System-wide install (may require sudo)
install:
	$(BUILD) -o $(BINARY) ./cmd/$(BINARY)
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	rm -f $(BINARY)

# User install in ~/bin (no sudo)
user-install:
	$(BUILD) -o $(HOME)/bin/$(BINARY) ./cmd/$(BINARY)
	@printf "\nInstalled to $(HOME)/bin/$(BINARY). Ensure ~/bin is in your PATH.\n"

run:
	go run ./cmd/$(BINARY)

version: build
	@./$(BIN_DIR)/$(BINARY) --version

clean:
	rm -rf $(BIN_DIR) $(BINARY)

