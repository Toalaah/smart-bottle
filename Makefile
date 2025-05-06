TARGET ?= pico2-w
FLAGS   = -stack-size=32kb
PROG   ?= ./cmd/bottle

help: ## Show this help.
	@echo "Valid targets: "; grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) |  sort |  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'
.PHONY: help

flash: ## Flash program to pico.
	@tinygo flash -target=$(TARGET) $(FLAGS) $(PROG)
.PHONY: flash

build-client: ## Build desktop client.
	@go build ./cmd/client
.PHONY: build-client

build: ## Build executable.
	@tinygo build -target=$(TARGET) $(FLAGS) -o main.elf $(PROG)
.PHONY: build

build-uf2: ## Build UF2 file for flashing.
	@tinygo build -target=$(TARGET) $(FLAGS) -o main.uf2 $(PROG)
.PHONY: build-uf2

clean: ## Remove all build artifacts
	@find . -maxdepth 1 -type f -executable -exec sh -c 'echo {} && rm {}' \;
	@find . -type f -name '*.pem' -not -path './backend/.venv/*' -exec sh -c 'echo {} && rm {}' \;
.PHONY: clean

.DEFAULT: help
