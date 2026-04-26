package config

const (
	// JuarFile is the versioned config file in project root (team-shared)
	JuarFile = "juarvis.yaml"

	// JuarDir is the personal directory (gitignored) - memory, sessions, snapshots
	JuarDir = ".juar"

	// JuarvisPluginDir is the manifest directory for each plugin
	JuarvisPluginDir = ".juarvis-plugin"

	// JuarvisDir is the local config directory (IDE preferences, local hooks)
	JuarvisDir = ".juarvis"

	// AgentSkillsDir is the agent SDK skills directory
	AgentSkillsDir = ".agent/skills"

	// HookifyPattern is the pattern for hookify rule files
	HookifyPattern = "hookify.*.local.md"

	// RalphStateFile is the Ralph loop state file
	RalphStateFile = "ralph-loop.local.md"

	// SkillRegistryFile is the skill registry filename
	SkillRegistryFile = "skill-registry.md"

	// MemoryDir is the memory subdirectory within .juar
	MemoryDir = "memory"

	// WatcherPIDFile is the watcher PID file (no dot prefix - visible file)
	WatcherPIDFile = "watcher.pid"
)
