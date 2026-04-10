package tools

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeTool struct {
	name    string
	schema  json.RawMessage
	calls   int
	result  Result
	callErr error
}

func (f *fakeTool) Name() string            { return f.name }
func (f *fakeTool) Description() string     { return "fake " + f.name }
func (f *fakeTool) Schema() json.RawMessage { return f.schema }
func (f *fakeTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	f.calls++
	return f.result, f.callErr
}

func TestMemRegistry_RegisterAndGet(t *testing.T) {
	reg := NewMemRegistry()
	tool := &fakeTool{name: "search_products", schema: json.RawMessage(`{"type":"object"}`)}
	reg.Register(tool)

	got, ok := reg.Get("search_products")
	if !ok {
		t.Fatal("expected to find search_products")
	}
	if got.Name() != "search_products" {
		t.Errorf("got %q", got.Name())
	}
	if _, ok := reg.Get("nope"); ok {
		t.Error("expected miss for unknown tool")
	}
}

func TestMemRegistry_Schemas(t *testing.T) {
	reg := NewMemRegistry()
	reg.Register(&fakeTool{name: "a", schema: json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}}}`)})
	reg.Register(&fakeTool{name: "b", schema: json.RawMessage(`{"type":"object"}`)})

	schemas := reg.Schemas()
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
	names := map[string]bool{}
	for _, s := range schemas {
		names[s.Name] = true
		if len(s.Parameters) == 0 {
			t.Errorf("schema %q has empty Parameters", s.Name)
		}
	}
	if !names["a"] || !names["b"] {
		t.Errorf("missing schemas: %v", names)
	}
}

func TestMemRegistry_All(t *testing.T) {
	reg := NewMemRegistry()
	reg.Register(&fakeTool{name: "a", schema: json.RawMessage(`{"type":"object"}`)})
	reg.Register(&fakeTool{name: "b", schema: json.RawMessage(`{"type":"object"}`)})

	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(all))
	}
	names := map[string]bool{}
	for _, tool := range all {
		names[tool.Name()] = true
	}
	if !names["a"] || !names["b"] {
		t.Errorf("missing tools: %v", names)
	}
}
