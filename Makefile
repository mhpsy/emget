.PHONY: build test lint vet coverage clean run

BIN := bin/emget

build:
	@mkdir -p bin
	go build -o $(BIN) ./cmd/emget

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
