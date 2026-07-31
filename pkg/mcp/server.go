package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weverton/wst-link-mcp/pkg/linkedin"
)

// NewServer initializes and returns a configured MCP server with registered LinkedIn tools.
func NewServer(accessToken string) (*server.MCPServer, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token cannot be empty")
	}

	client := linkedin.NewClient(accessToken)

	s := server.NewMCPServer(
		"wst-link-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool 1: publish_linkedin_post
	publishTool := mcp.NewTool(
		"publish_linkedin_post",
		mcp.WithDescription("Publishes a text post to LinkedIn on behalf of the authenticated user"),
		mcp.WithString("commentary", mcp.Required(), mcp.Description("The text content of the post to publish")),
	)
	s.AddTool(publishTool, handlePublishPost(client))

	// Tool 2: get_linkedin_profile
	getProfileTool := mcp.NewTool(
		"get_linkedin_profile",
		mcp.WithDescription("Retrieves the authenticated user's LinkedIn profile data and person URN"),
	)
	s.AddTool(getProfileTool, handleGetProfile(client))

	return s, nil
}

// handlePublishPost processes requests to create a new post on LinkedIn.
func handlePublishPost(client *linkedin.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		commentaryRaw, ok := args["commentary"]
		if !ok {
			return mcp.NewToolResultError("missing required parameter 'commentary'"), nil
		}

		commentary, ok := commentaryRaw.(string)
		if !ok || commentary == "" {
			return mcp.NewToolResultError("parameter 'commentary' must be a non-empty string"), nil
		}

		resp, err := client.CreateTextPost(ctx, commentary)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to publish LinkedIn post: %v", err)), nil
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Post published successfully. URN: %s", resp.URN)), nil
		}

		return mcp.NewToolResultText(string(respJSON)), nil
	}
}

// handleGetProfile processes requests to fetch the user's LinkedIn profile.
func handleGetProfile(client *linkedin.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		profile, err := client.GetProfile(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch LinkedIn profile: %v", err)), nil
		}

		profileJSON, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal profile data: %v", err)), nil
		}

		return mcp.NewToolResultText(string(profileJSON)), nil
	}
}
