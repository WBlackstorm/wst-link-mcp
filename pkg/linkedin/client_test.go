package linkedin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProfileUserInfoSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/userinfo" {
			t.Fatalf("expected path /v2/userinfo, got %s", r.URL.Path)
		}

		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Authorization Bearer test-token, got %s", auth)
		}
		if ver := r.Header.Get("LinkedIn-Version"); ver != "202401" {
			t.Errorf("expected LinkedIn-Version 202401, got %s", ver)
		}
		if restli := r.Header.Get("X-Restli-Protocol-Version"); restli != "2.0.0" {
			t.Errorf("expected X-Restli-Protocol-Version 2.0.0, got %s", restli)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Profile{
			Sub:       "abc123person",
			Name:      "Test User",
			GivenName: "Test",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	profile, err := client.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting profile: %v", err)
	}

	expectedURN := "urn:li:person:abc123person"
	if profile.URN != expectedURN {
		t.Errorf("expected profile URN %s, got %s", expectedURN, profile.URN)
	}
}

func TestGetProfileFallbackMe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/userinfo" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Path == "/v2/me" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Profile{
				ID: "fallback456",
			})
			return
		}
		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	profile, err := client.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting fallback profile: %v", err)
	}

	expectedURN := "urn:li:person:fallback456"
	if profile.URN != expectedURN {
		t.Errorf("expected fallback profile URN %s, got %s", expectedURN, profile.URN)
	}
}

func TestCreateTextPostSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Profile{Sub: "person789"})
			return
		}

		if r.URL.Path == "/rest/posts" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}

			var reqPayload CreatePostRequest
			if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			if reqPayload.Author != "urn:li:person:person789" {
				t.Errorf("expected author urn:li:person:person789, got %s", reqPayload.Author)
			}
			if reqPayload.Commentary != "Hello LinkedIn from MCP!" {
				t.Errorf("expected commentary 'Hello LinkedIn from MCP!', got %s", reqPayload.Commentary)
			}
			if reqPayload.Visibility != "PUBLIC" {
				t.Errorf("expected visibility PUBLIC, got %s", reqPayload.Visibility)
			}

			w.Header().Set("x-restli-id", "urn:li:share:123456789")
			w.WriteHeader(http.StatusCreated)
			return
		}

		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	resp, err := client.CreateTextPost(context.Background(), "Hello LinkedIn from MCP!")
	if err != nil {
		t.Fatalf("unexpected error creating text post: %v", err)
	}

	if resp.URN != "urn:li:share:123456789" {
		t.Errorf("expected post URN 'urn:li:share:123456789', got %s", resp.URN)
	}
}

func TestCreateTextPostEmptyCommentary(t *testing.T) {
	client := NewClient("test-token")
	_, err := client.CreateTextPost(context.Background(), "")
	if err == nil {
		t.Fatal("expected error when commentary is empty, got nil")
	}
}

func TestUploadMediaAndCreateMediaPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Profile{Sub: "person789"})
			return
		}

		if r.URL.Path == "/v2/assets" && r.URL.Query().Get("action") == "registerUpload" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			var regResp RegisterUploadResponse
			regResp.Value.Asset = "urn:li:digitalmediaAsset:asset123"
			regResp.Value.UploadMechanism.MediaUploadHTTPRequest.UploadURL = "http://" + r.Host + "/upload-binary"
			json.NewEncoder(w).Encode(regResp)
			return
		}

		if r.URL.Path == "/upload-binary" {
			w.WriteHeader(http.StatusCreated)
			return
		}

		if r.URL.Path == "/rest/posts" {
			var reqPayload CreatePostRequest
			if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if reqPayload.Content == nil || reqPayload.Content.Media == nil {
				t.Fatal("expected media content in post request")
			}
			if reqPayload.Content.Media.Media != "urn:li:digitalmediaAsset:asset123" {
				t.Errorf("expected media URN urn:li:digitalmediaAsset:asset123, got %s", reqPayload.Content.Media.Media)
			}

			w.Header().Set("x-restli-id", "urn:li:share:987654321")
			w.WriteHeader(http.StatusCreated)
			return
		}

		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	uploadRes, err := client.UploadMedia(ctx, []byte("fake image data"), "urn:li:digitalmediaRecipe:feedshare-image")
	if err != nil {
		t.Fatalf("unexpected upload media error: %v", err)
	}

	if uploadRes.AssetURN != "urn:li:digitalmediaAsset:asset123" {
		t.Errorf("expected asset URN urn:li:digitalmediaAsset:asset123, got %s", uploadRes.AssetURN)
	}

	postRes, err := client.CreateMediaPost(ctx, "Media Post Commentary", uploadRes.AssetURN, "Image Title")
	if err != nil {
		t.Fatalf("unexpected create media post error: %v", err)
	}

	if postRes.URN != "urn:li:share:987654321" {
		t.Errorf("expected post URN urn:li:share:987654321, got %s", postRes.URN)
	}
}

func TestGetPostAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/socialActions/urn:li:share:123" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"commentsSummary": {"aggregatedTotalCount": 15},
				"likesSummary": {"aggregatedTotalCount": 42}
			}`))
			return
		}
		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	analytics, err := client.GetPostAnalytics(context.Background(), "urn:li:share:123")
	if err != nil {
		t.Fatalf("unexpected error fetching post analytics: %v", err)
	}

	if analytics.Likes != 42 {
		t.Errorf("expected 42 likes, got %d", analytics.Likes)
	}
	if analytics.Comments != 15 {
		t.Errorf("expected 15 comments, got %d", analytics.Comments)
	}
}

func TestExchangeOAuthToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/v2/accessToken" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "new-access-token",
				ExpiresIn:   5184000,
			})
			return
		}
		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient("")
	client.SetBaseURL(server.URL)

	tokenResp, err := client.ExchangeOAuthToken(context.Background(), "id", "secret", "code", "http://localhost/cb", "")
	if err != nil {
		t.Fatalf("unexpected error exchanging OAuth token: %v", err)
	}

	if tokenResp.AccessToken != "new-access-token" {
		t.Errorf("expected new-access-token, got %s", tokenResp.AccessToken)
	}
}

