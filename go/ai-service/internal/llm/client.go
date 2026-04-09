package llm

import "context"

// Client is the abstraction every LLM backend implements. The agent loop
// depends only on this interface.
type Client interface {
	// Chat sends the full message history and the advertised tool schemas
	// and returns either a final text or a list of tool calls.
	Chat(ctx context.Context, messages []Message, tools []ToolSchema) (ChatResponse, error)
}
