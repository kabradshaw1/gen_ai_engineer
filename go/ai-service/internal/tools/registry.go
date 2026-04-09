package tools

import (
	"context"
	"encoding/json"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
)

// Tool is the only interface a future MCP adapter needs to satisfy.
type Tool interface {
	Name() string
	Description() string
	Schema() json.RawMessage
	Call(ctx context.Context, args json.RawMessage, userID string) (Result, error)
}

// Result is what a tool returns. Content is what the LLM sees (compact, JSON-serializable).
// Display is an optional richer payload for the frontend (e.g. product cards).
type Result struct {
	Content any `json:"content"`
	Display any `json:"display,omitempty"`
}

// Registry holds tool implementations keyed by name.
type Registry interface {
	Register(Tool)
	Get(name string) (Tool, bool)
	Schemas() []llm.ToolSchema
}

// NewMemRegistry returns an in-memory Registry.
func NewMemRegistry() *MemRegistry {
	return &MemRegistry{tools: map[string]Tool{}}
}

type MemRegistry struct {
	tools map[string]Tool
}

func (r *MemRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *MemRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *MemRegistry) Schemas() []llm.ToolSchema {
	out := make([]llm.ToolSchema, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, llm.ToolSchema{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return out
}
