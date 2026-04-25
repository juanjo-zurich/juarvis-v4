#!/bin/bash
# validate-ci.sh - Valida el workflow CI localmente antes de push a GitHub
# Uso: ./scripts/validate-ci.sh [--install-act]

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "🔍 Validación CI Local - Juarvis"
echo "=========================================="

# 1. Validar YAML del workflow
echo -e "\n${YELLOW}1. Validando YAML del workflow...${NC}"
if command -v actionlint &> /dev/null; then
    actionlint .github/workflows/*.yml
    echo -e "${GREEN}✅ YAML válido${NC}"
else
    echo -e "${YELLOW}⚠️  actionlint no instalado - skipping validación YAML${NC}"
    echo "   Instalar: brew install actionlint"
fi

# 2. Validar Go formatting
echo -e "\n${YELLOW}2. Validando gofmt...${NC}"
if [ -n "$(gofmt -l .)" ]; then
    echo -e "${RED}❌ Archivos sin formatear:${NC}"
    gofmt -l .
    exit 1
else
    echo -e "${GREEN}✅ gofmt OK${NC}"
fi

# 3. Validar go vet
echo -e "\n${YELLOW}3. Validando go vet...${NC}"
if go vet ./... 2>&1; then
    echo -e "${GREEN}✅ go vet OK${NC}"
else
    echo -e "${RED}❌ go vet falló${NC}"
    exit 1
fi

# 4. Build local
echo -e "\n${YELLOW}4. Build local...${NC}"
cd "$PROJECT_ROOT"
export CI=true
export JUARVIS_SKIP_NETWORK=true
if make build; then
    echo -e "${GREEN}✅ Build OK${NC}"
else
    echo -e "${RED}❌ Build falló${NC}"
    exit 1
fi

# 5. Tests locales
echo -e "\n${YELLOW}5. Ejecutando tests...${NC}"
if go test -timeout 60s -cover ./... 2>&1; then
    echo -e "${GREEN}✅ Tests OK${NC}"
else
    echo -e "${RED}❌ Tests fallaron${NC}"
    exit 1
fi

# 6. Juarvis verify
echo -e "\n${YELLOW}6. Ejecutando juarvis verify...${NC}"
if ./juarvis verify; then
    echo -e "${GREEN}✅ Juarvis Verify OK${NC}"
else
    echo -e "${RED}❌ Juarvis Verify falló${NC}"
    exit 1
fi

# 7. Verificar lock files no estén en git (opcional)
echo -e "\n${YELLOW}7. Verificando lock files...${NC}"
cd "$PROJECT_ROOT"
LOCKED_FILES=$(git status --porcelain 2>/dev/null | grep '\.lock\.yml' || true)
if [ -n "$LOCKED_FILES" ]; then
    echo -e "${RED}❌ Archivos .lock.yml encontrados en git:${NC}"
    echo "$LOCKED_FILES"
    echo -e "${YELLOW}💡 Ejecutar: git rm --cached '**/*.lock.yml'${NC}"
    # No fallamos aquí porque es warning
else
    echo -e "${GREEN}✅ No hay lock files en git${NC}"
fi

# 8. Si install-act, ejecutar con act
if [ "$1" = "--install-act" ] || [ "$1" = "--act" ]; then
    echo -e "\n${YELLOW}8. Ejecutando GitHub Actions localmente con act...${NC}"
    if command -v act &> /dev/null; then
        cd "$PROJECT_ROOT"
        echo -e "${YELLOW}🔄 Esto puede tomar unos minutos en la primera ejecución...${NC}"
        if act --dryrun; then
            echo -e "${GREEN}✅ Workflow CI válido (dry-run OK)${NC}"
        else
            echo -e "${RED}❌ Workflow CI tiene errores${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  act no instalado${NC}"
        echo "   Instalar: brew install act"
        echo "   O ejecutar: curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    fi
fi

echo -e "\n=========================================="
echo -e "${GREEN}✅ Validación CI completa${NC}"
echo -e "${GREEN}Puedes hacer push a GitHub sin miedo${NC}"
echo "=========================================="