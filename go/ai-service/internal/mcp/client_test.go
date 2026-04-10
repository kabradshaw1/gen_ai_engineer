package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
)

func TestDiscoverTools_WrapsRemoteTools(t *testing.T) {
	// Stand up an in-process MCP server with one tool.
	reg := tools.NewMemRegistry()
	reg.Register(&fakeTool{
		name:   "search_products",
		desc:   "Search products",
		schema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`),
		result: tools.Result{Content: map[string]any{"products": []string{"p1"}}},
	})
	srv := NewServer(reg, Defaults{})

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	ctx := context.Background()
	discovered, err := DiscoverTools(ctx, session, "remote")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(discovered) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(discovered))
	}

	tool := discovered[0]
	if tool.Name() != "remote.search_products" {
		t.Errorf("expected 'remote.search_products', got %q", tool.Name())
	}
	if tool.Description() != "Search products" {
		t.Errorf("unexpected description: %q", tool.Description())
	}
	if len(tool.Schema()) == 0 {
		t.Error("expected non-empty schema")
	}
}

func TestMCPClientTool_Call_Success(t *testing.T) {
	reg := tools.NewMemRegistry()
	reg.Register(&fakeTool{
		name:   "get_product",
		desc:   "Get product",
		schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
		result: tools.Result{Content: map[string]any{"id": "p1", "name": "Widget"}},
	})
	srv := NewServer(reg, Defaults{})

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	ctx := context.Background()
	discovered, err := DiscoverTools(ctx, session, "remote")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}

	result, err := discovered[0].Call(ctx, json.RawMessage(`{"id":"p1"}`), "")
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if result.Content == nil {
		t.Error("expected non-nil content")
	}
}

func TestMCPClientTool_Call_ToolError(t *testing.T) {
	reg := tools.NewMemRegistry()
	reg.Register(&fakeTool{
		name:   "bad_tool",
		desc:   "Always errors",
		schema: json.RawMessage(`{"type":"object"}`),
		err:    fmt.Errorf("boom"),
	})
	srv := NewServer(reg, Defaults{})

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	ctx := context.Background()
	discovered, err := DiscoverTools(ctx, session, "remote")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}

	_, err = discovered[0].Call(ctx, json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error from failing tool")
	}
}
