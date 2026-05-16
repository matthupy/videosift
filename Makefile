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

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint (must be installed separately)
lint:
	golangci-lint run

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf bin/

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //'
