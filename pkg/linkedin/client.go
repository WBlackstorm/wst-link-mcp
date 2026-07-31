package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.linkedin.com"
	apiVersion     = "202401"
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
	return &Client{
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
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
