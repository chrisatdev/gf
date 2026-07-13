BINARY    := gf
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X github.com/chrisatdev/gf/cmd/gf.version=$(VERSION)"
GOFLAGS   :=

.PHONY: build test lint cover snapshot clean

build:
	go build $(LDFLAGS) $(GOFLAGS) -o bin/$(BINARY) ./cmd/gf

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
