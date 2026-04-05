package utils

import "strings"

// ExtractFrontmatterBlock extrae el bloque de frontmatter delimitado por --- de un contenido markdown.
// Retorna el frontmatter, el body, y si se encontró un bloque válido.
func ExtractFrontmatterBlock(content string) (frontmatter, body string, found bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return "", content, false
	}

	endIdx := strings.Index(content[3:], "\n---")
	if endIdx == -1 {
		return "", content, false
	}
	endIdx += 3

	return content[3:endIdx], strings.TrimSpace(content[endIdx+4:]), true
}
