package setup

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"juarvis/pkg/output"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

//go:embed ui/*
var uiFS embed.FS

type InstallRequest struct {
	Targets []string `json:"targets"`
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("plataforma no soportada")
	}
	if err != nil {
		output.Warning("No se pudo abrir el navegador automáticamente. Visita %s", url)
	}
}

// RunServer levanta la Interfaz Gráfica (UI) Web embebida
func RunServer() error {
	port := "8989"
	url := fmt.Sprintf("http://localhost:%s", port)

	staticFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		return fmt.Errorf("error cargando frontend embebido: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/api/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		var req InstallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Error parseando body: %v", err), http.StatusBadRequest)
			return
		}

		if len(req.Targets) == 0 {
			http.Error(w, "Targets vacios", http.StatusBadRequest)
			return
		}

		// Capturar la salida de consola (Stdout) para enviarla al Frontend
		origStdout := os.Stdout
		pipeR, pipeW, _ := os.Pipe()
		os.Stdout = pipeW

		err := RunSetupCore(req.Targets)

		pipeW.Close()
		os.Stdout = origStdout

		var logBuf bytes.Buffer
		io.Copy(&logBuf, pipeR)
		logs := logBuf.String()

		if err != nil {
			resp := map[string]string{"status": "error", "error": err.Error(), "logs": logs}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := map[string]string{"status": "ok", "logs": logs}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	output.Info("Levantando Interfaz Gráfica (Web UI) en %s...", url)

	go func() {
		time.Sleep(1 * time.Second)
		openBrowser(url)
	}()

	err = http.ListenAndServe("127.0.0.1:"+port, mux)
	if err != nil {
		return err
	}
	return nil
}
