.DEFAULT: help
.PHONY: helps deps tidy lint test gen clean

GOLANG_PATH=$(CURDIR)/.go
GOLANG_BIN=$(GOLANG_PATH)/bin
BUILD_DIRECTORY=$(CURDIR)/.builds
REPORT_DIRECTORY=$(CURDIR)/.reports

GOLANGCI_LINT_VERSION=1.60.3
MOCKERY_VERSION=2.45.0
GOTESTSUM_VERSION=1.12.0

GOTESTSUM_FORMAT_NAME?=pkgname

APP_VERSION?=$(shell git describe --always --tags | sed 's/^v//')

help: ## Shows makefile's help
	@grep -h "##" $(MAKEFILE_LIST) | grep -v grep | sed -e 's/\\$$//' | column -t -s '##'

deps: $(GOLANG_BIN)/golangci-lint $(GOLANG_BIN)/mockery $(GOLANG_BIN)/gotestsum # Setup project dependencies
$(GOLANG_BIN)/golangci-lint:
	GOPATH=$(GOLANG_PATH) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
$(GOLANG_BIN)/mockery:
	GOPATH=$(GOLANG_PATH) go install github.com/vektra/mockery/v2@v$(MOCKERY_VERSION)
$(GOLANG_BIN)/gotestsum:
	GOPATH=$(GOLANG_PATH) go install gotest.tools/gotestsum@v$(GOTESTSUM_VERSION)

tidy: deps ## Run go mod tidy
	GOPATH=$(GOLANG_PATH) go mod tidy

lint: deps ## Lint project source code
	$(GOLANG_BIN)/golangci-lint run -v

test: deps ## Run project tests
	@rm -rf $(REPORT_DIRECTORY)
	@mkdir $(REPORT_DIRECTORY)

	$(GOLANG_BIN)/gotestsum \
		--format=$(GOTESTSUM_FORMAT_NAME) \
		-- \
		-race \
		-timeout=5s \
		-coverprofile=$(REPORT_DIRECTORY)/cover \
		./...

	go tool cover -func $(REPORT_DIRECTORY)/cover

gen: deps ## Run code gen
	go generate ./...

clean: ## Cleanup project dependencies
	@rm -rf $(GOLANG_PATH) $(BUILD_DIRECTORY)
