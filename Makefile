# Rishi AI RStudio Add-in Makefile
# Simple build system for local development and distribution

# ==============================================================================
# Configuration
# ==============================================================================

PACKAGE_NAME := rishi
VERSION := $(shell cat addin/DESCRIPTION | grep Version | cut -d' ' -f2)
GITHUB_USER := Omkar-Waingankar
REPO_NAME := rishi

GO := go
BINARY_NAME := rishi-daemon
DIST_DIR := dist
DAEMON_DIR := daemon
ADDIN_DIR := addin

# Cross-compilation targets
PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64 windows/amd64

# ==============================================================================
# Main Commands
# ==============================================================================

.PHONY: help up package install install-local uninstall-local package-all clean

help: ## Show available commands
	@echo "Rishi AI RStudio Add-in"
	@echo ""
	@echo "Commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

up: ## One command for local dev: build daemon, install addin, launch RStudio
	@echo "Setting up local development environment..."
	@$(MAKE) _install-deps
	@$(MAKE) _build-daemon
	@$(MAKE) _build-addin
	@$(MAKE) _install-addin
	@$(MAKE) _launch-rstudio
	@echo "✓ Development environment ready!"

package: ## Create distribution packages in dist/
	@echo "Creating distribution packages..."
	@rm -rf $(DIST_DIR) && mkdir -p $(DIST_DIR)
	@$(MAKE) _build-daemon
	@$(MAKE) _build-addin
	@$(MAKE) _package-r-source
	@$(MAKE) _package-daemon
	@echo "✓ Packages created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

install-local: package ## Install from built packages (simulates user installation)
	@echo "Installing from local packages..."
	@R -e "install.packages('$(DIST_DIR)/$(PACKAGE_NAME)_$(VERSION).tar.gz', repos=NULL, type='source')"
	@echo "✓ Installed $(PACKAGE_NAME) from local tarball"

uninstall-local: ## Uninstall the locally installed package
	@echo "Uninstalling $(PACKAGE_NAME)..."
	@R -e "if ('$(PACKAGE_NAME)' %in% rownames(installed.packages())) { remove.packages('$(PACKAGE_NAME)'); cat('✓ Uninstalled $(PACKAGE_NAME)\n') } else { cat('$(PACKAGE_NAME) is not installed\n') }"

release: ## Create and publish a new GitHub release with binaries
	@echo "Creating new release v$(VERSION)..."
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION not found in DESCRIPTION"; exit 1; fi
	@$(MAKE) package-all
	@git tag -a "v$(VERSION)" -m "Release v$(VERSION)" || (echo "Tag already exists. Delete with: git tag -d v$(VERSION)" && exit 1)
	@git push origin "v$(VERSION)"
	@echo ""
	@echo "Creating GitHub release..."
	@gh release create "v$(VERSION)" \
		--title "v$(VERSION)" \
		--generate-notes \
		dist/rishi-daemon-linux-amd64.tar.gz \
		dist/rishi-daemon-darwin-amd64.tar.gz \
		dist/rishi-daemon-darwin-arm64.tar.gz \
		dist/rishi-daemon-windows-amd64.tar.gz
	@echo ""
	@echo "✓ Release v$(VERSION) published!"
	@echo "Users can now install with: remotes::install_github(\"$(GITHUB_USER)/$(REPO_NAME)\", subdir = \"addin\")"

package-all: ## Cross-compile daemon for all platforms and create complete distribution
	@echo "Creating complete distribution for all platforms..."
	@rm -rf $(DIST_DIR) && mkdir -p $(DIST_DIR)
	@$(MAKE) _build-addin
	@$(MAKE) _package-r-source
	@$(MAKE) _package-daemon-all
	@echo "✓ Complete distribution created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@rm -f $(DAEMON_DIR)/bin/$(BINARY_NAME)
	@rm -rf $(ADDIN_DIR)/node_modules
	@rm -f $(ADDIN_DIR)/inst/www/chat-app.js*
	@rm -f $(ADDIN_DIR)/*.tar.gz
	@cd $(DAEMON_DIR) && $(GO) clean
	@echo "✓ Build artifacts cleaned"

# ==============================================================================
# Internal Build Steps (prefixed with _ to indicate they're internal)
# ==============================================================================

_install-deps:
	@echo "Installing dependencies..."
	@command -v go >/dev/null || { echo "Error: Go required"; exit 1; }
	@command -v R >/dev/null || { echo "Error: R required"; exit 1; }
	@command -v npm >/dev/null || { echo "Error: npm required"; exit 1; }
	@R -e "install.packages(c('rstudioapi', 'httpuv', 'httr', 'tools', 'plumber', 'jsonlite', 'shiny', 'miniUI', 'htmltools', 'websocket', 'remotes'), repos='https://cran.rstudio.com/', quiet=TRUE)" >/dev/null
	@cd $(ADDIN_DIR) && npm install --silent

_build-daemon:
	@echo "Building daemon..."
	@mkdir -p $(DAEMON_DIR)/bin
	@cd $(DAEMON_DIR) && $(GO) build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd/server

_build-addin:
	@echo "Building React frontend..."
	@cd $(ADDIN_DIR) && npm run build --silent
	@echo "Building daemon binaries for all platforms..."
	@$(MAKE) _build-daemon-all-platforms

_install-addin:
	@echo "Installing R addin..."
	@R CMD INSTALL $(ADDIN_DIR) --quiet

_launch-rstudio:
	@echo "Launching RStudio..."
	@if command -v rstudio >/dev/null 2>&1; then \
		rstudio >/dev/null 2>&1 & \
	elif [ -f "/Applications/RStudio.app/Contents/MacOS/RStudio" ]; then \
		open -a RStudio; \
	else \
		echo "Error: RStudio not found"; exit 1; \
	fi

_package-r-source:
	@echo "Creating R source package..."
	@cd $(ADDIN_DIR) && R CMD build . --quiet
	@mv $(ADDIN_DIR)/$(PACKAGE_NAME)_*.tar.gz $(DIST_DIR)/

_package-daemon:
	@echo "Packaging daemon binary..."
	@cp $(DAEMON_DIR)/bin/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz $(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)
	@rm $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)

_build-daemon-all-platforms:
	@mkdir -p $(ADDIN_DIR)/inst/bin
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "  Building daemon for $$os/$$arch..."; \
		(cd $(DAEMON_DIR) && GOOS=$$os GOARCH=$$arch $(GO) build -ldflags="-s -w" -o ../$(ADDIN_DIR)/inst/bin/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/server); \
	done

_package-daemon-all:
	@echo "Cross-compiling daemon for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "  Building for $$os/$$arch..."; \
		(cd $(DAEMON_DIR) && GOOS=$$os GOARCH=$$arch $(GO) build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/server); \
		(cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$$os-$$arch.tar.gz $(BINARY_NAME)-$$os-$$arch$$ext); \
		rm $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext; \
	done

# Default target
.DEFAULT_GOAL := help