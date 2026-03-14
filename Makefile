BINARY     := snowctl
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
              -X github.com/mjamalu/snowctl/cmd.version=$(VERSION) \
              -X github.com/mjamalu/snowctl/cmd.commit=$(COMMIT) \
              -X github.com/mjamalu/snowctl/cmd.buildDate=$(BUILD_DATE)

.PHONY: build test lint clean install fmt

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/ coverage.out

fmt:
	gofmt -w .
	goimports -w .
