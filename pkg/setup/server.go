package setup

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed ui/*
var uiFS embed.FS

type InstallRequest struct {
	Targets   []string `json:"targets"`
	CSRFToken string   `json:"csrf_token,omitempty"`
}

// generateCSRFToken genera un token aleatorio de 32 bytes hex-encoded
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
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

	csrfToken, err := generateCSRFToken()
	if err != nil {
		return fmt.Errorf("error generando CSRF token: %w", err)
	}

	staticFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		return fmt.Errorf("error cargando frontend embebido: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/api/csrf-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": csrfToken})
	})

	mux.HandleFunc("/api/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "application/json") {
			http.Error(w, "Content-Type debe ser application/json", http.StatusUnsupportedMediaType)
			return
		}

		var req InstallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Error parseando body: %v", err), http.StatusBadRequest)
			return
		}

		// Verify CSRF token
		if req.CSRFToken != "" && req.CSRFToken != csrfToken {
			http.Error(w, "CSRF token invalido", http.StatusForbidden)
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

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		rootPath, _ := root.GetRoot()
		// Reutilizar lógica de 'vibe' simplificada
		snapshotsCount := 0
		out, err := exec.Command("git", "stash", "list").CombinedOutput()
		if err == nil {
			snapshotsCount = strings.Count(string(out), "juarvis-snapshot|")
		}

		resp := map[string]interface{}{
			"root":      rootPath,
			"snapshots": snapshotsCount,
			"time":      time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	output.Info("Levantando Interfaz Gráfica (Web UI) en %s...", url)

	go func() {
		time.Sleep(1 * time.Second)
		openBrowser(url)
	}()

	srv := &http.Server{Addr: "127.0.0.1:" + port, Handler: mux}

	// Graceful shutdown on SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		output.Info("Apagando servidor GUI...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
