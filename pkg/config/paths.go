package config

const (
	// JuarDir es el directorio de memoria/registry del ecosistema
	JuarDir = ".juar"

	// JuarvisPluginDir es el directorio de manifiesto de cada plugin
	JuarvisPluginDir = ".juarvis-plugin"

	// JuarvisDir es el directorio de configuración de Juarvis en el proyecto
	JuarvisDir = ".juarvis"

	// AgentSkillsDir es el directorio de skills del agente SDK
	AgentSkillsDir = ".agent/skills"

	// HookifyPattern es el patrón de archivos de reglas de hookify
	HookifyPattern = "hookify.*.local.md"

	// RalphStateFile es el archivo de estado del bucle de Ralph
	RalphStateFile = "ralph-loop.local.md"

	// SkillRegistryFile es el nombre del archivo de registry de skills
	SkillRegistryFile = "skill-registry.md"

	// MemoryDir es el subdirectorio de memoria dentro de .juar
	MemoryDir = "memory"
)
