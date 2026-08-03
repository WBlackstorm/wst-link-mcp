package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weverton/wst-link-mcp/pkg/auth"
	"github.com/weverton/wst-link-mcp/pkg/mcp"
)

func main() {
	modeFlag := flag.String("mode", "stdio", "Server mode: 'stdio' or 'sse'")
	portFlag := flag.String("port", "8080", "Port to run SSE server on (e.g. 8080)")
	clientIDFlag := flag.String("client-id", "", "LinkedIn Client ID")
	clientSecretFlag := flag.String("client-secret", "", "LinkedIn Client Secret")
	redirectURIFlag := flag.String("redirect-uri", "", "LinkedIn OAuth Redirect URI")
	refreshTokenFlag := flag.String("refresh-token", "", "LinkedIn Refresh Token")
	flag.Parse()

	clientID := *clientIDFlag
	if clientID == "" {
		clientID = os.Getenv("LINKEDIN_CLIENT_ID")
	}

	clientSecret := *clientSecretFlag
	if clientSecret == "" {
		clientSecret = os.Getenv("LINKEDIN_CLIENT_SECRET")
	}

	redirectURI := *redirectURIFlag
	if redirectURI == "" {
		redirectURI = os.Getenv("LINKEDIN_REDIRECT_URI")
	}
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/callback"
	}

	refreshToken := *refreshTokenFlag
	if refreshToken == "" {
		refreshToken = os.Getenv("LINKEDIN_REFRESH_TOKEN")
	}

	if clientID == "" || clientSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: LINKEDIN_CLIENT_ID and LINKEDIN_CLIENT_SECRET are required (or set via --client-id and --client-secret)\n")
		os.Exit(1)
	}

	authenticator := auth.NewAuthenticator(auth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		RefreshToken: refreshToken,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	accessToken, err := authenticator.GetOrFetchToken(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error obtaining access token: %v\n", err)
		os.Exit(1)
	}

	s, err := mcp.NewServer(accessToken)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing MCP server: %v\n", err)
		os.Exit(1)
	}

	mode := *modeFlag
	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		*portFlag = envPort
	}

	if mode == "sse" {
		addr := ":" + *portFlag
		fmt.Fprintf(os.Stderr, "Starting LinkMCP server in SSE mode on %s...\n", addr)
		sseServer := mcp.NewSSEServer(s)
		if err := sseServer.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "SSE Server error: %v\n", err)
			os.Exit(1)
		}
	} else {


		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}

