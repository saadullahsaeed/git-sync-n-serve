PROJECT = git-sync-static

GOTOOLS = \
	golang.org/x/tools/cmd/cover \
	golang.org/x/tools/cmd/goimports \

PREFIX ?= $(shell pwd)
SOURCE_FILES ?= ./...

TEST_PATTERN ?= .
TEST_OPTS ?=

GO111MODULE ?= off

.PHONY: setup
setup: ## Install dev tools
	@if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5; \
	fi
	go get $(GOTOOLS)

.PHONY: test
test: ## Run all the tests
	go test $(TEST_OPTS) -v -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run gofmt and goimports on all go files
	@find . -name '*.go' -not -wholename './proto/*' -not -wholename './vendor/*' -not -wholename './ui/*' -not -wholename './swagger/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	golangci-lint run

.PHONY: clean
clean: ## Remove built binaries
	go clean -i $(SOURCE_FILES)
	rm -rf ./bin/* ./dist/*

.PHONY: build
build: ## Build a local copy
	go build -o ./bin/$(PROJECT) ./cmd/$(PROJECT)/main.go
	GOOS=linux GOARCH=386 go build -o ./bin/$(PROJECT)_linux_386 ./cmd/$(PROJECT)/main.go

.PHONY: dev
dev: ## Build and run in development mode
	go run -tags=dev ./cmd/$(PROJECT)/main.go

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: container 
container:
	docker build -t $(REGISTRY):latest .

.DEFAULT_GOAL := help
default: help
