package general

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ConfigureCTIXConnectionTool(s *server.MCPServer) {
	tool := mcp.NewTool(
		"configure-ctix-connection",
		mcp.WithDescription("Configure CTIX base_url and credentials for the current MCP session. This enables per-client CTIX access on a hosted MCP server."),
		mcp.WithString("base_url", mcp.Required(), mcp.Description("CTIX base URL, e.g. https://demo.cyware.com/ctix/")),
		mcp.WithString("auth_type", mcp.Required(), mcp.Description(`Auth type: "basic", "token", or "openapicreds"`)),
		mcp.WithString("username", mcp.Description("Required for auth_type=basic")),
		mcp.WithString("password", mcp.Description("Required for auth_type=basic")),
		mcp.WithString("token", mcp.Description("Required for auth_type=token")),
		mcp.WithString("access_id", mcp.Description("Required for auth_type=openapicreds")),
		mcp.WithString("secret_key", mcp.Description("Required for auth_type=openapicreds")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sid, ok := common.SessionIDFromContext(ctx)
		if !ok {
			return mcp.NewToolResultText("No active session found; cannot set CTIX connection."), nil
		}

		args := request.Params.Arguments
		app := common.Application{
			BASE_URL: args["base_url"].(string),
			Auth: common.Auth{
				Type: args["auth_type"].(string),
			},
		}

		if v, ok := args["username"].(string); ok {
			app.Auth.Username = v
		}
		if v, ok := args["password"].(string); ok {
			app.Auth.Password = v
		}
		if v, ok := args["token"].(string); ok {
			app.Auth.Token = v
		}
		if v, ok := args["access_id"].(string); ok {
			app.Auth.AccessID = v
		}
		if v, ok := args["secret_key"].(string); ok {
			app.Auth.SecretKey = v
		}

		common.SetSessionCTIX(sid, app)
		out, _ := json.Marshal(map[string]any{
			"ok":        true,
			"sessionId": sid,
			"auth_type": app.Auth.Type,
		})
		return mcp.NewToolResultText(fmt.Sprintf("%s", out)), nil
	})
}

func ConfigureCOConnectionTool(s *server.MCPServer) {
	tool := mcp.NewTool(
		"configure-co-connection",
		mcp.WithDescription("Configure CO base_url and credentials for the current MCP session. This enables per-client CO access on a hosted MCP server."),
		mcp.WithString("base_url", mcp.Required(), mcp.Description("CO base URL, e.g. https://demo.cyware.com/soar/")),
		mcp.WithString("auth_type", mcp.Required(), mcp.Description(`Auth type: "basic", "token", or "openapicreds"`)),
		mcp.WithString("username", mcp.Description("Required for auth_type=basic")),
		mcp.WithString("password", mcp.Description("Required for auth_type=basic")),
		mcp.WithString("token", mcp.Description("Required for auth_type=token")),
		mcp.WithString("access_id", mcp.Description("Required for auth_type=openapicreds")),
		mcp.WithString("secret_key", mcp.Description("Required for auth_type=openapicreds")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sid, ok := common.SessionIDFromContext(ctx)
		if !ok {
			return mcp.NewToolResultText("No active session found; cannot set CO connection."), nil
		}

		args := request.Params.Arguments
		app := common.Application{
			BASE_URL: args["base_url"].(string),
			Auth: common.Auth{
				Type: args["auth_type"].(string),
			},
		}

		if v, ok := args["username"].(string); ok {
			app.Auth.Username = v
		}
		if v, ok := args["password"].(string); ok {
			app.Auth.Password = v
		}
		if v, ok := args["token"].(string); ok {
			app.Auth.Token = v
		}
		if v, ok := args["access_id"].(string); ok {
			app.Auth.AccessID = v
		}
		if v, ok := args["secret_key"].(string); ok {
			app.Auth.SecretKey = v
		}

		common.SetSessionCO(sid, app)
		out, _ := json.Marshal(map[string]any{
			"ok":        true,
			"sessionId": sid,
			"auth_type": app.Auth.Type,
		})
		return mcp.NewToolResultText(fmt.Sprintf("%s", out)), nil
	})
}

