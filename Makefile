TARGET ?= pico2-w
FLAGS   = -stack-size=8kb
TAGS   ?=
PROG   ?= ./cmd/bottle

help: ## Show this help
	@echo "Valid targets: "; grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) |  sort |  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'
.PHONY: help

flash: ## Flash program to pico
	@tinygo flash -target=$(TARGET) -tags $(TAGS) $(FLAGS) $(PROG)
.PHONY: flash

generate: ## Generate build secrets
	@go generate ./...
.PHONY: generate

build-client: generate ## Build desktop GUI client
	@go build ./cmd/gui
.PHONY: build-client

build-client-headless: generate ## Build headless desktop client
	@go build ./cmd/client
.PHONY: build-client-headless

build: generate ## Build firmware
	@tinygo build -target=$(TARGET) -tags $(TAGS) $(FLAGS) -o main.elf $(PROG)
.PHONY: build

build-uf2: generate ## Build UF2 file for flashing
	@tinygo build -target=$(TARGET) -tags $(TAGS) $(FLAGS) -o main.uf2 $(PROG)
.PHONY: build-uf2

clean: ## Remove all build artifacts
	@find . -maxdepth 1 -type f -executable -exec sh -c 'echo {} && rm {}' \;
	@find . -type f -name '*.pem' -not -path './backend/.venv/*' -exec sh -c 'echo {} && rm {}' \;
.PHONY: clean

.DEFAULT: help
