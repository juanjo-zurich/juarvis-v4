package snapshot

import (
	"bytes"
	"fmt"
	"juarvis/pkg/output"
	"os/exec"
	"strings"
)

// CreateSnapshot utiliza git stash de forma transparente
func CreateSnapshot(name string) error {
	output.Info("Creando snapshot de seguridad interno: %s", name)

	statusCmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	statusCmd.Stdout = &out
	if err := statusCmd.Run(); err == nil && len(strings.TrimSpace(out.String())) == 0 {
		output.Warning("No hay cambios pendientes en el árbol de trabajo. Snapshot omitido.")
		return nil
	}

	cmd := exec.Command("git", "stash", "push", "-u", "-m", fmt.Sprintf("juarvis-snapshot|%s", name))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fallo al crear snapshot git stash: %s", string(out))
	}

	if err := exec.Command("git", "stash", "pop", "stash@{0}").Run(); err != nil {
		return fmt.Errorf("error aplicando snapshot: %w", err)
	}

	output.Success("Snapshot de seguridad guardado satisfactoriamente.")
	return nil
}

// RestoreLatestSnapshot recupera el último snapshot creado por juarvis
func RestoreLatestSnapshot() error {
	output.Info("Buscando el último snapshot de Juarvis...")

	// Listamos stashes que contengan juarvis-snapshot
	cmdList := exec.Command("git", "stash", "list")
	out, err := cmdList.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error listando snapshots: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	stashTarget := ""
	for _, line := range lines {
		if strings.Contains(line, "juarvis-snapshot|") {
			// Extract stash@{X}
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				stashTarget = parts[0]
				break
			}
		}
	}

	if stashTarget == "" {
		return fmt.Errorf("no se encontraron snapshots de juarvis en el historial de git stash")
	}

	output.Info("Restaurando %s manteniendo su entrada original...", stashTarget)

	cmdRestore := exec.Command("git", "stash", "apply", stashTarget)
	if outRest, err := cmdRestore.CombinedOutput(); err != nil {
		output.Warning("Atención: Hubo conflictos al restaurar el snapshot.\nDetalles:\n%s\n", string(outRest))
		output.Info("Por favor, resuelve los conflictos en tu editor y completa el proceso.")
		return fmt.Errorf("conflictos de merge detectados durante la restauración")
	}

	output.Success("Código revertido al estado del snapshot de forma segura.")
	return nil
}

// PruneSnapshots elimina stashes de juarvis.
// Si all es true, elimina todos. Si olderThan > 0, elimina solo los más antiguos.
func PruneSnapshots(all bool) (int, error) {
	cmdList := exec.Command("git", "stash", "list")
	output, err := cmdList.CombinedOutput()
	if err != nil {
		// No hay stashes o git no disponible
		return 0, nil
	}

	lines := strings.Split(string(output), "\n")
	pruned := 0

	// Collect juarvis stash refs first, then drop in reverse order to avoid index shifting
	var juarvisRefs []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "juarvis-snapshot|") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 0 {
			continue
		}
		juarvisRefs = append(juarvisRefs, strings.TrimSpace(parts[0]))
	}

	// Drop in reverse order to prevent index shifting
	for i := len(juarvisRefs) - 1; i >= 0; i-- {
		cmdDrop := exec.Command("git", "stash", "drop", juarvisRefs[i])
		if err := cmdDrop.Run(); err == nil {
			pruned++
		}
	}

	return pruned, nil
}
