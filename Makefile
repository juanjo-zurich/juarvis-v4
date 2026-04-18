VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X juarvis/cmd.Version=$(VERSION) -X juarvis/cmd.Commit=$(COMMIT) -X juarvis/cmd.BuildDate=$(BUILD_DATE)
export CI := true
exportJUARVIS_SKIP_NETWORK := true

.PHONY: build test test-integration test-regression test-e2e test-verify test-all lint install clean sync-assets ci

build:
	go build -ldflags "$(LDFLAGS)" -o juarvis .

test:
	go test -timeout 60s -cover ./...

test-integration: build
	go test -v -tags=integration -timeout 60s ./tests/integration/...

test-regression: build
	go test -v -tags=regression -timeout 60s ./tests/regression/...

test-e2e: build
	go test -v -tags=e2e -timeout 120s ./tests/e2e/...

test-verify: build
	./juarvis verify

test-all: build test test-verify

# CI target: ejecuta solo lo que funciona en GitHub Actions
ci: build
	@echo "==> Running CI checks..."
	@echo "     - go vet..."
	@$(MAKE) vet
	@echo "     - unit tests..."
	@go test -timeout 60s ./...
	@echo "     - verify CLI..."
	@./juarvis verify
	@echo "✅ CI passed!"

fmt:
	gofmt -w -s .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

check: fmt vet lint

install: build
	@echo "Instalando juarvis en /usr/local/bin/..."
	@if [ -w /usr/local/bin ]; then \
		cp juarvis /usr/local/bin/; \
		echo "✅ Instalado globalmente."; \
	else \
		echo "❌ Error: No tienes permisos para escribir en /usr/local/bin/"; \
		echo "Ejecuta: sudo make install"; \
		exit 1; \
	fi

clean:
	rm -f juarvis

sync-assets:
	@echo "Sincronizando assets desde proyecto padre..."
	@if [ -f ../AGENTS.md ]; then cp ../AGENTS.md pkg/assets/data/; fi
	@if [ -f ../marketplace.json ]; then cp ../marketplace.json pkg/assets/data/; fi
	@if [ -f ../agent-settings.json ]; then cp ../agent-settings.json pkg/assets/data/; fi
	@if [ -f ../permissions.yaml ]; then cp ../permissions.yaml pkg/assets/data/; fi
	@echo "Assets sincronizados."
