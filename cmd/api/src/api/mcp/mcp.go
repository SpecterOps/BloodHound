package mcp

import (
	"log/slog"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/specterops/bloodhound/cmd/api/src/database"
)

// NewHandler creates an http.Handler that serves the MCP SSE endpoint.
// It mounts at /api/v2/mcp/ with SSE at /api/v2/mcp/sse and messages at /api/v2/mcp/message.
// Auth is via the X-BH-Token header (format: token_id:token_key).
func NewHandler(loopbackURL string, db database.Database) http.Handler {
	mcpServer := server.NewMCPServer(
		"bloodhound-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerAllTools(mcpServer, loopbackURL)

	sseServer := server.NewSSEServer(mcpServer,
		server.WithStaticBasePath("/api/v2/mcp"),
		server.WithKeepAlive(true),
	)

	mux := http.NewServeMux()
	mux.Handle("/api/v2/mcp/sse", sseServer.SSEHandler())
	mux.Handle("/api/v2/mcp/message", sseServer.MessageHandler())

	slog.Info("MCP server registered at /api/v2/mcp/sse", "loopback", loopbackURL)

	return authMiddleware(db)(mux)
}
