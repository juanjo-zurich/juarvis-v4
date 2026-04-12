VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -X juarvis/cmd.Version=$(VERSION) -X juarvis/cmd.Commit=$(COMMIT) -X juarvis/cmd.BuildDate=$(BUILDDATE)

.PHONY: build test test-integration test-regression test-e2e test-verify test-all lint install clean sync-assets

build:
	go build -ldflags "$(LDFLAGS)" -o juarvis .

test:
	go test -v -cover ./...

test-integration: build
	go test -v -tags=integration ./tests/integration/...

test-regression: build
	go test -v -tags=regression ./tests/regression/...

test-e2e: build
	go test -v -tags=e2e -timeout 120s ./tests/e2e/...

test-verify: build
	./juarvis verify

test-all: test test-integration test-regression test-e2e test-verify

fmt:
	gofmt -w -s .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

check: fmt vet lint

install: build
	cp juarvis /usr/local/bin/

clean:
	rm -f juarvis

sync-assets:
	@echo "Sincronizando assets desde proyecto padre..."
	@if [ -f ../AGENTS.md ]; then cp ../AGENTS.md pkg/assets/data/; fi
	@if [ -f ../marketplace.json ]; then cp ../marketplace.json pkg/assets/data/; fi
	@if [ -f ../opencode.json ]; then cp ../opencode.json pkg/assets/data/; fi
	@if [ -f ../permissions.yaml ]; then cp ../permissions.yaml pkg/assets/data/; fi
	@echo "Assets sincronizados."
