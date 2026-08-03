package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/weverton/wst-link-mcp/pkg/linkedin"
	"github.com/weverton/wst-link-mcp/pkg/mcp"
)

func TestIntegrationFullFlow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/userinfo":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(linkedin.Profile{
				Sub:  "user_integration_123",
				Name: "Integration Test User",
			})
		case "/v2/assets":
			w.WriteHeader(http.StatusOK)
			var resp linkedin.RegisterUploadResponse
			resp.Value.Asset = "urn:li:digitalmediaAsset:asset_integration_456"
			resp.Value.UploadMechanism.MediaUploadHTTPRequest.UploadURL = "http://" + r.Host + "/binary-upload"
			json.NewEncoder(w).Encode(resp)
		case "/binary-upload":
			w.WriteHeader(http.StatusCreated)
		case "/rest/posts":
			w.Header().Set("x-restli-id", "urn:li:share:integration_post_789")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(linkedin.PostResponse{
				URN: "urn:li:share:integration_post_789",
			})
		case "/v2/socialActions/urn:li:share:integration_post_789":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"commentsSummary":{"aggregatedTotalCount":5},"likesSummary":{"aggregatedTotalCount":20}}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	os.Setenv("LINKEDIN_BASE_URL", mockServer.URL)
	defer os.Unsetenv("LINKEDIN_BASE_URL")

	client := linkedin.NewClient("test-token")
	client.SetBaseURL(mockServer.URL)

	ctx := context.Background()

	// 1. Get profile
	profile, err := client.GetProfile(ctx)
	if err != nil {
		t.Fatalf("Integration: GetProfile failed: %v", err)
	}
	if profile.URN != "urn:li:person:user_integration_123" {
		t.Errorf("Integration: Unexpected profile URN: %s", profile.URN)
	}

	// 2. Upload media
	mediaResult, err := client.UploadMedia(ctx, []byte("test image data"), "urn:li:digitalmediaRecipe:feedshare-image")
	if err != nil {
		t.Fatalf("Integration: UploadMedia failed: %v", err)
	}
	if mediaResult.AssetURN != "urn:li:digitalmediaAsset:asset_integration_456" {
		t.Errorf("Integration: Unexpected asset URN: %s", mediaResult.AssetURN)
	}

	// 3. Create media post
	postResp, err := client.CreateMediaPost(ctx, "Integration Post Commentary", mediaResult.AssetURN, "Image Title")
	if err != nil {
		t.Fatalf("Integration: CreateMediaPost failed: %v", err)
	}
	if postResp.URN != "urn:li:share:integration_post_789" {
		t.Errorf("Integration: Unexpected post URN: %s", postResp.URN)
	}

	// 4. Get post analytics
	analytics, err := client.GetPostAnalytics(ctx, postResp.URN)
	if err != nil {
		t.Fatalf("Integration: GetPostAnalytics failed: %v", err)
	}
	if analytics.Likes != 20 || analytics.Comments != 5 {
		t.Errorf("Integration: Unexpected analytics values. Likes: %d, Comments: %d", analytics.Likes, analytics.Comments)
	}

	// 5. Verify MCP server initialization with new tools
	mcpServer, err := mcp.NewServer("test-token")
	if err != nil {
		t.Fatalf("Integration: NewServer failed: %v", err)
	}
	if mcpServer == nil {
		t.Fatal("Integration: MCP Server instance is nil")
	}
}
