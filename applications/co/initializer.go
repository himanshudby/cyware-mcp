package co

import (
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/server"
)

// InitClient initializes the default CO client using the application
// configuration from the main config. Hosted deployments can override CO
// connection details per MCP session using `configure-co-connection`.
func InitClient(cfg *common.Config) {
	coDefaultConfig = cfg.Applications["co"]
	if coDefaultConfig.BASE_URL == "" {
		return
	}
	coDefaultClient = buildCOClient(coDefaultConfig)
	coDefaultConfig = normalizeCOConfig(coDefaultConfig)
}

// Initialize sets up all CO tools and the API client within the MCP server context.
//
// It first initializes the client configuration, then registers all CO-specific tools
// to the server instance.
func Initialize(cfg *common.Config, s *server.MCPServer) {
	InitClient(cfg)
	InitTools(s)
}

// InitTools performs login and registers all CO-specific tools with the MCP server.
// It ensures a valid session token is set via Login(), and then exposes all relevant
// CO tools such as getting playbook list, executing the playbook, executing actions capabilities.
func InitTools(s *server.MCPServer) {
	// Login and workspace selection are handled per-session (or by default config).
	GetPlayBookListTool(s)
	GetPlaybookDetailsTool(s)
	ExecutePlaybookTool(s)
	GetCOAppsListingTool(s)
	GetCOAppDetailsTool(s)
	COAppActionsListingTool(s)
	GetCOAppActionDetailsTool(s)
	GetConfiguredInstancesOfCOAppTool(s)
	ExecuteActionOfCOAppTool(s)
}
