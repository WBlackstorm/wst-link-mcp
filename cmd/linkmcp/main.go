package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weverton/wst-link-mcp/pkg/mcp"
)

func main() {
	modeFlag := flag.String("mode", "stdio", "Server mode: 'stdio' or 'sse'")
	portFlag := flag.String("port", "8080", "Port to run SSE server on (e.g. 8080)")
	flag.Parse()

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

	mode := *modeFlag
	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		*portFlag = envPort
	}

	if mode == "sse" {
		addr := ":" + *portFlag
		fmt.Printf("Starting LinkMCP server in SSE mode on %s...\n", addr)
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

