package mcp

import (
	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/version"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps an MCP server with pixshift tool handlers.
type Server struct {
	registry  *codec.Registry
	mcpServer *server.MCPServer
}

// NewServer creates an MCP server with all pixshift tools registered.
func NewServer(registry *codec.Registry) *Server {
	s := &Server{
		registry: registry,
		mcpServer: server.NewMCPServer(
			"pixshift",
			version.Version,
			server.WithToolCapabilities(false),
			server.WithInstructions("Pixshift is an image conversion and analysis toolkit. Use convert_image to convert between formats with optional transforms, get_formats to list supported formats, analyze_image to inspect image metadata, and compare_images to compute structural similarity."),
		),
	}

	s.registerTools()
	return s
}

// Run starts the MCP server on stdio.
func (s *Server) Run() error {
	return server.ServeStdio(s.mcpServer)
}

// registerTools adds all pixshift tools to the MCP server.
func (s *Server) registerTools() {
	s.mcpServer.AddTool(convertImageTool(), s.handleConvertImage())
	s.mcpServer.AddTool(getFormatsTool(), s.handleGetFormats())
	s.mcpServer.AddTool(analyzeImageTool(), s.handleAnalyzeImage())
	s.mcpServer.AddTool(compareImagesTool(), s.handleCompareImages())
}
