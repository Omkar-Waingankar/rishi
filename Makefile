# Makefile for Tibbl add-in with local daemon backend

GO := go
BINARY := daemon/bin/tibbl-daemon
CMD_DIR := ./cmd/server
ADDIN_DIR := ./addin

.PHONY: all build-server run-server start-server clean-server tidy-server fmt-server test-server install-addin build-addin addin-deps addin-dev r-deps

all: build-server

build-server:
	mkdir -p $(dir $(BINARY))
	cd daemon && $(GO) build -o ../$(BINARY) $(CMD_DIR)

run-server:
	@echo "Starting server"
	cd daemon && $(GO) run $(CMD_DIR)

# Run the compiled binary (requires `make build-server` first)
start-server: build-server
	@echo "Starting binary"
	./$(BINARY)

clean-server:
	rm -f $(BINARY)

tidy-server:
	cd daemon && $(GO) mod tidy

fmt-server:
	cd daemon && $(GO) fmt ./...

test-server:
	cd daemon && $(GO) test ./...

# Add-in targets
r-deps:
	@echo "Installing R package dependencies"
	@if command -v R >/dev/null 2>&1; then \
		R -e "install.packages(c('rstudioapi', 'httpuv', 'tools'), repos='https://cran.rstudio.com/')"; \
	else \
		echo "Error: R is not installed or not in PATH"; \
		exit 1; \
	fi

addin-deps:
	@echo "Installing Node.js dependencies for React app"
	cd $(ADDIN_DIR) && npm install

build-addin: addin-deps
	@echo "Building React app for RStudio add-in"
	cd $(ADDIN_DIR) && npm run build

install-addin: r-deps build-addin
	@echo "Installing RStudio add-in"
	@if command -v R >/dev/null 2>&1; then \
		R CMD INSTALL $(ADDIN_DIR); \
	else \
		echo "Error: R is not installed or not in PATH"; \
		exit 1; \
	fi

addin-dev: addin-deps
	@echo "Starting development mode for React app (watch mode)"
	cd $(ADDIN_DIR) && npm run dev

clean-addin:
	@echo "Cleaning add-in build artifacts"
	rm -rf $(ADDIN_DIR)/node_modules
	rm -f $(ADDIN_DIR)/inst/www/chat-app.js

clean: clean-addin
	rm -f $(BINARY)
