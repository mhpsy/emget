.PHONY: build test lint vet coverage clean run

BIN := bin/emget

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%d)

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/emget

test:
	go test ./...

vet:
	go vet ./...

lint: vet

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

run: build
	./$(BIN)

clean:
	rm -rf bin coverage.out
