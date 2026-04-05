package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/pm"
	"os"

	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Gestión autónoma de Agent Skills (.agent/skills/)",
}

var createSkillCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Crea el andamiaje oficial para una nueva Agent Skill en el proyecto",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]
		output.Info("Creando andamiaje para la skill '%s'...", skillName)
		
		if err := pm.CreateSkill(skillName); err != nil {
			output.Error("Fallo al crear la skill: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	skillCmd.AddCommand(createSkillCmd)
	rootCmd.AddCommand(skillCmd)
}
