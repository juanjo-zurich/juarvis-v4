VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -X juarvis/cmd.Version=$(VERSION) -X juarvis/cmd.Commit=$(COMMIT) -X juarvis/cmd.BuildDate=$(BUILDDATE)

.PHONY: build test test-integration test-all lint install clean sync-assets

build:
	go build -ldflags "$(LDFLAGS)" -o juarvis .

test:
	go test -v -cover ./...

test-integration: build
	go test -v -tags=integration ./tests/integration/...

test-all: test test-integration

lint:
	go vet ./...

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
