package assets

import (
	"embed"
	"io/fs"
	"path/filepath"

	"juarvis/pkg/utils"
)

//go:embed all:data
var embeddedAssets embed.FS

// GetEmbeddedFS retorna el sistema de archivos de las reglas nativas compiladas dentro del binario
func GetEmbeddedFS() (fs.FS, error) {
	return fs.Sub(embeddedAssets, "data")
}

// CopyEmbeddedToDisk copia un directorio del embed.FS al filesystem
func CopyEmbeddedToDisk(srcPath, destPath string) error {
	return utils.CopyEmbeddedDir(embeddedAssets, filepath.Join("data", srcPath), destPath)
}
