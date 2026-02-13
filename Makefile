VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -X github.com/DanielTso/pixshift/internal/version.Version=$(VERSION) \
           -X github.com/DanielTso/pixshift/internal/version.Commit=$(COMMIT) \
           -X github.com/DanielTso/pixshift/internal/version.Date=$(DATE)

.PHONY: build build-static build-web build-all clean test lint help install bench coverage fmt vet docker package-deb package-rpm

build: ## Build binary (requires CGO_ENABLED=1)
	CGO_ENABLED=1 go build -ldflags '$(LDFLAGS)' -o pixshift ./cmd/pixshift
	@echo "Built: pixshift"

# Build with static linking (Linux/Windows)
build-static:
	CGO_ENABLED=1 go build -ldflags '$(LDFLAGS) -extldflags "-static"' -o pixshift ./cmd/pixshift
	@echo "Built: pixshift (static)"

# Cross-compile for Windows (requires mingw-w64)
build-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
		go build -ldflags '$(LDFLAGS) -extldflags "-static"' -o pixshift.exe ./cmd/pixshift
	@echo "Built: pixshift.exe"

clean: ## Remove build artifacts
	rm -f pixshift pixshift.exe coverage.out coverage.html

test: ## Run all tests
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: build ## Install to $GOPATH/bin
	cp pixshift $(GOPATH)/bin/ 2>/dev/null || cp pixshift $(HOME)/go/bin/

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

coverage: ## Generate HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt: ## Format code
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

build-web: ## Build frontend
	cd web && npm install && npm run build

build-all: build-web build ## Build frontend then backend

docker: ## Build Docker image
	docker build -t pixshift .

package-deb: build ## Build .deb package (requires nfpm)
	VERSION=$$(echo "$(VERSION)" | sed 's/^v//') ARCH=amd64 nfpm package --packager deb
	@echo "Built .deb package"

package-rpm: build ## Build .rpm package (requires nfpm)
	VERSION=$$(echo "$(VERSION)" | sed 's/^v//') ARCH=x86_64 nfpm package --packager rpm
	@echo "Built .rpm package"
