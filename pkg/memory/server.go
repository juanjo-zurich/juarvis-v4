package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ServeStdio(rootPath string) error {
	s := server.NewMCPServer(
		"juarvis-memory",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	storage, err := NewStorage(rootPath)
	if err != nil {
		return fmt.Errorf("error inicializando storage: %w", err)
	}

	registerTools(s, storage)

	return server.ServeStdio(s)
}

func registerTools(s *server.MCPServer, storage *Storage) {
	s.AddTool(mcp.NewTool("mem_save",
		mcp.WithDescription("Guardar una observación de memoria persistente"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Título corto y descriptivo")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Contenido estructurado de la observación")),
		mcp.WithString("type", mcp.Description("Categoría: decision, architecture, bugfix, pattern, config, discovery, learning")),
		mcp.WithString("project", mcp.Description("Nombre del proyecto")),
		mcp.WithString("scope", mcp.Description("scope: project o personal")),
		mcp.WithString("topic_key", mcp.Description("Clave de tema estable para upserts")),
		mcp.WithString("session_id", mcp.Description("ID de sesión asociada")),
	), memSaveHandler(storage))

	s.AddTool(mcp.NewTool("mem_search",
		mcp.WithDescription("Buscar observaciones por query full-text"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Query de búsqueda")),
		mcp.WithString("project", mcp.Description("Filtrar por proyecto")),
		mcp.WithString("type", mcp.Description("Filtrar por tipo")),
		mcp.WithString("scope", mcp.Description("Filtrar por scope")),
		mcp.WithNumber("limit", mcp.Description("Máximo de resultados (default 10)")),
	), memSearchHandler(storage))

	s.AddTool(mcp.NewTool("mem_context",
		mcp.WithDescription("Obtener contexto de sesiones recientes"),
		mcp.WithString("project", mcp.Description("Filtrar por proyecto")),
		mcp.WithNumber("limit", mcp.Description("Máximo de sesiones (default 20)")),
	), memContextHandler(storage))

	s.AddTool(mcp.NewTool("mem_session_summary",
		mcp.WithDescription("Guardar resumen de fin de sesión"),
		mcp.WithString("content", mcp.Required(), mcp.Description("Resumen estructurado de la sesión")),
		mcp.WithString("project", mcp.Required(), mcp.Description("Nombre del proyecto")),
		mcp.WithString("session_id", mcp.Description("ID de sesión")),
	), memSessionSummaryHandler(storage))

	s.AddTool(mcp.NewTool("mem_get_observation",
		mcp.WithDescription("Obtener observación completa por ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID de la observación")),
	), memGetObservationHandler(storage))

	s.AddTool(mcp.NewTool("mem_suggest_topic_key",
		mcp.WithDescription("Sugerir una clave de tema estable para una observación"),
		mcp.WithString("type", mcp.Description("Tipo de observación")),
		mcp.WithString("title", mcp.Description("Título de la observación")),
		mcp.WithString("content", mcp.Description("Contenido de la observación")),
	), memSuggestTopicKeyHandler())

	s.AddTool(mcp.NewTool("mem_update",
		mcp.WithDescription("Actualizar campos de una observación existente"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID de la observación")),
		mcp.WithString("title", mcp.Description("Nuevo título")),
		mcp.WithString("content", mcp.Description("Nuevo contenido")),
		mcp.WithString("type", mcp.Description("Nuevo tipo")),
		mcp.WithString("project", mcp.Description("Nuevo proyecto")),
		mcp.WithString("scope", mcp.Description("Nuevo scope")),
		mcp.WithString("topic_key", mcp.Description("Nueva topic_key")),
	), memUpdateHandler(storage))

	s.AddTool(mcp.NewTool("mem_delete",
		mcp.WithDescription("Borrar una observación"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID de la observación")),
		mcp.WithBoolean("hard_delete", mcp.Description("Si true, elimina permanentemente (default false)")),
	), memDeleteHandler(storage))

	s.AddTool(mcp.NewTool("mem_session_start",
		mcp.WithDescription("Registrar inicio de sesión"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID único de sesión")),
		mcp.WithString("project", mcp.Required(), mcp.Description("Nombre del proyecto")),
		mcp.WithString("directory", mcp.Description("Directorio de trabajo")),
	), memSessionStartHandler(storage))

	s.AddTool(mcp.NewTool("mem_session_end",
		mcp.WithDescription("Marcar sesión como completada"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID de sesión")),
		mcp.WithString("summary", mcp.Description("Resumen de lo accomplished")),
	), memSessionEndHandler(storage))
}

func memSaveHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		obsType := req.GetString("type", "manual")
		project := req.GetString("project", "")
		scope := req.GetString("scope", "project")
		topicKey := req.GetString("topic_key", "")
		sessionID := req.GetString("session_id", "")

		obs := &Observation{
			Title:     title,
			Type:      obsType,
			Scope:     scope,
			Project:   project,
			TopicKey:  topicKey,
			Content:   content,
			SessionID: sessionID,
		}

		if err := storage.SaveObservation(obs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error guardando observación: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Observación guardada: %s (ID: %s)", obs.Title, obs.ID)), nil
	}
}

func memSearchHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		project := req.GetString("project", "")
		obsType := req.GetString("type", "")
		scope := req.GetString("scope", "")
		limit := int(req.GetFloat("limit", 10))

		results, err := storage.SearchObservations(query, project, obsType, scope, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error buscando: %v", err)), nil
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("No se encontraron observaciones"), nil
		}

		var sb strings.Builder
		for _, obs := range results {
			sb.WriteString(fmt.Sprintf("## %s (ID: %s)\n", obs.Title, obs.ID))
			sb.WriteString(fmt.Sprintf("- Type: %s | Project: %s | Scope: %s\n", obs.Type, obs.Project, obs.Scope))
			sb.WriteString(fmt.Sprintf("- Created: %s\n", obs.CreatedAt.Format(time.RFC3339)))
			sb.WriteString(fmt.Sprintf("- Content: %s\n\n", obs.Content))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func memContextHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		project := req.GetString("project", "")
		limit := int(req.GetFloat("limit", 20))

		sessions, err := storage.ListSessions(project, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error listando sesiones: %v", err)), nil
		}

		if len(sessions) == 0 {
			return mcp.NewToolResultText("No hay sesiones recientes"), nil
		}

		var sb strings.Builder
		for _, sess := range sessions {
			sb.WriteString(fmt.Sprintf("## Sesión: %s\n", sess.ID))
			sb.WriteString(fmt.Sprintf("- Project: %s | Directory: %s\n", sess.Project, sess.Directory))
			sb.WriteString(fmt.Sprintf("- Started: %s\n", sess.StartedAt.Format(time.RFC3339)))
			if sess.Summary != "" {
				sb.WriteString(fmt.Sprintf("- Summary: %s\n", sess.Summary))
			}
			sb.WriteString("\n")
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func memSessionSummaryHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		project, err := req.RequireString("project")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sessionID := req.GetString("session_id", "")

		sess := &Session{
			ID:      sessionID,
			Project: project,
			Summary: content,
		}

		if err := storage.SaveSession(sess); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error guardando resumen: %v", err)), nil
		}

		return mcp.NewToolResultText("Resumen de sesión guardado"), nil
	}
}

func memGetObservationHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		obs, err := storage.GetObservation(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("observación no encontrada: %s", id)), nil
		}

		data, _ := json.MarshalIndent(obs, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func memSuggestTopicKeyHandler() func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		obsType := req.GetString("type", "manual")
		title := req.GetString("title", "")

		key := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		if key == "" {
			key = fmt.Sprintf("%s/untitled", obsType)
		} else {
			key = fmt.Sprintf("%s/%s", obsType, key)
		}

		return mcp.NewToolResultText(fmt.Sprintf("topic_key sugerida: %s", key)), nil
	}
}

func memUpdateHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		updates := make(map[string]interface{})
		for _, field := range []string{"title", "content", "type", "project", "scope", "topic_key"} {
			if v := req.GetString(field, ""); v != "" {
				updates[field] = v
			}
		}

		if len(updates) == 0 {
			return mcp.NewToolResultError("no se proporcionaron campos para actualizar"), nil
		}

		if err := storage.UpdateObservation(id, updates); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error actualizando: %v", err)), nil
		}

		return mcp.NewToolResultText("Observación actualizada"), nil
	}
}

func memDeleteHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		hard := req.GetBool("hard_delete", false)

		if err := storage.DeleteObservation(id, hard); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error borrando: %v", err)), nil
		}

		return mcp.NewToolResultText("Observación borrada"), nil
	}
}

func memSessionStartHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		project, err := req.RequireString("project")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		directory := req.GetString("directory", "")

		sess := &Session{
			ID:        id,
			Project:   project,
			Directory: directory,
			StartedAt: time.Now(),
		}

		if err := storage.SaveSession(sess); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error guardando sesión: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Sesión iniciada: %s", id)), nil
	}
}

func memSessionEndHandler(storage *Storage) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		summary := req.GetString("summary", "")

		sess, err := storage.GetSession(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("sesión no encontrada: %s", id)), nil
		}

		now := time.Now()
		sess.EndedAt = &now
		sess.Summary = summary

		if err := storage.SaveSession(sess); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error actualizando sesión: %v", err)), nil
		}

		return mcp.NewToolResultText("Sesión completada"), nil
	}
}
