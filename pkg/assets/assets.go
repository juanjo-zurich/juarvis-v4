package assets

import (
	"embed"
	"io/fs"
)

//go:embed data/*
var embeddedAssets embed.FS

// GetEmbeddedFS retorna el sistema de archivos de las reglas nativas compiladas dentro del binario
func GetEmbeddedFS() (fs.FS, error) {
	return fs.Sub(embeddedAssets, "data")
}
