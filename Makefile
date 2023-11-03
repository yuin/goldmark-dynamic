SOURCES = $(shell find . -type f -name '*.go')
RM=rm -f
GOCILINT=golangci-lint

test: cover.html ## Run tests

cover.html: $(SOURCES)
	go test -coverprofile cover.out .
	go tool cover -html cover.out -o cover.html

.PHONY: lint
lint: ## Run lints
	$(GOCILINT) run -c .golangci.yml ./...

.PHONY: clean
clean: ## Remove generated files
	$(RM) cmd/$(APP)/$(APP) 
	$(RM) cmd/$(APP)/$(APP).exe
	$(RM) cmd/$(APP)/$(APP)_macosx
	$(RM) cover.*

.PHONY: help
help: ## Display this help message
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
