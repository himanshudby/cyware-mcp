package common

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"resty.dev/v3"
)

type APIResponse struct {
	RawResponse     *resty.Response
	FilteredReponse any
}

func JsonifyResponse(obj any) any {
	resp, _ := json.Marshal(obj)
	return string(resp)
}

func MCPToolResponse(resp *APIResponse, expected_status_code []int, err error) (*mcp.CallToolResult, error) {
	// Be defensive: when an upstream call fails, resp and/or resp.RawResponse can be nil.
	if err != nil {
		status := "unknown"
		body := ""
		if resp != nil && resp.RawResponse != nil {
			status = fmt.Sprintf("%d", resp.RawResponse.StatusCode())
			body = resp.RawResponse.String()
		}
		return mcp.NewToolResultText(
			fmt.Sprintf("An error occurred (status=%s). Response: %s", status, body),
		), err
	}

	if resp == nil || resp.RawResponse == nil {
		return mcp.NewToolResultText("An error occurred: empty upstream response"), fmt.Errorf("empty upstream response")
	}

	if !ContainsStatusCode(expected_status_code, resp.RawResponse.StatusCode()) {
		return mcp.NewToolResultText(
			fmt.Sprintf(
				"An error occurred, Server responded with status code %v and response %v",
				resp.RawResponse.StatusCode(),
				resp.RawResponse.String(),
			),
		), fmt.Errorf("unexpected status code %d", resp.RawResponse.StatusCode())
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", resp.FilteredReponse)), nil
}

func ContainsStatusCode(codes []int, statusCode int) bool {
	for _, code := range codes {
		if code == statusCode {
			return true
		}
	}
	return false
}
