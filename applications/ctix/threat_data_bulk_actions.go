package ctix

import (
	"context"
	"fmt"

	"github.com/cyware-labs/cyware-mcpserver/applications/ctix/helpers"
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type BulkActionResponse struct {
	Message string `json:"message"`
}

func ThreatDataListBulkAction(ctx context.Context, endpoint string, payload any) (*common.APIResponse, error) {
	bulkResp := BulkActionResponse{}
	client, _, ok := CTIXClientFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("CTIX is not configured for this session")
	}
	resp, err := client.MakeRequest("POST", endpoint, nil, &bulkResp, payload, nil)
	return &common.APIResponse{
		FilteredReponse: common.JsonifyResponse(bulkResp),
		RawResponse:     resp,
	}, err
}

// This function uses an action map and registers tools for all the bulk actions of threat data
func ThreatDataListBulkActionTools(s *server.MCPServer) {
	mp := helpers.GetThreatDataBulkActionsMapping()
	for _, v := range mp {
		tool := mcp.NewToolWithRawSchema(v["tool_name"], v["tool_description"], []byte(v["schema"]))

		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			resp, err := ThreatDataListBulkAction(ctx, v["endpoint"], request.Params.Arguments)
			return common.MCPToolResponse(resp, []int{200}, err)
		})
	}
}
