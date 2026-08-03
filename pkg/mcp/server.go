package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
		mcp.WithDescription("Publishes a text or media post to LinkedIn on behalf of the authenticated user"),
		mcp.WithString("commentary", mcp.Required(), mcp.Description("The text content of the post to publish")),
		mcp.WithString("media_urn", mcp.Description("Optional LinkedIn asset URN for image/document attachment")),
		mcp.WithString("media_title", mcp.Description("Optional title for attached media")),
	)
	s.AddTool(publishTool, handlePublishPost(client))

	// Tool 2: get_linkedin_profile
	getProfileTool := mcp.NewTool(
		"get_linkedin_profile",
		mcp.WithDescription("Retrieves the authenticated user's LinkedIn profile data and person URN"),
	)
	s.AddTool(getProfileTool, handleGetProfile(client))

	// Tool 3: upload_linkedin_media
	uploadMediaTool := mcp.NewTool(
		"upload_linkedin_media",
		mcp.WithDescription("Uploads an image or document file to LinkedIn and returns the asset URN"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Local file path of the image or PDF/document to upload")),
		mcp.WithString("recipe", mcp.Description("Upload recipe (urn:li:digitalmediaRecipe:feedshare-image or urn:li:digitalmediaRecipe:feedshare-document)")),
	)
	s.AddTool(uploadMediaTool, handleUploadMedia(client))

	// Tool 4: get_linkedin_post_analytics
	getAnalyticsTool := mcp.NewTool(
		"get_linkedin_post_analytics",
		mcp.WithDescription("Retrieves analytics and social metrics (likes, comments, etc.) for a specific post"),
		mcp.WithString("post_urn", mcp.Required(), mcp.Description("The URN or ID of the LinkedIn post")),
	)
	s.AddTool(getAnalyticsTool, handleGetPostAnalytics(client))

	return s, nil
}

// NewSSEServer creates an SSE MCP server listening for SSE connections.
func NewSSEServer(s *server.MCPServer) *server.SSEServer {
	return server.NewSSEServer(s)
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

		mediaURN, _ := args["media_urn"].(string)
		mediaTitle, _ := args["media_title"].(string)

		var resp *linkedin.PostResponse
		var err error

		if mediaURN != "" {
			resp, err = client.CreateMediaPost(ctx, commentary, mediaURN, mediaTitle)
		} else {
			resp, err = client.CreateTextPost(ctx, commentary)
		}

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

// handleUploadMedia handles upload of local image or document files to LinkedIn.
func handleUploadMedia(client *linkedin.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		filePath, ok := args["file_path"].(string)
		if !ok || filePath == "" {
			return mcp.NewToolResultError("parameter 'file_path' must be a non-empty string"), nil
		}

		recipe, _ := args["recipe"].(string)
		if recipe == "" {
			recipe = "urn:li:digitalmediaRecipe:feedshare-image"
		}

		mediaBytes, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read file '%s': %v", filePath, err)), nil
		}

		result, err := client.UploadMedia(ctx, mediaBytes, recipe)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to upload media to LinkedIn: %v", err)), nil
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Media uploaded successfully. Asset URN: %s", result.AssetURN)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

// handleGetPostAnalytics handles fetching analytics for a post.
func handleGetPostAnalytics(client *linkedin.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		postURN, ok := args["post_urn"].(string)
		if !ok || postURN == "" {
			return mcp.NewToolResultError("parameter 'post_urn' must be a non-empty string"), nil
		}

		analytics, err := client.GetPostAnalytics(ctx, postURN)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch post analytics: %v", err)), nil
		}

		analyticsJSON, err := json.MarshalIndent(analytics, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal analytics data: %v", err)), nil
		}

		return mcp.NewToolResultText(string(analyticsJSON)), nil
	}
}

