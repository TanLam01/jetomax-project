.PHONY: run run-mcp test fmt tidy build

run:
	go run ./cmd/api

run-mcp:
	go run ./cmd/mcp

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

build:
	go build ./cmd/api ./cmd/mcp
