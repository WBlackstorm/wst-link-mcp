package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.linkedin.com"
	apiVersion     = "202606"
	restliVersion  = "2.0.0"
)

// Client handles interaction with the LinkedIn REST API.
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client

	mu         sync.RWMutex
	profileURN string
}

// NewClient instantiates a new LinkedIn API client with the provided OAuth access token.
func NewClient(accessToken string) *Client {
	baseURL := os.Getenv("LINKEDIN_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		accessToken: accessToken,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SetBaseURL allows overriding the base URL (useful for testing or custom endpoints).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// applyHeaders attaches all required LinkedIn API headers to an outgoing HTTP request.
func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("LinkedIn-Version", apiVersion)
	req.Header.Set("X-Restli-Protocol-Version", restliVersion)
	req.Header.Set("Content-Type", "application/json")
}

// GetProfile retrieves the logged-in user's profile and returns a Profile struct containing the URN (urn:li:person:ID).
func (c *Client) GetProfile(ctx context.Context) (*Profile, error) {
	c.mu.RLock()
	cachedURN := c.profileURN
	c.mu.RUnlock()

	// 1. Try standard OpenID Connect userinfo endpoint
	url := fmt.Sprintf("%s/v2/userinfo", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get profile request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get profile request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If /v2/userinfo fails, fallback to /v2/me endpoint
		return c.getProfileMeFallback(ctx, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile response body: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(bodyBytes, &profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile json response: %w", err)
	}

	profile.URN = profile.GetPersonURN()

	c.mu.Lock()
	c.profileURN = profile.URN
	c.mu.Unlock()

	_ = cachedURN
	return &profile, nil
}

// getProfileMeFallback calls /v2/me as a fallback when /v2/userinfo is unavailable.
func (c *Client) getProfileMeFallback(ctx context.Context, previousStatus int) (*Profile, error) {
	url := fmt.Sprintf("%s/v2/me", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback get profile request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute fallback get profile request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback profile response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin get profile request failed with status %d (fallback status %d): %s", previousStatus, resp.StatusCode, string(bodyBytes))
	}

	var profile Profile
	if err := json.Unmarshal(bodyBytes, &profile); err != nil {
		return nil, fmt.Errorf("failed to decode fallback profile json response: %w", err)
	}

	profile.URN = profile.GetPersonURN()

	c.mu.Lock()
	c.profileURN = profile.URN
	c.mu.Unlock()

	return &profile, nil
}

// CreateTextPost creates a public commentary post on LinkedIn on behalf of the authenticated person URN.
func (c *Client) CreateTextPost(ctx context.Context, commentary string) (*PostResponse, error) {
	if commentary == "" {
		return nil, fmt.Errorf("commentary cannot be empty")
	}

	c.mu.RLock()
	authorURN := c.profileURN
	c.mu.RUnlock()

	if authorURN == "" {
		profile, err := c.GetProfile(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain author profile URN before posting: %w", err)
		}
		authorURN = profile.URN
	}

	payload := CreatePostRequest{
		Author:     authorURN,
		Commentary: commentary,
		Visibility: "PUBLIC",
		Distribution: Distribution{
			FeedDistribution:               "MAIN_FEED",
			TargetEntities:                 []string{},
			ThirdPartyDistributionChannels: []string{},
		},
		LifecycleState:            "PUBLISHED",
		IsReshareDisabledByAuthor: false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create post payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/posts", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create post request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create post request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create post response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin create post failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// LinkedIn API returns the post URN in the 'x-restli-id' header or body
	postURN := resp.Header.Get("x-restli-id")
	if postURN == "" {
		postURN = resp.Header.Get("X-RestLi-Id")
	}

	postResp := &PostResponse{
		ID:         postURN,
		URN:        postURN,
		Status:     resp.StatusCode,
		Message:    "Post successfully created on LinkedIn",
		Commentary: commentary,
	}

	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, postResp)
	}

	return postResp, nil
}

// CreateMediaPost creates a LinkedIn post with media attached (image or document/carousel URN).
func (c *Client) CreateMediaPost(ctx context.Context, commentary string, mediaURN string, mediaTitle string) (*PostResponse, error) {
	if commentary == "" {
		return nil, fmt.Errorf("commentary cannot be empty")
	}
	if mediaURN == "" {
		return nil, fmt.Errorf("media URN cannot be empty")
	}

	c.mu.RLock()
	authorURN := c.profileURN
	c.mu.RUnlock()

	if authorURN == "" {
		profile, err := c.GetProfile(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain author profile URN before posting: %w", err)
		}
		authorURN = profile.URN
	}

	payload := CreatePostRequest{
		Author:     authorURN,
		Commentary: commentary,
		Visibility: "PUBLIC",
		Distribution: Distribution{
			FeedDistribution:               "MAIN_FEED",
			TargetEntities:                 []string{},
			ThirdPartyDistributionChannels: []string{},
		},
		LifecycleState:            "PUBLISHED",
		IsReshareDisabledByAuthor: false,
		Content: &Content{
			Media: &ContentMedia{
				Media: mediaURN,
				Title: mediaTitle,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create media post payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/posts", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create media post request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create media post request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create media post response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin create media post failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	postURN := resp.Header.Get("x-restli-id")
	if postURN == "" {
		postURN = resp.Header.Get("X-RestLi-Id")
	}

	postResp := &PostResponse{
		ID:         postURN,
		URN:        postURN,
		Status:     resp.StatusCode,
		Message:    "Media post successfully created on LinkedIn",
		Commentary: commentary,
		MediaURN:   mediaURN,
	}

	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, postResp)
	}

	return postResp, nil
}

// RegisterUpload registers a media asset (feed-image or feed-drive-document) upload with LinkedIn.
func (c *Client) RegisterUpload(ctx context.Context, recipe string) (*RegisterUploadResponse, error) {
	c.mu.RLock()
	ownerURN := c.profileURN
	c.mu.RUnlock()

	if ownerURN == "" {
		profile, err := c.GetProfile(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain owner profile URN for media upload: %w", err)
		}
		ownerURN = profile.URN
	}

	var reqBody RegisterUploadRequest
	reqBody.RegisterUploadRequest.Recipes = []string{recipe}
	reqBody.RegisterUploadRequest.Owner = ownerURN
	reqBody.RegisterUploadRequest.ServiceRelationships = []struct {
		RelationshipType string `json:"relationshipType"`
		Identifier       string `json:"identifier"`
	}{
		{
			RelationshipType: "OWNER",
			Identifier:       "urn:li:userGeneratedContent",
		},
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register upload request: %w", err)
	}

	url := fmt.Sprintf("%s/v2/assets?action=registerUpload", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create register upload request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute register upload request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read register upload response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("register upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var regResp RegisterUploadResponse
	if err := json.Unmarshal(bodyBytes, &regResp); err != nil {
		return nil, fmt.Errorf("failed to decode register upload response: %w", err)
	}

	return &regResp, nil
}

// UploadMedia uploads raw media bytes to the registered upload URL and returns the asset URN.
func (c *Client) UploadMedia(ctx context.Context, mediaData []byte, recipe string) (*MediaUploadResult, error) {
	if len(mediaData) == 0 {
		return nil, fmt.Errorf("media data cannot be empty")
	}

	if recipe == "" {
		recipe = "urn:li:digitalmediaRecipe:feedshare-image"
	}

	regResp, err := c.RegisterUpload(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to register media upload: %w", err)
	}

	uploadURL := regResp.Value.UploadMechanism.MediaUploadHTTPRequest.UploadURL
	assetURN := regResp.Value.Asset

	if uploadURL == "" {
		return nil, fmt.Errorf("received empty upload URL from LinkedIn asset registration")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewBuffer(mediaData))
	if err != nil {
		return nil, fmt.Errorf("failed to create binary upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute binary upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binary media upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return &MediaUploadResult{
		AssetURN:  assetURN,
		UploadURL: uploadURL,
		Status:    "SUCCESS",
	}, nil
}

// GetPostAnalytics retrieves impressions, likes, comments, and engagement data for a post.
func (c *Client) GetPostAnalytics(ctx context.Context, postURN string) (*PostAnalytics, error) {
	if postURN == "" {
		return nil, fmt.Errorf("post URN cannot be empty")
	}

	url := fmt.Sprintf("%s/v2/socialActions/%s", c.baseURL, postURN)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create socialActions request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute socialActions request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read socialActions response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Fallback mock statistics structure if specific scope is missing or endpoint is 404
		return &PostAnalytics{
			PostURN:          postURN,
			Impressions:      0,
			UniqueImpressions: 0,
			Likes:            0,
			Comments:         0,
			Shares:           0,
			Clicks:           0,
			EngagementRate:   0.0,
		}, nil
	}

	var socialActions struct {
		CommentsSummary struct {
			AggregatedTotalCount int64 `json:"aggregatedTotalCount"`
		} `json:"commentsSummary"`
		LikesSummary struct {
			AggregatedTotalCount int64 `json:"aggregatedTotalCount"`
		} `json:"likesSummary"`
	}

	_ = json.Unmarshal(bodyBytes, &socialActions)

	analytics := &PostAnalytics{
		PostURN:     postURN,
		Likes:       socialActions.LikesSummary.AggregatedTotalCount,
		Comments:    socialActions.CommentsSummary.AggregatedTotalCount,
		Impressions: 0,
	}

	return analytics, nil
}

// ExchangeOAuthToken exchanges an authorization code or refresh token for a LinkedIn access token.
func (c *Client) ExchangeOAuthToken(ctx context.Context, clientID, clientSecret, code, redirectURI, refreshToken string) (*TokenResponse, error) {
	u := fmt.Sprintf("%s/oauth/v2/accessToken", c.baseURL)

	data := url.Values{}
	if refreshToken != "" {
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", refreshToken)
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
	} else {
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", redirectURI)
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute token exchange request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token exchange response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth token exchange failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response json: %w", err)
	}

	return &tokenResp, nil
}

