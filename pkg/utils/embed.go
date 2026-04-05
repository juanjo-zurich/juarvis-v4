package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyEmbeddedDir copia un directorio completo de un fs.FS al filesystem local.
func CopyEmbeddedDir(targetFS fs.FS, srcPath, destPath string) error {
	return fs.WalkDir(targetFS, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, srcPath)
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
		content, err := fs.ReadFile(targetFS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0644)
	})
}
