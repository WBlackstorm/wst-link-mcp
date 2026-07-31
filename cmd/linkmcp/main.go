package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weverton/wst-link-mcp/pkg/mcp"
)

func main() {
	accessToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Fprintf(os.Stderr, "Error: LINKEDIN_ACCESS_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	s, err := mcp.NewServer(accessToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing MCP server: %v\n", err)
		os.Exit(1)
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
