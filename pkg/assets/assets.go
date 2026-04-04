package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed data/*
var embeddedAssets embed.FS

// GetEmbeddedFS retorna el sistema de archivos de las reglas nativas compiladas dentro del binario
func GetEmbeddedFS() (fs.FS, error) {
	return fs.Sub(embeddedAssets, "data")
}

// CopyEmbeddedToDisk copia un directorio del embed.FS al filesystem
func CopyEmbeddedToDisk(srcPath, destPath string) error {
	return fs.WalkDir(embeddedAssets, filepath.Join("data", srcPath), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, filepath.Join("data", srcPath))
		if relPath == "" {
			relPath = "."
		}
		dest := filepath.Join(destPath, strings.TrimPrefix(relPath, string(filepath.Separator)))
		if relPath == "." {
			dest = destPath
		}
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		content, err := fs.ReadFile(embeddedAssets, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0644)
	})
}
