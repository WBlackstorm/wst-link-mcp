package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/weverton/wst-link-mcp/pkg/linkedin"
)

func main() {
	clientID := flag.String("client-id", "", "LinkedIn App Client ID")
	clientSecret := flag.String("client-secret", "", "LinkedIn App Client Secret")
	code := flag.String("code", "", "Authorization code from OAuth callback")
	redirectURI := flag.String("redirect-uri", "", "OAuth Redirect URI (e.g. http://localhost:8080/callback)")
	refreshToken := flag.String("refresh-token", "", "Refresh token (if renewing existing token)")
	baseURL := flag.String("base-url", "https://www.linkedin.com", "OAuth Base URL")
	flag.Parse()

	if *clientID == "" {
		*clientID = os.Getenv("LINKEDIN_CLIENT_ID")
	}
	if *clientSecret == "" {
		*clientSecret = os.Getenv("LINKEDIN_CLIENT_SECRET")
	}
	if *refreshToken == "" {
		*refreshToken = os.Getenv("LINKEDIN_REFRESH_TOKEN")
	}

	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: client-id and client-secret are required (or set LINKEDIN_CLIENT_ID and LINKEDIN_CLIENT_SECRET)\n")
		os.Exit(1)
	}

	if *code == "" && *refreshToken == "" {
		fmt.Fprintf(os.Stderr, "Usage: token-refresher --client-id=ID --client-secret=SECRET --code=AUTH_CODE --redirect-uri=URI\n")
		fmt.Fprintf(os.Stderr, "   or: token-refresher --client-id=ID --client-secret=SECRET --refresh-token=TOKEN\n")
		os.Exit(1)
	}

	client := linkedin.NewClient("")
	client.SetBaseURL(*baseURL)

	ctx := context.Background()
	tokenResp, err := client.ExchangeOAuthToken(ctx, *clientID, *clientSecret, *code, *redirectURI, *refreshToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exchanging token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🎉 Token exchange successful!")
	fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
	fmt.Printf("Expires In: %d seconds\n", tokenResp.ExpiresIn)
	if tokenResp.RefreshToken != "" {
		fmt.Printf("Refresh Token: %s\n", tokenResp.RefreshToken)
	}
}
