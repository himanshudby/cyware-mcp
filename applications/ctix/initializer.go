package ctix

import (
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/server"
)

// InitClient initializes the default CTIX client using the application
// configuration from the main config. Hosted deployments can override CTIX
// connection details per MCP session using `configure-ctix-connection`.
func InitClient(cfg *common.Config) {
	ctixDefaultConfig = cfg.Applications["ctix"]
	if ctixDefaultConfig.BASE_URL == "" {
		return
	}
	ctixDefaultClient = buildCTIXClient(ctixDefaultConfig)
	ctixDefaultConfig = normalizeCTIXConfig(ctixDefaultConfig)
}

// Initialize sets up all CTIX tools and the API client within the MCP server context.
//
// It first initializes the client configuration, then registers all CTIX-specific tools
// to the server instance.
func Initialize(cfg *common.Config, s *server.MCPServer) {

	InitClient(cfg)
	InitTools(s)
}

// InitTools performs login and registers all CTIX-specific tools with the MCP server.
//
// It ensures a valid session token is set via Login(), and then exposes all relevant
// CTIX tools such as user info, threat data actions, tagging, and CQL search capabilities.
func InitTools(s *server.MCPServer) {
	// Login is handled by the per-session client configuration (or default client).
	GetLoggedInUserDetailsTool(s)

	// cql and search
	CQLCTIXSearchGrammarTool(s)
	GetCQLQuerySearchResultTool(s)
	GetThreatDataObjectDetailsTool(s)
	GetThreatDataObjectRelationsTool(s)
	GetAvailableRelationTypeListingTool(s)

	// bulk action threat data
	ThreatDataListBulkActionTools(s)

	// tag management
	CreateTaginCTIXTool(s)
	GetCTIXTagListingTool(s)

	// enrichment
	GetEnrichmenToolsListTool(s)
	GetEnrichmentToolsDetailsTool(s)
	GetEnrichmentToolActionConfigsTool(s)
	GetAllEnrichmentToolSupportedForThreatDataObjectTool(s)
	EnrichThreatDataObjectTool(s)

	// intel creation
	CreateQuickAddIntelTool(s)

}
