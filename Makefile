.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/postiee-pp-cli$(BIN_EXT) ./cmd/postiee-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/postiee-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/postiee-pp-mcp$(BIN_EXT) ./cmd/postiee-pp-mcp

install-mcp:
	go install ./cmd/postiee-pp-mcp

build-all: build build-mcp
