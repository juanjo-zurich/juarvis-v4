#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

ASSETS_SRC="${ASSETS_SRC:-$PROJECT_DIR/..}"
ASSETS_DST="$PROJECT_DIR/pkg/assets/data"

echo "Sincronizando assets desde $ASSETS_SRC..."

for file in AGENTS.md marketplace.json opencode.json permissions.yaml; do
	if [ -f "$ASSETS_SRC/$file" ]; then
		cp -f "$ASSETS_SRC/$file" "$ASSETS_DST/$file"
		echo "  ✅ $file"
	else
		echo "  ⚠️  $file no encontrado en $ASSETS_SRC"
	fi
done

echo "Assets sincronizados en $ASSETS_DST"
