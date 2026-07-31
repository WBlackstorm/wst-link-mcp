package mcp

import (
	"testing"
)

func TestNewServerEmptyToken(t *testing.T) {
	_, err := NewServer("")
	if err == nil {
		t.Fatal("expected error when access token is empty, got nil")
	}
}

func TestNewServerSuccess(t *testing.T) {
	s, err := NewServer("test-access-token")
	if err != nil {
		t.Fatalf("unexpected error initializing MCP server: %v", err)
	}

	if s == nil {
		t.Fatal("expected non-nil server instance")
	}
}
