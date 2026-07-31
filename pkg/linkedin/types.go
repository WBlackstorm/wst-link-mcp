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
}

// PostResponse represents the response received after creating a post.
type PostResponse struct {
	ID         string `json:"id,omitempty"`
	URN        string `json:"urn,omitempty"`
	Status     int    `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
	Commentary string `json:"commentary,omitempty"`
}
