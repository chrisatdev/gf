BINARY    := gf
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X main.version=$(VERSION)"
GOFLAGS   :=

INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build install test lint cover snapshot clean

build:
	go build $(LDFLAGS) $(GOFLAGS) -o bin/$(BINARY) ./cmd/gf

install: build
	mkdir -p $(INSTALL_DIR)
	cp bin/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(INSTALL_DIR)/$(BINARY) ($(VERSION))"

test:
	go test ./...

lint:
	golangci-lint run ./...

cover:
	go test -cover ./...

clean:
	rm -rf bin/

snapshot:
	goreleaser release --snapshot --clean
