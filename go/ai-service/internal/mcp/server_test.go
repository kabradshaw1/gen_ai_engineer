package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
)

// fakeTool is a minimal tools.Tool for testing.
type fakeTool struct {
	name   string
	desc   string
	schema json.RawMessage
	result tools.Result
	err    error
	calls  int
	seenID string
}

func (f *fakeTool) Name() string            { return f.name }
func (f *fakeTool) Description() string     { return f.desc }
func (f *fakeTool) Schema() json.RawMessage { return f.schema }
func (f *fakeTool) Call(ctx context.Context, args json.RawMessage, userID string) (tools.Result, error) {
	f.calls++
	f.seenID = userID
	return f.result, f.err
}

// connectInProcess wires a server to a client using in-process transports.
// The server transport is connected first (as required by the SDK), then the client.
// Returns the connected client session and a cleanup function.
func connectInProcess(t *testing.T, srv *sdkmcp.Server) (*sdkmcp.ClientSession, func()) {
	t.Helper()
	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()

	ctx := context.Background()

	// Connect server side first (SDK requirement).
	_, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}

	return session, func() { session.Close() }
}

func TestNewServer_RegistersAllTools(t *testing.T) {
	reg := tools.NewMemRegistry()
	reg.Register(&fakeTool{
		name:   "search_products",
		desc:   "Search products",
		schema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`),
	})
	reg.Register(&fakeTool{
		name:   "get_product",
		desc:   "Get one product",
		schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
	})

	srv := NewServer(reg, Defaults{})
	ctx := context.Background()

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(result.Tools))
	}
	names := map[string]bool{}
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	if !names["search_products"] || !names["get_product"] {
		t.Errorf("unexpected tools: %v", names)
	}
}

func TestServer_CallTool_Success(t *testing.T) {
	ft := &fakeTool{
		name:   "get_product",
		desc:   "Get a product",
		schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
		result: tools.Result{Content: map[string]any{"id": "p1", "name": "Widget"}},
	}
	reg := tools.NewMemRegistry()
	reg.Register(ft)

	srv := NewServer(reg, Defaults{})
	ctx := context.Background()

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "get_product",
		Arguments: map[string]any{"id": "p1"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	if ft.calls != 1 {
		t.Errorf("expected 1 call, got %d", ft.calls)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
}

func TestServer_CallTool_ToolError(t *testing.T) {
	ft := &fakeTool{
		name:   "get_order",
		desc:   "Get order",
		schema: json.RawMessage(`{"type":"object"}`),
		err:    fmt.Errorf("upstream 500"),
	}
	reg := tools.NewMemRegistry()
	reg.Register(ft)

	srv := NewServer(reg, Defaults{})
	ctx := context.Background()

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "get_order",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result")
	}
}

func TestServer_DefaultUserID(t *testing.T) {
	ft := &fakeTool{
		name:   "list_orders",
		desc:   "List orders",
		schema: json.RawMessage(`{"type":"object"}`),
		result: tools.Result{Content: []string{"order1"}},
	}
	reg := tools.NewMemRegistry()
	reg.Register(ft)

	srv := NewServer(reg, Defaults{UserID: "user-42"})
	ctx := context.Background()

	session, cleanup := connectInProcess(t, srv)
	defer cleanup()

	_, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "list_orders",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if ft.seenID != "user-42" {
		t.Errorf("expected userID 'user-42', got %q", ft.seenID)
	}
}
