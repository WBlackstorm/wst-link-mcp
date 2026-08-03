package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/weverton/wst-link-mcp/pkg/linkedin"
)

// Config holds OAuth client parameters.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	RefreshToken string
}

// TokenStore represents the structure of the saved token JSON file.
type TokenStore struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// Authenticator handles automatic OAuth 2.0 flow for LinkedIn.
type Authenticator struct {
	cfg    Config
	client *linkedin.Client
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(cfg Config) *Authenticator {
	return &Authenticator{
		cfg:    cfg,
		client: linkedin.NewClient(""),
	}
}

func (a *Authenticator) getTokenFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".wst-link-mcp-token.json"
	}
	return filepath.Join(home, ".wst-link-mcp-token.json")
}

func (a *Authenticator) loadToken() (*TokenStore, error) {
	path := a.getTokenFilePath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ts TokenStore
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, err
	}
	return &ts, nil
}

func (a *Authenticator) saveToken(accessToken, refreshToken string, expiresIn int64) error {
	path := a.getTokenFilePath()
	ts := TokenStore{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if expiresIn > 0 {
		ts.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

// GetOrFetchToken retrieves access token via local store, refresh token or opens local browser for OAuth 2.0 flow.
func (a *Authenticator) GetOrFetchToken(ctx context.Context) (string, error) {
	// 1. Try to load cached token from file
	if ts, err := a.loadToken(); err == nil {
		// If access token is valid and not expired, use it
		if ts.AccessToken != "" && (ts.Expiry.IsZero() || ts.Expiry.After(time.Now().Add(5*time.Minute))) {
			return ts.AccessToken, nil
		}
		// If expired but we have a refresh token, try that
		if ts.RefreshToken != "" {
			tokenResp, err := a.client.ExchangeOAuthToken(ctx, a.cfg.ClientID, a.cfg.ClientSecret, "", "", ts.RefreshToken)
			if err == nil && tokenResp.AccessToken != "" {
				// Save refreshed token
				newRefresh := ts.RefreshToken
				if tokenResp.RefreshToken != "" {
					newRefresh = tokenResp.RefreshToken
				}
				_ = a.saveToken(tokenResp.AccessToken, newRefresh, tokenResp.ExpiresIn)
				return tokenResp.AccessToken, nil
			}
		}
	}

	// 2. If refresh token is available from config, attempt to use it
	if a.cfg.RefreshToken != "" {
		tokenResp, err := a.client.ExchangeOAuthToken(ctx, a.cfg.ClientID, a.cfg.ClientSecret, "", "", a.cfg.RefreshToken)
		if err == nil && tokenResp.AccessToken != "" {
			_ = a.saveToken(tokenResp.AccessToken, a.cfg.RefreshToken, tokenResp.ExpiresIn)
			return tokenResp.AccessToken, nil
		}
		fmt.Fprintf(os.Stderr, "Refresh token exchange failed or expired (%v). Starting interactive OAuth flow...\n", err)
	}

	// 2. Start local HTTP server to receive OAuth callback code automatically
	redirectURL, err := url.Parse(a.cfg.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	port := redirectURL.Port()
	if port == "" {
		port = "8080"
	}
	path := redirectURL.Path
	if path == "" {
		path = "/callback"
	}

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	var server *http.Server
	var once sync.Once

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		authErr := r.URL.Query().Get("error")

		if authErr != "" {
			desc := r.URL.Query().Get("error_description")
			fmt.Fprintf(w, "<html><body><h2>LinkedIn Authentication Failed</h2><p>%s (%s)</p></body></html>", authErr, desc)
			once.Do(func() {
				errChan <- fmt.Errorf("authentication denied by user: %s - %s", authErr, desc)
			})
			return
		}

		if code == "" {
			http.Error(w, "Missing code query parameter", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "<html><body><h2>Authentication Successful!</h2><p>You can close this tab and return to your application/Antigravity.</p></body></html>")

		once.Do(func() {
			codeChan <- code
		})
	})

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return "", fmt.Errorf("failed to listen on port %s for OAuth callback: %w", port, err)
	}

	server = &http.Server{Handler: mux}

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			once.Do(func() {
				errChan <- serveErr
			})
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	authURL := fmt.Sprintf(
		"https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=%s&redirect_uri=%s&scope=w_member_social%%20openid%%20profile%%20email",
		url.QueryEscape(a.cfg.ClientID),
		url.QueryEscape(a.cfg.RedirectURI),
	)

	fmt.Fprintf(os.Stderr, "\n======================================================\n")
	fmt.Fprintf(os.Stderr, "🔑 Opening browser for LinkedIn OAuth authorization:\n")
	fmt.Fprintf(os.Stderr, "%s\n", authURL)
	fmt.Fprintf(os.Stderr, "======================================================\n\n")

	_ = openBrowser(authURL)

	select {
	case code := <-codeChan:
		tokenResp, err := a.client.ExchangeOAuthToken(ctx, a.cfg.ClientID, a.cfg.ClientSecret, code, a.cfg.RedirectURI, "")
		if err != nil {
			return "", fmt.Errorf("failed to exchange authorization code for access token: %w", err)
		}
		_ = a.saveToken(tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn)
		if tokenResp.RefreshToken != "" {
			fmt.Fprintf(os.Stderr, "💡 Refresh Token: %s\n", tokenResp.RefreshToken)
		}
		return tokenResp.AccessToken, nil
	case err := <-errChan:
		return "", fmt.Errorf("OAuth authorization error: %w", err)
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(5 * time.Minute):
		return "", fmt.Errorf("OAuth authorization timed out after 5 minutes")
	}
}

func openBrowser(urlStr string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", urlStr).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr).Start()
	case "darwin":
		err = exec.Command("open", urlStr).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
