export PATH := $(PATH):$(GOPATH)/bin

INTEGRATION  		:= nr-entity-tag-sync
BINARY_NAME   		= $(INTEGRATION)
LAMBDA_BINARY_NAME	= bootstrap
BIN_FILES			:= ./cmd/nr-entity-tag-sync/...
LAMBDA_BIN_FILES	:= ./cmd/nr-entity-tag-sync-lambda/...

GIT_COMMIT = $(shell git rev-parse HEAD)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

LDFLAGS += -X main.gInterationVersion=$(GIT_TAG)
LDFLAGS += -X main.gGitCommit=${GIT_COMMIT}
LDFLAGS += -X main.gBuildDate=${BUILD_DATE}

all: build

build: clean compile compile-lambda

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: removing binaries..."
	@rm -rfv bin

bin/$(BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@go mod tidy
	@go build -v -ldflags '$(LDFLAGS)' -o bin/$(BINARY_NAME) $(BIN_FILES)

bin/$(LAMBDA_BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(LAMBDA_BINARY_NAME)..."
	@go mod tidy
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags '$(LDFLAGS)' \
		-tags lambda.norpc -o bin/$(LAMBDA_BINARY_NAME) $(LAMBDA_BIN_FILES)

compile: bin/$(BINARY_NAME)

compile-lambda: bin/$(LAMBDA_BINARY_NAME)

package-lambda: build
	@./scripts/lambda/build.sh

deploy-lambda: build
	@./scripts/lambda/deploy.sh

update-lambda: build
	@./scripts/lambda/update.sh

delete-lambda:
	@./scripts/lambda/delete.sh

.PHONY: all build clean compile compile-lambda package-lambda deploy-lambda update-lambda delete-lambda
