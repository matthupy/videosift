MODULE    = github.com/matthupy/videosift
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOEXE    := $(shell go env GOEXE)
BINARY    = bin/extract$(GOEXE)
LDFLAGS   = -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: all build install test vet lint clean tidy

all: build

## build: compile the extract CLI into ./bin/
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/extract

## install: install the extract CLI to GOBIN
install:
	go install $(LDFLAGS) ./cmd/extract

## test: run all tests with race detector
test:
	go test ./... -race -timeout 120s

## vet: run go vet (standard Go static analysis)
vet:
	go vet ./...

## lint: run go vet (standard Go static analysis)
# This project uses go vet instead of external linters like golangci-lint to reduce
# external dependencies and simplify the build process. go vet catches critical bugs,
# unreachable code, type mismatches, and printf errors. For style conventions and
# additional checks, see README for details.
lint:
	go vet ./...

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf bin/

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //'
