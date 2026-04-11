.PHONY: build test lint install

build:
	go build -o gf ./cmd/gf

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/gf

.DEFAULT_GOAL := build
