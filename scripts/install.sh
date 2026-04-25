#!/bin/bash
set -e

JUARVIS_VERSION="${JUARVIS_VERSION:-dev}"
JUARVIS_REPO="${JUARVIS_REPO:-https://github.com/juanjo-zurich/juarvis-v4}"

detect_os() {
  local os
  case "$OSTYPE" in
    darwin*)  os="macos" ;;
    linux-gnu*) os="linux" ;;
    linux-musl*) os="linux" ;;
    msys*|cygwin*|mingw*) os="windows" ;;
    *) os="unknown" ;;
  esac
  echo "$os"
}

detect_arch() {
  local arch
  case $(uname -m) in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    armv7l) arch="arm" ;;
    *) arch="unknown" ;;
  esac
  echo "$arch"
}

get_install_dir() {
  local dir
  
  if [ -w "$HOME/.local/bin" ]; then
    dir="$HOME/.local/bin"
  elif [ -w "/usr/local/bin" ]; then
    dir="/usr/local/bin"
  elif [ -w "$HOME/bin" ]; then
    dir="$HOME/bin"
  else
    return 1
  fi
  
  echo "$dir"
  return 0
}

add_to_path() {
  local shell_rc=""
  case "${SHELL##*/}" in
    zsh) shell_rc="$HOME/.zshrc" ;;
    bash)
      if [ -f "$HOME/.bashrc" ]; then
        shell_rc="$HOME/.bashrc"
      elif [ -f "$HOME/.bash_profile" ]; then
        shell_rc="$HOME/.bash_profile"
      fi
      ;;
    fish)
      if [ -d "$HOME/.config/fish" ]; then
        fish_path="$HOME/.config/fish/conf.d/juarvis.fish"
        mkdir -p "$(dirname "$fish_path")"
        echo "set -gx PATH $HOME/.local/bin \$PATH" > "$fish_path"
        return 0
      fi
      ;;
  esac
  
  if [ -z "$shell_rc" ]; then
    return 1
  fi
  
  if ! grep -q '\.local/bin' "$shell_rc" 2>/dev/null; then
    echo "" >> "$shell_rc"
    echo "# Juarvis" >> "$shell_rc"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$shell_rc"
  fi
  
  return 0
}

install_juarvis() {
  local os
  local arch
  local install_dir
  
  os=$(detect_os)
  arch=$(detect_arch)
  
  echo "🔍 Detectado: $os ($arch)"
  
  if ! install_dir=$(get_install_dir); then
    echo "❌ Error: No hay directorio escribible para instalar"
    echo ""
    echo "Opciones:"
    echo "  1. mkdir -p \$HOME/.local/bin"
    echo "  2. export PATH=\$HOME/.local/bin:\$PATH"
    echo "  3. sudo make install"
    return 1
  fi
  
  echo "📦 Instalando en: $install_dir"
  
  BIN_PATH="$(cd "$(dirname "$0")/.." && pwd)/juarvis"
  
  if [ -f "$BIN_PATH" ]; then
    cp "$BIN_PATH" "$install_dir/juarvis"
  else
    echo "⚠️  Binario no encontrado en repo, descargando..."
    
    local ext=""
    if [ "$os" = "windows" ]; then
      ext=".exe"
    fi
    
    local download_url="$JUARVIS_REPO/releases/download/$JUARVIS_VERSION/juarvis-${os}-${arch}${ext}"
    
    if ! curl -sSL "$download_url" -o "$install_dir/juarvis$ext"; then
      echo "❌ Error descargando desde: $download_url"
      return 1
    fi
    
    chmod +x "$install_dir/juarvis$ext"
  fi
  
  chmod +x "$install_dir/juarvis"
  
  if add_to_path; then
    echo "✅ Añadido al PATH en ~/.local/bin"
  fi
  
  echo ""
  echo "✅ Juarvis instalado!"
  echo "   Ejecuta: juarvis --version"
  echo ""
  echo "   Para activar en esta terminal:"
  echo "   $ source ~/.bashrc  # o ~/.zshrc"
  
  return 0
}

uninstall_juarvis() {
  local install_dir
  
  if ! install_dir=$(get_install_dir); then
    echo "❌ No se encontró instalación"
    return 1
  fi
  
  rm -f "$install_dir/juarvis"
  echo "✅ Desinstalado de: $install_dir"
}

show_help() {
  cat <<EOF
Juarvis Installer

Uso: $(basename "$0") [comando]

Comandos:
  install     Instalar Juarvis (default)
  uninstall   Desinstalar Juarvis
  help        Mostrar esta ayuda

Opciones de entorno:
  JUARVIS_VERSION    Versión a instalar (default: dev)
  JUARVIS_REPO     Repo GitHub (default: $JUARVIS_REPO)

EOF
}

main() {
  local cmd="${1:-install}"
  
  case "$cmd" in
    install) install_juarvis ;;
    uninstall) uninstall_juarvis ;;
    help|--help|-h) show_help ;;
    *) echo "Comando desconocido: $cmd"; show_help; exit 1 ;;
  esac
}

main "$@"