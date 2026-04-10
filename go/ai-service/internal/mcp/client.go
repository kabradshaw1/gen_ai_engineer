package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
)

// MCPClientTool wraps a tool discovered from an MCP server as a tools.Tool.
type MCPClientTool struct {
	prefixedName string
	description  string
	schema       json.RawMessage
	session      *sdkmcp.ClientSession
	remoteName   string
}

func (t *MCPClientTool) Name() string            { return t.prefixedName }
func (t *MCPClientTool) Description() string     { return t.description }
func (t *MCPClientTool) Schema() json.RawMessage { return t.schema }

func (t *MCPClientTool) Call(ctx context.Context, args json.RawMessage, userID string) (tools.Result, error) {
	var arguments map[string]any
	if err := json.Unmarshal(args, &arguments); err != nil {
		return tools.Result{}, fmt.Errorf("mcp client: bad args: %w", err)
	}

	result, err := t.session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      t.remoteName,
		Arguments: arguments,
	})
	if err != nil {
		return tools.Result{}, fmt.Errorf("mcp client: call %s: %w", t.remoteName, err)
	}

	if result.IsError {
		text := extractText(result)
		return tools.Result{}, fmt.Errorf("mcp tool error: %s", text)
	}

	text := extractText(result)
	var content any
	if err := json.Unmarshal([]byte(text), &content); err != nil {
		return tools.Result{Content: text}, nil
	}
	return tools.Result{Content: content}, nil
}

// extractText returns the concatenated text content from a CallToolResult.
func extractText(r *sdkmcp.CallToolResult) string {
	var parts []string
	for _, c := range r.Content {
		if tc, ok := c.(*sdkmcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "")
}

// DiscoverTools connects to an MCP server session, lists all available tools,
// and returns them wrapped as tools.Tool. Each tool name is prefixed with
// serverName + "." to avoid collisions with local tools.
func DiscoverTools(ctx context.Context, session *sdkmcp.ClientSession, serverName string) ([]tools.Tool, error) {
	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp discover %s: list tools: %w", serverName, err)
	}

	out := make([]tools.Tool, 0, len(result.Tools))
	for _, t := range result.Tools {
		schema, _ := json.Marshal(t.InputSchema)
		out = append(out, &MCPClientTool{
			prefixedName: serverName + "." + t.Name,
			description:  t.Description,
			schema:       schema,
			session:      session,
			remoteName:   t.Name,
		})
	}
	return out, nil
}
