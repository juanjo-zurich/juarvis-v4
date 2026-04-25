VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X juarvis/cmd.Version=$(VERSION) -X juarvis/cmd.Commit=$(COMMIT) -X juarvis/cmd.BuildDate=$(BUILDDATE)
export CI := true
export JUARVIS_SKIP_NETWORK := true

.PHONY: build test test-integration test-regression test-e2e test-verify test-all lint install clean sync-assets sync-plugins ci ci-local ci-validate ci-install-act

# Core build target
build: sync-plugins
	go build -ldflags "$(LDFLAGS)" -o juarvis .

sync-plugins:
	@rm -rf pkg/assets/data/plugins
	@cp -r plugins/ pkg/assets/data/plugins/
	@echo "✅ $(shell ls plugins/ | wc -l | tr -d ' ') plugins sincronizados"

# Targets de CI
ci: build test test-verify lint
	@echo "✅ CI pipeline completo"

ci-local: ci-validate
	@echo "✅ Validación CI local completada"
	@echo ""
	@echo "💡 Para ejecutar workflow completo con act:"
	@echo "   make ci-act"

ci-validate:
	@./scripts/validate-ci.sh

ci-install-act:
	@echo "📦 Instalando act (GitHub Actions local runner)..."
	@if command -v brew &> /dev/null; then \
		brew install act; \
	elif command -v curl &> /dev/null; then \
		curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash; \
	else \
		echo "❌ No se pudo instalar act"; \
		echo "   Instalar manualmente: https://nektosact.com/installation/"; \
		exit 1; \
	fi
	@echo "✅ act instalado"
	@echo ""
	@echo "💡 Para validar workflow localmente:"
	@echo "   make ci-local"

ci-act:
	@if ! command -v act &> /dev/null; then \
		echo "❌ act no instalado"; \
		echo "   Ejecutar: make ci-install-act"; \
		exit 1; \
	fi
	@echo "🔄 Ejecutando workflow CI con act..."
	act --workflows .github/workflows --dryrun
	@echo "✅ Workflow CI validado con act"
