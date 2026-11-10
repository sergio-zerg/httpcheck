BINARY  := httpcheck
PKG     := github.com/sergio-zerg/httpcheck
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint vuln fmt fmt-check vet tidy ci clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

fmt:
	gofmt -w .

fmt-check:
	@diff=$$(gofmt -l .); if [ -n "$$diff" ]; then echo "gofmt needed:"; echo "$$diff"; exit 1; fi

vet:
	go vet ./...

tidy:
	go mod tidy

ci: fmt-check vet lint test build

clean:
	rm -rf bin
