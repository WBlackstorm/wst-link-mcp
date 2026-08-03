package linkedin

// Profile represents the user profile returned by LinkedIn userinfo/me endpoint.
type Profile struct {
	ID         string `json:"id,omitempty"`
	Sub        string `json:"sub,omitempty"`
	Name       string `json:"name,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Email      string `json:"email,omitempty"`
	Picture    string `json:"picture,omitempty"`
	URN        string `json:"urn,omitempty"`
}

// GetPersonURN returns the formatted URN (urn:li:person:ID) for the user profile.
func (p *Profile) GetPersonURN() string {
	if p.URN != "" {
		return p.URN
	}
	id := p.Sub
	if id == "" {
		id = p.ID
	}
	return "urn:li:person:" + id
}

// ContentMedia represents media attached to a post (image, document/carousel).
type ContentMedia struct {
	Media string `json:"media"` // URN of registered asset/image/document (e.g., urn:li:image:1234)
	Title string `json:"title,omitempty"`
}

// Content represents post content payload.
type Content struct {
	Media *ContentMedia `json:"media,omitempty"`
}

// Distribution defines the distribution parameters for a LinkedIn post.
type Distribution struct {
	FeedDistribution               string   `json:"feedDistribution"`
	TargetEntities                 []string `json:"targetEntities"`
	ThirdPartyDistributionChannels []string `json:"thirdPartyDistributionChannels"`
}

// CreatePostRequest represents the JSON payload sent to POST https://api.linkedin.com/rest/posts.
type CreatePostRequest struct {
	Author                    string       `json:"author"`
	Commentary                string       `json:"commentary"`
	Visibility                string       `json:"visibility"`
	Distribution              Distribution `json:"distribution"`
	LifecycleState            string       `json:"lifecycleState"`
	IsReshareDisabledByAuthor bool         `json:"isReshareDisabledByAuthor"`
	Content                   *Content     `json:"content,omitempty"`
}

// PostResponse represents the response received after creating a post.
type PostResponse struct {
	ID         string `json:"id,omitempty"`
	URN        string `json:"urn,omitempty"`
	Status     int    `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
	Commentary string `json:"commentary,omitempty"`
	MediaURN   string `json:"mediaUrn,omitempty"`
}

// RegisterUploadRequest represents request to register an image or document upload.
type RegisterUploadRequest struct {
	RegisterUploadRequest struct {
		Recipes []string `json:"recipes"`
		Owner   string   `json:"owner"`
		ServiceRelationships []struct {
			RelationshipType string `json:"relationshipType"`
			Identifier       string `json:"identifier"`
		} `json:"serviceRelationships"`
	} `json:"registerUploadRequest"`
}

// RegisterUploadResponse represents response from asset registration.
type RegisterUploadResponse struct {
	Value struct {
		Asset       string `json:"asset"`
		UploadMechanism struct {
			MediaUploadHTTPRequest struct {
				UploadURL string            `json:"uploadUrl"`
				Headers   map[string]string `json:"headers,omitempty"`
			} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
		} `json:"uploadMechanism"`
	} `json:"value"`
}

// MediaUploadResult represents result of uploading media file.
type MediaUploadResult struct {
	AssetURN  string `json:"assetUrn"`
	UploadURL string `json:"uploadUrl"`
	Status    string `json:"status"`
}

// PostAnalytics represents statistics for a specific post.
type PostAnalytics struct {
	PostURN          string `json:"postUrn"`
	Impressions      int64  `json:"impressions"`
	UniqueImpressions int64 `json:"uniqueImpressions"`
	Likes            int64  `json:"likes"`
	Comments         int64  `json:"comments"`
	Shares           int64  `json:"shares"`
	Clicks           int64  `json:"clicks"`
	EngagementRate   float64 `json:"engagementRate"`
}

// TokenResponse represents response from OAuth token exchange or refresh.
type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in,omitempty"`
	Scope                 string `json:"scope,omitempty"`
	TokenType             string `json:"token_type,omitempty"`
}

