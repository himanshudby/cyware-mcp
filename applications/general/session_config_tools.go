package general

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func toolStringArg(args map[string]interface{}, key string) string {
	if args == nil {
		return ""
	}
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func ConfigureCTIXConnectionTool(s *server.MCPServer) {
	tool := mcp.NewTool(
		"configure-ctix-connection",
		mcp.WithDescription("Configure CTIX base_url and credentials for the current MCP session. This enables per-client CTIX access on a hosted MCP server."),
		mcp.WithString("base_url", mcp.Required(), mcp.Description("CTIX base URL, e.g. https://demo.cyware.com/ctix/")),
		mcp.WithString("auth_type", mcp.Description(`Optional. One of "basic", "token", or "openapicreds". Defaults to "openapicreds" when access_id and secret_key are set.`)),
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
		baseURL := toolStringArg(args, "base_url")
		if baseURL == "" {
			return mcp.NewToolResultText("base_url is required"), nil
		}
		if _, err := common.NormalizeDomainURL(baseURL); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Invalid CTIX base_url: %v", err)), err
		}
		app := common.Application{
			BASE_URL: baseURL,
			Auth: common.Auth{
				Type: toolStringArg(args, "auth_type"),
			},
		}

		app.Auth.Username = toolStringArg(args, "username")
		app.Auth.Password = toolStringArg(args, "password")
		app.Auth.Token = toolStringArg(args, "token")
		app.Auth.AccessID = toolStringArg(args, "access_id")
		app.Auth.SecretKey = toolStringArg(args, "secret_key")
		app.Auth.Type = common.NormalizeAuthType(app.Auth)

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
		mcp.WithString("auth_type", mcp.Description(`Optional. One of "basic", "token", or "openapicreds". Defaults to "openapicreds" when access_id and secret_key are set.`)),
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
		baseURL := toolStringArg(args, "base_url")
		if baseURL == "" {
			return mcp.NewToolResultText("base_url is required"), nil
		}
		if _, err := common.NormalizeDomainURL(baseURL); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Invalid CO base_url: %v", err)), err
		}
		app := common.Application{
			BASE_URL: baseURL,
			Auth: common.Auth{
				Type: toolStringArg(args, "auth_type"),
			},
		}

		app.Auth.Username = toolStringArg(args, "username")
		app.Auth.Password = toolStringArg(args, "password")
		app.Auth.Token = toolStringArg(args, "token")
		app.Auth.AccessID = toolStringArg(args, "access_id")
		app.Auth.SecretKey = toolStringArg(args, "secret_key")
		app.Auth.Type = common.NormalizeAuthType(app.Auth)

		common.SetSessionCO(sid, app)
		out, _ := json.Marshal(map[string]any{
			"ok":        true,
			"sessionId": sid,
			"auth_type": app.Auth.Type,
		})
		return mcp.NewToolResultText(fmt.Sprintf("%s", out)), nil
	})
}
