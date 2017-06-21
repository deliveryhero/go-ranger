# force bash to use build-ins
SHELL:=bash

PROJECT     := github.com/foodora/go-ranger

# runtime options
VERSION     := $(shell date +%Y%m%d_%H%M%S)
GOVERSION   := $(shell go version | sed 's/^go version //')

# go options
GO          ?= go
FLAGS       ?=
SOURCES     := $(shell find . -type f -name "*.go" | grep -v "^./vendor/")
PKGS        := $(shell glide novendor 2>/dev/null || echo .)
LDFLAGS     := -s -X "main.version=$(VERSION)" -X "main.goVersion=$(GOVERSION)"

.PHONY: help
help: ## Display this help
	@ echo "Please use \`make <target>' where <target> is one of:"
	@ echo
	@ grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-10s\033[0m - %s\n", $$1, $$2}'
	@ echo

.PHONY: all
all: deps build ## Install dependencies and build binaries

.PHONY: build
build: clean fmt ## Build binaries
	GOARCH=$(ARCH) CGO_ENABLED=0 $(GO) build $(FLAGS) -ldflags '$(LDFLAGS)' -o $(CURDIR)/bin/feed ./cmd/...

.PHONY: test
test: fmt-check lint test-unit ## Execute all checks and tests

.PHONY: test-unit
test-unit: ## Execute unit tests
	$(GO) test -race -short -v $(FLAGS) $(PKGS)

.PHONY: cover
cover: ## Execute unit tests with coverage
	overalls -project=$(PROJECT) -covermode=atomic -debug -- -race $(FLAGS)
	$(GO) tool cover -func=overalls.coverprofile
	@ rm -f profile.coverprofile

.PHONY: cover-html
cover-html: cover ## Generate and show HTML coverage report
	$(GO) tool cover -html=overalls.coverprofile

.PHONY: lint
lint: ## Perform lint checks (golint, go vet)
	golint -set_exit_status .
	go vet .

.PHONY: fmt
fmt: ## Fix formatting (goimports)
	goimports -w $(SOURCES)

.PHONY: fmt-check
fmt-check: ## Check formatting (goimports)
	@ echo goimports -e -l .
	@ diff -u \
		--label "ERROR: unformatted files detected" <(printf "") \
		--label "Run <make fmt> to fix formatting" <(goimports -e -l $(SOURCES))

HAS_GLIDE   := $(shell command -v glide)

.PHONY: deps
deps: ## Install dependencies
	@ echo "Installing tools ... "
ifndef HAS_GLIDE
	$(GO) get $(FLAGS) github.com/Masterminds/glide
endif
	$(GO) get $(FLAGS) github.com/golang/lint/golint
	$(GO) get $(FLAGS) golang.org/x/tools/cmd/goimports
	$(GO) get $(FLAGS) github.com/go-playground/overalls

	@ echo "Installing dependencies ... "
	glide install

.PHONY: clean
clean: ## Cleanup runtime files
	rm -rf bin *.coverprofile *.out

.PHONY: clean-all
clean-all: clean ## Cleanup ALL runtime files
	rm -rf vendor