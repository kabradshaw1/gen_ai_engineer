# Plan 1 — `go/ai-service` MVP: Scaffold, Agent Loop, Catalog Tools

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a runnable Go microservice (`go/ai-service`) with an LLM agent loop over Ollama's tool-calling API, a tool registry, three unauthenticated catalog tools, and an SSE `/chat` endpoint — end-to-end usable from `curl` against `docker compose`.

**Architecture:** New Go service sibling to `auth-service` and `ecommerce-service`. `agent` package owns the loop over an `llm.Client` interface and a `tools.Registry` interface (both injected, both fake-able). Three catalog tools (`search_products`, `get_product`, `check_inventory`) wrap the existing `ecommerce-service` REST API. HTTP layer uses Gin + SSE. No auth in this plan — catalog tools are public.

**Tech Stack:** Go 1.26, Gin, `net/http` (Ollama client), `httptest` (unit tests), Docker Compose, existing `ecommerce-service` as the data source.

**Scope boundaries:**
- No JWT, no user-scoped tools, no orders/cart tools (→ Plan 2).
- `search_products` uses text search via `ecommerce-service` `GET /products?q=...`. Semantic Qdrant search is deferred to Plan 3 where it lands together with the embedding cache.
- No Redis caching, no evals, no metrics, no guardrails (→ Plan 3).
- No frontend (→ Plan 4).
- No K8s manifests, no CI changes (→ Plan 5).

**Reference:** `docs/superpowers/specs/2026-04-09-go-ai-service-agent-design.md`, sections 2 (architecture), 3 (agent loop), 4 (tool interface and catalog tools), 6.1 (HTTP surface).

**Module path:** `github.com/kabradshaw1/portfolio/go/ai-service`

---

## File Map

Files created in this plan:

```
go/ai-service/
├── go.mod
├── go.sum
├── Dockerfile
├── cmd/server/main.go
├── internal/
│   ├── llm/
│   │   ├── types.go            # Message, ToolCall, ChatResponse, ToolSchema
│   │   ├── client.go           # Client interface
│   │   ├── ollama.go           # OllamaClient
│   │   └── ollama_test.go
│   ├── tools/
│   │   ├── registry.go         # Tool, Result, Registry interface + memRegistry
│   │   ├── registry_test.go
│   │   ├── clients/
│   │   │   ├── ecommerce.go    # Typed HTTP client for ecommerce-service
│   │   │   └── ecommerce_test.go
│   │   └── catalog.go          # search_products, get_product, check_inventory
│   │   └── catalog_test.go
│   ├── agent/
│   │   ├── events.go           # Event types (ToolCallEvent, ToolResultEvent, ...)
│   │   ├── agent.go            # Agent.Run
│   │   └── agent_test.go
│   └── http/
│       ├── chat.go             # POST /chat SSE handler
│       ├── chat_test.go
│       └── health.go           # GET /health, GET /ready
```

Files modified:
- `docker-compose.yml` — add `ai-service` service
- `Makefile` — add `preflight-ai-service` target, extend `preflight-go`

---

## Task 1: Scaffold the Go module and smoke test

**Files:**
- Create: `go/ai-service/go.mod`
- Create: `go/ai-service/cmd/server/main.go`
- Create: `go/ai-service/internal/.gitkeep`

- [ ] **Step 1: Create directory layout**

```bash
mkdir -p go/ai-service/cmd/server
mkdir -p go/ai-service/internal/{llm,tools/clients,agent,http}
touch go/ai-service/internal/.gitkeep
```

- [ ] **Step 2: Initialize the Go module**

```bash
cd go/ai-service && go mod init github.com/kabradshaw1/portfolio/go/ai-service
```

Expected: creates `go/ai-service/go.mod` with `module github.com/kabradshaw1/portfolio/go/ai-service` and `go 1.26.x`.

- [ ] **Step 3: Add Gin dependency**

```bash
cd go/ai-service && go get github.com/gin-gonic/gin@v1.12.0
```

- [ ] **Step 4: Write a minimal `main.go` that boots Gin and returns 200 on `/health`**

Contents of `go/ai-service/cmd/server/main.go`:

```go
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8093"
	}

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	slog.Info("ai-service starting", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
```

- [ ] **Step 5: Verify it builds and runs**

```bash
cd go/ai-service && go build ./...
```

Expected: no output, exit 0.

```bash
cd go/ai-service && go run ./cmd/server &
sleep 1 && curl -s localhost:8093/health && kill %1
```

Expected: `{"status":"ok"}`.

- [ ] **Step 6: Commit**

```bash
git add go/ai-service/
git commit -m "feat(ai-service): scaffold Go module with health endpoint"
```

---

## Task 2: Define the `llm.Client` interface and message types

**Files:**
- Create: `go/ai-service/internal/llm/types.go`
- Create: `go/ai-service/internal/llm/client.go`

- [ ] **Step 1: Write `types.go`**

```go
package llm

import "encoding/json"

// Role is the sender of a chat message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is one entry in the chat history sent to the LLM.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"` // tool name for role=tool
}

// ToolCall is a single tool invocation requested by the model.
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"arguments"`
}

// ChatResponse is what the model returns for one Chat call.
type ChatResponse struct {
	Content   string     // final text if no tool calls
	ToolCalls []ToolCall // non-empty when the model wants to call tools
}

// ToolSchema is the JSON-Schema-shaped description of a tool we advertise
// to the LLM on every turn.
type ToolSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ToolResultMessage builds a role="tool" message from a tool's JSON-serializable result.
func ToolResultMessage(callID, toolName string, content any) (Message, error) {
	body, err := json.Marshal(content)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Role:       RoleTool,
		ToolCallID: callID,
		Name:       toolName,
		Content:    string(body),
	}, nil
}
```

- [ ] **Step 2: Write `client.go`**

```go
package llm

import "context"

// Client is the abstraction every LLM backend implements. The agent loop
// depends only on this interface.
type Client interface {
	// Chat sends the full message history and the advertised tool schemas
	// and returns either a final text or a list of tool calls.
	Chat(ctx context.Context, messages []Message, tools []ToolSchema) (ChatResponse, error)
}
```

- [ ] **Step 3: Verify the package compiles**

```bash
cd go/ai-service && go build ./internal/llm/...
```

Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add go/ai-service/internal/llm/
git commit -m "feat(ai-service): define llm.Client interface and message types"
```

---

## Task 3: Implement `OllamaClient` with an httptest-driven test

**Files:**
- Create: `go/ai-service/internal/llm/ollama.go`
- Create: `go/ai-service/internal/llm/ollama_test.go`

- [ ] **Step 1: Write the failing test first**

Contents of `go/ai-service/internal/llm/ollama_test.go`:

```go
package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOllamaClient_Chat_FinalText(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		_, _ = w.Write([]byte(`{
			"model":"qwen2.5",
			"message":{"role":"assistant","content":"hello there"},
			"done":true
		}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "qwen2.5")
	resp, err := client.Chat(context.Background(),
		[]Message{{Role: RoleUser, Content: "hi"}},
		nil,
	)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if resp.Content != "hello there" {
		t.Errorf("expected content 'hello there', got %q", resp.Content)
	}
	if len(resp.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(resp.ToolCalls))
	}
	if gotBody["model"] != "qwen2.5" {
		t.Errorf("expected model qwen2.5, got %v", gotBody["model"])
	}
	if gotBody["stream"] != false {
		t.Errorf("expected stream=false, got %v", gotBody["stream"])
	}
}

func TestOllamaClient_Chat_ToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"model":"qwen2.5",
			"message":{
				"role":"assistant",
				"content":"",
				"tool_calls":[
					{"function":{"name":"search_products","arguments":{"query":"jacket","max_price":150}}}
				]
			},
			"done":true
		}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "qwen2.5")
	resp, err := client.Chat(context.Background(),
		[]Message{{Role: RoleUser, Content: "find a jacket"}},
		[]ToolSchema{{Name: "search_products", Description: "", Parameters: json.RawMessage(`{}`)}},
	)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "search_products" {
		t.Errorf("expected name search_products, got %q", resp.ToolCalls[0].Name)
	}
	if !strings.Contains(string(resp.ToolCalls[0].Args), "jacket") {
		t.Errorf("expected args to contain 'jacket', got %s", resp.ToolCalls[0].Args)
	}
	if resp.ToolCalls[0].ID == "" {
		t.Errorf("expected a generated tool call id")
	}
}

func TestOllamaClient_Chat_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "qwen2.5")
	_, err := client.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}
```

- [ ] **Step 2: Run the test, expect a compile failure**

```bash
cd go/ai-service && go test ./internal/llm/...
```

Expected: compile error — `undefined: NewOllamaClient`.

- [ ] **Step 3: Implement `ollama.go`**

Contents of `go/ai-service/internal/llm/ollama.go`:

```go
package llm

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient talks to an Ollama server's /api/chat endpoint using
// the OpenAI-compatible tool-calling shape that Qwen 2.5 supports.
type OllamaClient struct {
	baseURL string
	model   string
	http    *http.Client
}

// NewOllamaClient returns a Client pointed at baseURL (e.g. "http://ollama:11434").
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

type ollamaTool struct {
	Type     string      `json:"type"`
	Function ollamaToolF `json:"function"`
}
type ollamaToolF struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ollamaReq struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Tools    []ollamaTool `json:"tools,omitempty"`
	Stream   bool         `json:"stream"`
}

type ollamaResp struct {
	Message struct {
		Role      Role   `json:"role"`
		Content   string `json:"content"`
		ToolCalls []struct {
			Function struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (c *OllamaClient) Chat(ctx context.Context, messages []Message, tools []ToolSchema) (ChatResponse, error) {
	reqBody := ollamaReq{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}
	for _, t := range tools {
		reqBody.Tools = append(reqBody.Tools, ollamaTool{
			Type: "function",
			Function: ollamaToolF{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(payload))
	}

	var parsed ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ChatResponse{}, fmt.Errorf("decode response: %w", err)
	}

	out := ChatResponse{Content: parsed.Message.Content}
	for _, tc := range parsed.Message.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCall{
			ID:   newCallID(),
			Name: tc.Function.Name,
			Args: tc.Function.Arguments,
		})
	}
	return out, nil
}

func newCallID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "call_" + hex.EncodeToString(b[:])
}
```

- [ ] **Step 4: Run the tests, expect pass**

```bash
cd go/ai-service && go test ./internal/llm/... -v
```

Expected: `PASS` for all three tests.

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/llm/ go/ai-service/go.sum
git commit -m "feat(ai-service): implement OllamaClient with tool-calling support"
```

---

## Task 4: Define `tools.Tool`, `tools.Result`, and a `Registry`

**Files:**
- Create: `go/ai-service/internal/tools/registry.go`
- Create: `go/ai-service/internal/tools/registry_test.go`

- [ ] **Step 1: Write the failing test**

Contents of `go/ai-service/internal/tools/registry_test.go`:

```go
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

func (f *fakeTool) Name() string                { return f.name }
func (f *fakeTool) Description() string         { return "fake " + f.name }
func (f *fakeTool) Schema() json.RawMessage     { return f.schema }
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
```

- [ ] **Step 2: Run the test, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/...
```

Expected: compile error — `undefined: Result`, `undefined: NewMemRegistry`.

- [ ] **Step 3: Implement `registry.go`**

Contents of `go/ai-service/internal/tools/registry.go`:

```go
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
```

- [ ] **Step 4: Run tests, expect pass**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

Expected: `PASS` for `TestMemRegistry_RegisterAndGet` and `TestMemRegistry_Schemas`.

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/tools/
git commit -m "feat(ai-service): add tool registry and Tool/Result interface"
```

---

## Task 5: Implement the `agent.Agent` loop with fake dependencies

**Files:**
- Create: `go/ai-service/internal/agent/events.go`
- Create: `go/ai-service/internal/agent/agent.go`
- Create: `go/ai-service/internal/agent/agent_test.go`

- [ ] **Step 1: Write `events.go`**

```go
package agent

import "encoding/json"

// Event is the sum type emitted by the agent loop.
// Exactly one of the concrete event structs is non-zero.
type Event struct {
	ToolCall   *ToolCallEvent   `json:"tool_call,omitempty"`
	ToolResult *ToolResultEvent `json:"tool_result,omitempty"`
	ToolError  *ToolErrorEvent  `json:"tool_error,omitempty"`
	Final      *FinalEvent      `json:"final,omitempty"`
	Error      *ErrorEvent      `json:"error,omitempty"`
}

type ToolCallEvent struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type ToolResultEvent struct {
	Name    string `json:"name"`
	Display any    `json:"display,omitempty"`
}

type ToolErrorEvent struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type FinalEvent struct {
	Text string `json:"text"`
}

type ErrorEvent struct {
	Reason string `json:"reason"`
}
```

- [ ] **Step 2: Write the failing test**

Contents of `go/ai-service/internal/agent/agent_test.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
)

// --- fake llm.Client that returns canned responses in order ---

type fakeLLM struct {
	responses []llm.ChatResponse
	err       error
	calls     int
}

func (f *fakeLLM) Chat(ctx context.Context, msgs []llm.Message, ts []llm.ToolSchema) (llm.ChatResponse, error) {
	if f.err != nil {
		return llm.ChatResponse{}, f.err
	}
	if f.calls >= len(f.responses) {
		return llm.ChatResponse{}, errors.New("unexpected extra call")
	}
	r := f.responses[f.calls]
	f.calls++
	return r, nil
}

// --- fake tool ---

type scriptedTool struct {
	name   string
	result tools.Result
	err    error
	calls  int
}

func (s *scriptedTool) Name() string            { return s.name }
func (s *scriptedTool) Description() string     { return "" }
func (s *scriptedTool) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (s *scriptedTool) Call(ctx context.Context, args json.RawMessage, userID string) (tools.Result, error) {
	s.calls++
	return s.result, s.err
}

func collect(events *[]Event) func(Event) {
	return func(e Event) { *events = append(*events, e) }
}

func TestAgent_FinalOnFirstTurn(t *testing.T) {
	llmc := &fakeLLM{responses: []llm.ChatResponse{{Content: "hi there"}}}
	reg := tools.NewMemRegistry()
	a := New(llmc, reg, 8, 5*time.Second)

	var events []Event
	err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}}}, collect(&events))
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(events) != 1 || events[0].Final == nil || events[0].Final.Text != "hi there" {
		t.Fatalf("expected single final event, got %+v", events)
	}
}

func TestAgent_ToolCallThenFinal(t *testing.T) {
	llmc := &fakeLLM{responses: []llm.ChatResponse{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "echo", Args: json.RawMessage(`{"x":1}`)}}},
		{Content: "done"},
	}}
	tool := &scriptedTool{name: "echo", result: tools.Result{Content: map[string]any{"ok": true}}}
	reg := tools.NewMemRegistry()
	reg.Register(tool)

	a := New(llmc, reg, 8, 5*time.Second)
	var events []Event
	err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "go"}}}, collect(&events))
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if tool.calls != 1 {
		t.Errorf("expected tool called once, got %d", tool.calls)
	}
	// Expect: ToolCallEvent, ToolResultEvent, FinalEvent
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d: %+v", len(events), events)
	}
	if events[0].ToolCall == nil || events[0].ToolCall.Name != "echo" {
		t.Errorf("event[0] = %+v", events[0])
	}
	if events[1].ToolResult == nil || events[1].ToolResult.Name != "echo" {
		t.Errorf("event[1] = %+v", events[1])
	}
	if events[2].Final == nil || events[2].Final.Text != "done" {
		t.Errorf("event[2] = %+v", events[2])
	}
}

func TestAgent_UnknownToolRecoversAndContinues(t *testing.T) {
	llmc := &fakeLLM{responses: []llm.ChatResponse{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "missing", Args: json.RawMessage(`{}`)}}},
		{Content: "ok"},
	}}
	a := New(llmc, tools.NewMemRegistry(), 8, 5*time.Second)
	var events []Event
	err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "go"}}}, collect(&events))
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// Unknown tool becomes a tool_error event, loop continues, final answer lands.
	if events[len(events)-1].Final == nil {
		t.Errorf("expected final event at end, got %+v", events[len(events)-1])
	}
	foundErr := false
	for _, e := range events {
		if e.ToolError != nil && e.ToolError.Name == "missing" {
			foundErr = true
		}
	}
	if !foundErr {
		t.Errorf("expected a tool_error event for missing tool")
	}
}

func TestAgent_ToolErrorIsFedBackNotBubbled(t *testing.T) {
	llmc := &fakeLLM{responses: []llm.ChatResponse{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "flaky", Args: json.RawMessage(`{}`)}}},
		{Content: "recovered"},
	}}
	reg := tools.NewMemRegistry()
	reg.Register(&scriptedTool{name: "flaky", err: errors.New("boom")})
	a := New(llmc, reg, 8, 5*time.Second)
	var events []Event
	if err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "go"}}}, collect(&events)); err != nil {
		t.Fatalf("Run error (should not bubble tool error): %v", err)
	}
	if events[len(events)-1].Final == nil || events[len(events)-1].Final.Text != "recovered" {
		t.Errorf("expected recovered final, got %+v", events[len(events)-1])
	}
}

func TestAgent_MaxStepsCap(t *testing.T) {
	// Model never stops calling a tool.
	loopTool := llm.ToolCall{ID: "c", Name: "echo", Args: json.RawMessage(`{}`)}
	llmc := &fakeLLM{responses: []llm.ChatResponse{
		{ToolCalls: []llm.ToolCall{loopTool}},
		{ToolCalls: []llm.ToolCall{loopTool}},
		{ToolCalls: []llm.ToolCall{loopTool}},
	}}
	reg := tools.NewMemRegistry()
	reg.Register(&scriptedTool{name: "echo", result: tools.Result{Content: map[string]any{"ok": true}}})
	a := New(llmc, reg, 3, 5*time.Second)
	var events []Event
	err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "go"}}}, collect(&events))
	if !errors.Is(err, ErrMaxSteps) {
		t.Fatalf("expected ErrMaxSteps, got %v", err)
	}
	if events[len(events)-1].Error == nil {
		t.Errorf("expected trailing error event, got %+v", events[len(events)-1])
	}
}

func TestAgent_LLMErrorBubbles(t *testing.T) {
	llmc := &fakeLLM{err: errors.New("ollama down")}
	a := New(llmc, tools.NewMemRegistry(), 8, 5*time.Second)
	err := a.Run(context.Background(), Turn{Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}}}, func(Event) {})
	if err == nil {
		t.Fatal("expected error from LLM failure")
	}
}
```

- [ ] **Step 3: Run the test, expect compile failure**

```bash
cd go/ai-service && go test ./internal/agent/...
```

Expected: compile error — `undefined: New`, `undefined: Turn`, `undefined: ErrMaxSteps`.

- [ ] **Step 4: Implement `agent.go`**

Contents of `go/ai-service/internal/agent/agent.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
)

// ErrMaxSteps is returned when the agent loop exceeds the configured step cap.
var ErrMaxSteps = errors.New("agent: max steps exceeded")

// Turn is one invocation of the agent — a user id plus the full conversation so far.
type Turn struct {
	UserID   string
	Messages []llm.Message
}

// Agent runs the LLM tool-calling loop.
type Agent struct {
	llm      llm.Client
	registry tools.Registry
	maxSteps int
	timeout  time.Duration
}

// New constructs an Agent.
func New(client llm.Client, registry tools.Registry, maxSteps int, timeout time.Duration) *Agent {
	return &Agent{llm: client, registry: registry, maxSteps: maxSteps, timeout: timeout}
}

// Run executes the loop. The emit callback receives every event in order.
// Infrastructure failures (LLM unreachable, ctx cancelled, max steps) are returned as errors.
// Tool-level failures are fed back into the conversation as tool results and do not return an error.
func (a *Agent) Run(ctx context.Context, turn Turn, emit func(Event)) error {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	messages := append([]llm.Message(nil), turn.Messages...)

	for step := 0; step < a.maxSteps; step++ {
		resp, err := a.llm.Chat(ctx, messages, a.registry.Schemas())
		if err != nil {
			emit(Event{Error: &ErrorEvent{Reason: err.Error()}})
			return fmt.Errorf("llm chat: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			emit(Event{Final: &FinalEvent{Text: resp.Content}})
			return nil
		}

		// Record the assistant's tool-call message in history.
		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, call := range resp.ToolCalls {
			emit(Event{ToolCall: &ToolCallEvent{Name: call.Name, Args: call.Args}})

			tool, ok := a.registry.Get(call.Name)
			if !ok {
				errMsg := "unknown tool: " + call.Name
				emit(Event{ToolError: &ToolErrorEvent{Name: call.Name, Error: errMsg}})
				msg, _ := llm.ToolResultMessage(call.ID, call.Name, map[string]string{"error": errMsg})
				messages = append(messages, msg)
				continue
			}

			result, toolErr := safeCall(ctx, tool, call.Args, turn.UserID)
			if toolErr != nil {
				emit(Event{ToolError: &ToolErrorEvent{Name: call.Name, Error: toolErr.Error()}})
				msg, _ := llm.ToolResultMessage(call.ID, call.Name, map[string]string{"error": toolErr.Error()})
				messages = append(messages, msg)
				continue
			}

			emit(Event{ToolResult: &ToolResultEvent{Name: call.Name, Display: result.Display}})
			msg, err := llm.ToolResultMessage(call.ID, call.Name, result.Content)
			if err != nil {
				// Unserializable tool result is a programmer bug in the tool — feed back an error.
				errMsg := "tool result not serializable: " + err.Error()
				emit(Event{ToolError: &ToolErrorEvent{Name: call.Name, Error: errMsg}})
				msg2, _ := llm.ToolResultMessage(call.ID, call.Name, map[string]string{"error": errMsg})
				messages = append(messages, msg2)
				continue
			}
			messages = append(messages, msg)
		}
	}

	emit(Event{Error: &ErrorEvent{Reason: ErrMaxSteps.Error()}})
	return ErrMaxSteps
}

// safeCall invokes a tool with a deferred recover so a panicking tool becomes an error.
func safeCall(ctx context.Context, t tools.Tool, args json.RawMessage, userID string) (result tools.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %q panicked: %v", t.Name(), r)
		}
	}()
	return t.Call(ctx, args, userID)
}
```

- [ ] **Step 5: Run the tests, expect pass**

```bash
cd go/ai-service && go test ./internal/agent/... -v
```

Expected: `PASS` for all six agent tests.

- [ ] **Step 6: Commit**

```bash
git add go/ai-service/internal/agent/
git commit -m "feat(ai-service): implement agent loop with step cap and tool error recovery"
```

---

## Task 6: Implement the ecommerce HTTP client and `get_product` tool

**Files:**
- Create: `go/ai-service/internal/tools/clients/ecommerce.go`
- Create: `go/ai-service/internal/tools/clients/ecommerce_test.go`
- Create: `go/ai-service/internal/tools/catalog.go`
- Create: `go/ai-service/internal/tools/catalog_test.go`

- [ ] **Step 1: Write failing test for the ecommerce client**

Contents of `go/ai-service/internal/tools/clients/ecommerce_test.go`:

```go
package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEcommerceClient_GetProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/abc-123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"abc-123","name":"Waterproof Jacket","price":129.99,"stock":4}`))
	}))
	defer server.Close()

	c := NewEcommerceClient(server.URL)
	p, err := c.GetProduct(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if p.ID != "abc-123" || p.Name != "Waterproof Jacket" || p.Price != 129.99 || p.Stock != 4 {
		t.Errorf("unexpected product: %+v", p)
	}
}

func TestEcommerceClient_GetProduct_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := NewEcommerceClient(server.URL).GetProduct(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error on 404")
	}
}

func TestEcommerceClient_ListProducts_TextSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "jacket" {
			t.Fatalf("expected q=jacket, got %q", r.URL.Query().Get("q"))
		}
		_, _ = w.Write([]byte(`[
			{"id":"p1","name":"Waterproof Jacket","price":129.99,"stock":4},
			{"id":"p2","name":"Rain Jacket","price":89.00,"stock":10}
		]`))
	}))
	defer server.Close()

	c := NewEcommerceClient(server.URL)
	ps, err := c.ListProducts(context.Background(), "jacket", 10)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(ps) != 2 {
		t.Fatalf("expected 2 results, got %d", len(ps))
	}
	if ps[0].Name != "Waterproof Jacket" {
		t.Errorf("first product wrong: %+v", ps[0])
	}
}
```

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/clients/...
```

Expected: `undefined: NewEcommerceClient`.

- [ ] **Step 3: Implement `ecommerce.go`**

Contents of `go/ai-service/internal/tools/clients/ecommerce.go`:

```go
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Product is the subset of ecommerce-service's product representation that ai-service needs.
type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// EcommerceClient is a typed HTTP client for ecommerce-service.
type EcommerceClient struct {
	baseURL string
	http    *http.Client
}

func NewEcommerceClient(baseURL string) *EcommerceClient {
	return &EcommerceClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *EcommerceClient) GetProduct(ctx context.Context, id string) (Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/products/"+url.PathEscape(id), nil)
	if err != nil {
		return Product{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Product{}, fmt.Errorf("get product: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		return Product{}, fmt.Errorf("get product %s: status %d: %s", id, resp.StatusCode, string(payload))
	}
	var p Product
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return Product{}, fmt.Errorf("decode product: %w", err)
	}
	return p, nil
}

// ListProducts does a text search via ecommerce-service. Limit is capped by the caller.
func (c *EcommerceClient) ListProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	u, err := url.Parse(c.baseURL + "/products")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if query != "" {
		q.Set("q", query)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list products: status %d: %s", resp.StatusCode, string(payload))
	}
	var out []Product
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode product list: %w", err)
	}
	return out, nil
}
```

- [ ] **Step 4: Run ecommerce client tests, expect pass**

```bash
cd go/ai-service && go test ./internal/tools/clients/... -v
```

Expected: `PASS` for all three client tests.

> **Note on the `q` parameter:** the existing `ecommerce-service` `GET /products` handler may or may not already support a `q` query parameter. Before running the end-to-end Task 10, open `go/ecommerce-service/internal/handler/product.go` and confirm. If it doesn't, add a case-insensitive `WHERE name ILIKE '%' || $1 || '%'` filter to `ProductRepository.List` and thread it through the handler. Adding that filter (plus a test in `ecommerce-service`) is in scope for this plan — it's a ~10-line change and it's the data path the agent needs.

- [ ] **Step 5: Write failing test for the catalog tools**

Contents of `go/ai-service/internal/tools/catalog_test.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

// fakeEcommerce is a stand-in for *clients.EcommerceClient that satisfies ecommerceAPI.
type fakeEcommerce struct {
	products map[string]clients.Product
	listOut  []clients.Product
	listErr  error
}

func (f *fakeEcommerce) GetProduct(ctx context.Context, id string) (clients.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return clients.Product{}, errors.New("not found")
	}
	return p, nil
}

func (f *fakeEcommerce) ListProducts(ctx context.Context, query string, limit int) ([]clients.Product, error) {
	return f.listOut, f.listErr
}

func TestGetProductTool_Success(t *testing.T) {
	fake := &fakeEcommerce{products: map[string]clients.Product{
		"p1": {ID: "p1", Name: "Waterproof Jacket", Price: 129.99, Stock: 4},
	}}
	tool := NewGetProductTool(fake)

	args := json.RawMessage(`{"product_id":"p1"}`)
	res, err := tool.Call(context.Background(), args, "")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	m, ok := res.Content.(map[string]any)
	if !ok {
		t.Fatalf("expected map content, got %T", res.Content)
	}
	if m["id"] != "p1" || m["name"] != "Waterproof Jacket" {
		t.Errorf("bad content: %+v", m)
	}
}

func TestGetProductTool_MissingArg(t *testing.T) {
	tool := NewGetProductTool(&fakeEcommerce{products: map[string]clients.Product{}})
	_, err := tool.Call(context.Background(), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error for missing product_id")
	}
}

func TestSearchProductsTool_BoundsAndFilters(t *testing.T) {
	fake := &fakeEcommerce{listOut: []clients.Product{
		{ID: "p1", Name: "Waterproof Jacket", Price: 129.99, Stock: 4},
		{ID: "p2", Name: "Rain Jacket", Price: 89.00, Stock: 10},
		{ID: "p3", Name: "Expensive Jacket", Price: 500.00, Stock: 1},
	}}
	tool := NewSearchProductsTool(fake)

	args := json.RawMessage(`{"query":"jacket","max_price":150}`)
	res, err := tool.Call(context.Background(), args, "")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	items, ok := res.Content.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map content, got %T", res.Content)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 filtered results, got %d", len(items))
	}
	for _, it := range items {
		price := it["price"].(float64)
		if price > 150 {
			t.Errorf("result above max_price: %+v", it)
		}
	}
}

func TestSearchProductsTool_MissingQuery(t *testing.T) {
	tool := NewSearchProductsTool(&fakeEcommerce{})
	_, err := tool.Call(context.Background(), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error for missing query")
	}
}

func TestCheckInventoryTool(t *testing.T) {
	fake := &fakeEcommerce{products: map[string]clients.Product{
		"p1": {ID: "p1", Name: "Waterproof Jacket", Price: 129.99, Stock: 4},
		"p2": {ID: "p2", Name: "Rain Jacket", Price: 89.00, Stock: 0},
	}}
	tool := NewCheckInventoryTool(fake)

	res, err := tool.Call(context.Background(), json.RawMessage(`{"product_id":"p1"}`), "")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	m := res.Content.(map[string]any)
	if m["in_stock"] != true || m["stock"].(int) != 4 {
		t.Errorf("bad content: %+v", m)
	}

	res, err = tool.Call(context.Background(), json.RawMessage(`{"product_id":"p2"}`), "")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if res.Content.(map[string]any)["in_stock"] != false {
		t.Errorf("expected out of stock: %+v", res.Content)
	}
}
```

- [ ] **Step 6: Run test, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/...
```

Expected: `undefined: NewGetProductTool`, etc.

- [ ] **Step 7: Implement `catalog.go`**

Contents of `go/ai-service/internal/tools/catalog.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

// ecommerceAPI is the subset of the ecommerce HTTP client the catalog tools use.
// Kept as an interface so tests can swap in a fake.
type ecommerceAPI interface {
	GetProduct(ctx context.Context, id string) (clients.Product, error)
	ListProducts(ctx context.Context, query string, limit int) ([]clients.Product, error)
}

// -------- get_product --------

type getProductTool struct {
	api ecommerceAPI
}

func NewGetProductTool(api ecommerceAPI) Tool { return &getProductTool{api: api} }

func (t *getProductTool) Name() string        { return "get_product" }
func (t *getProductTool) Description() string { return "Fetch the full details of one product by id." }
func (t *getProductTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"product_id":{"type":"string","description":"Opaque product id."}
		},
		"required":["product_id"]
	}`)
}

type getProductArgs struct {
	ProductID string `json:"product_id"`
}

func (t *getProductTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	var a getProductArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("get_product: bad args: %w", err)
	}
	if a.ProductID == "" {
		return Result{}, errors.New("get_product: product_id is required")
	}
	p, err := t.api.GetProduct(ctx, a.ProductID)
	if err != nil {
		return Result{}, fmt.Errorf("get_product: %w", err)
	}
	return Result{
		Content: map[string]any{"id": p.ID, "name": p.Name, "price": p.Price, "stock": p.Stock},
		Display: map[string]any{"kind": "product_card", "product": p},
	}, nil
}

// -------- search_products --------

type searchProductsTool struct {
	api ecommerceAPI
}

func NewSearchProductsTool(api ecommerceAPI) Tool { return &searchProductsTool{api: api} }

func (t *searchProductsTool) Name() string { return "search_products" }
func (t *searchProductsTool) Description() string {
	return "Search the product catalog by free-text query. Optional max_price filter. Returns at most 10 results."
}
func (t *searchProductsTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"query":{"type":"string","description":"Free-text product query."},
			"max_price":{"type":"number","description":"Optional upper bound on price."},
			"limit":{"type":"integer","description":"Max results to return (cap 10)."}
		},
		"required":["query"]
	}`)
}

type searchArgs struct {
	Query    string  `json:"query"`
	MaxPrice float64 `json:"max_price"`
	Limit    int     `json:"limit"`
}

const maxSearchResults = 10

func (t *searchProductsTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	var a searchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("search_products: bad args: %w", err)
	}
	if a.Query == "" {
		return Result{}, errors.New("search_products: query is required")
	}
	limit := a.Limit
	if limit <= 0 || limit > maxSearchResults {
		limit = maxSearchResults
	}

	prods, err := t.api.ListProducts(ctx, a.Query, limit)
	if err != nil {
		return Result{}, fmt.Errorf("search_products: %w", err)
	}

	out := make([]map[string]any, 0, len(prods))
	for _, p := range prods {
		if a.MaxPrice > 0 && p.Price > a.MaxPrice {
			continue
		}
		out = append(out, map[string]any{
			"id": p.ID, "name": p.Name, "price": p.Price, "stock": p.Stock,
		})
		if len(out) >= limit {
			break
		}
	}
	return Result{
		Content: out,
		Display: map[string]any{"kind": "product_list", "products": out},
	}, nil
}

// -------- check_inventory --------

type checkInventoryTool struct {
	api ecommerceAPI
}

func NewCheckInventoryTool(api ecommerceAPI) Tool { return &checkInventoryTool{api: api} }

func (t *checkInventoryTool) Name() string { return "check_inventory" }
func (t *checkInventoryTool) Description() string {
	return "Check whether a product is in stock. Returns stock count and a boolean."
}
func (t *checkInventoryTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"product_id":{"type":"string"}
		},
		"required":["product_id"]
	}`)
}

func (t *checkInventoryTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	var a getProductArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("check_inventory: bad args: %w", err)
	}
	if a.ProductID == "" {
		return Result{}, errors.New("check_inventory: product_id is required")
	}
	p, err := t.api.GetProduct(ctx, a.ProductID)
	if err != nil {
		return Result{}, fmt.Errorf("check_inventory: %w", err)
	}
	content := map[string]any{
		"product_id": p.ID,
		"stock":      p.Stock,
		"in_stock":   p.Stock > 0,
	}
	return Result{
		Content: content,
		Display: map[string]any{"kind": "inventory", "product_id": p.ID, "stock": p.Stock, "in_stock": p.Stock > 0},
	}, nil
}
```

- [ ] **Step 8: Run all tool tests, expect pass**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

Expected: `PASS` for every registry + catalog test.

- [ ] **Step 9: Commit**

```bash
git add go/ai-service/internal/tools/
git commit -m "feat(ai-service): add ecommerce client and catalog tools (search/get/inventory)"
```

---

## Task 7: (Conditional) Add text search to `ecommerce-service` `GET /products` if missing

**Skip this task entirely if `ecommerce-service`'s `GET /products` already supports a `q` query parameter.**

**Files (only if needed):**
- Modify: `go/ecommerce-service/internal/repository/product.go` (or equivalent)
- Modify: `go/ecommerce-service/internal/handler/product.go`
- Modify: `go/ecommerce-service/internal/repository/product_test.go` (or add one)

- [ ] **Step 1: Check whether `q` is already supported**

```bash
grep -n "\"q\"" go/ecommerce-service/internal/handler/product.go
grep -n "ILIKE" go/ecommerce-service/internal/repository/product.go
```

If either grep finds a match for a product-name text search, **skip to Task 8**.

- [ ] **Step 2: Add a failing repository test** asserting `List(ctx, "jack", 10)` returns products whose names match case-insensitively.

Show the test inline in the diff for clarity, then run:

```bash
cd go/ecommerce-service && go test ./internal/repository/... -run TestProductRepository_List_TextSearch
```

Expected: FAIL (no filter yet).

- [ ] **Step 3: Implement the filter**

In `ProductRepository.List`, change the query to accept an optional `query string` parameter and add `WHERE name ILIKE '%' || $1 || '%'` when non-empty. Thread `c.Query("q")` through the handler and pass it down.

- [ ] **Step 4: Run the repo test and the handler tests**

```bash
cd go/ecommerce-service && go test ./internal/repository/... ./internal/handler/...
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add go/ecommerce-service/internal/
git commit -m "feat(ecommerce): add name text search to GET /products?q="
```

---

## Task 8: `POST /chat` SSE handler with a fake agent

**Files:**
- Create: `go/ai-service/internal/http/chat.go`
- Create: `go/ai-service/internal/http/chat_test.go`
- Create: `go/ai-service/internal/http/health.go`

- [ ] **Step 1: Write the failing handler test**

Contents of `go/ai-service/internal/http/chat_test.go`:

```go
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/agent"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
)

type fakeRunner struct {
	events []agent.Event
	err    error
}

func (f *fakeRunner) Run(ctx context.Context, turn agent.Turn, emit func(agent.Event)) error {
	for _, e := range f.events {
		emit(e)
	}
	return f.err
}

func TestChatHandler_StreamsEventsAsSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	runner := &fakeRunner{events: []agent.Event{
		{ToolCall: &agent.ToolCallEvent{Name: "search_products", Args: json.RawMessage(`{"query":"jacket"}`)}},
		{ToolResult: &agent.ToolResultEvent{Name: "search_products", Display: map[string]any{"kind": "product_list"}}},
		{Final: &agent.FinalEvent{Text: "Here are some jackets."}},
	}}
	r := gin.New()
	RegisterChatRoutes(r, runner)

	body := strings.NewReader(`{"messages":[{"role":"user","content":"find a jacket"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/chat", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, expected text/event-stream", ct)
	}
	out := w.Body.String()
	for _, want := range []string{
		"event: tool_call",
		`"name":"search_products"`,
		"event: tool_result",
		"event: final",
		`"text":"Here are some jackets."`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("response missing %q:\n%s", want, out)
		}
	}
}

func TestChatHandler_BadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterChatRoutes(r, &fakeRunner{})
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

var _ = llm.RoleUser // keep import for future tests
```

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/http/...
```

Expected: `undefined: RegisterChatRoutes`.

- [ ] **Step 3: Implement `chat.go`**

Contents of `go/ai-service/internal/http/chat.go`:

```go
package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/agent"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
)

// Runner is the subset of *agent.Agent the HTTP handler needs.
type Runner interface {
	Run(ctx context.Context, turn agent.Turn, emit func(agent.Event)) error
}

type chatRequest struct {
	Messages  []llm.Message `json:"messages"`
	SessionID string        `json:"session_id,omitempty"`
}

const maxUserMessageBytes = 4000

// RegisterChatRoutes wires POST /chat onto r.
func RegisterChatRoutes(r *gin.Engine, runner Runner) {
	r.POST("/chat", func(c *gin.Context) {
		var req chatRequest
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxUserMessageBytes*4))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		if len(req.Messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages required"})
			return
		}
		// Last user message length guard (Plan 1 keeps it simple).
		for _, m := range req.Messages {
			if m.Role == llm.RoleUser && len(m.Content) > maxUserMessageBytes {
				c.JSON(http.StatusBadRequest, gin.H{"error": "message too long"})
				return
			}
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		c.Writer.WriteHeader(http.StatusOK)
		flusher, _ := c.Writer.(http.Flusher)

		emit := func(e agent.Event) {
			name, payload := eventName(e)
			data, _ := json.Marshal(payload)
			_, _ = c.Writer.WriteString("event: " + name + "\n")
			_, _ = c.Writer.WriteString("data: " + string(data) + "\n\n")
			if flusher != nil {
				flusher.Flush()
			}
		}

		turn := agent.Turn{UserID: "", Messages: req.Messages}
		if err := runner.Run(c.Request.Context(), turn, emit); err != nil {
			emit(agent.Event{Error: &agent.ErrorEvent{Reason: err.Error()}})
		}
	})
}

func eventName(e agent.Event) (string, any) {
	switch {
	case e.ToolCall != nil:
		return "tool_call", e.ToolCall
	case e.ToolResult != nil:
		return "tool_result", e.ToolResult
	case e.ToolError != nil:
		return "tool_error", e.ToolError
	case e.Final != nil:
		return "final", e.Final
	case e.Error != nil:
		return "error", e.Error
	default:
		return "unknown", struct{}{}
	}
}
```

- [ ] **Step 4: Implement `health.go`**

Contents of `go/ai-service/internal/http/health.go`:

```go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadyCheck is a single dependency probe. Returns nil when healthy.
type ReadyCheck func() error

// RegisterHealthRoutes adds GET /health (liveness) and GET /ready (dependency probes).
func RegisterHealthRoutes(r *gin.Engine, checks map[string]ReadyCheck) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		results := map[string]string{}
		allOK := true
		for name, fn := range checks {
			if err := fn(); err != nil {
				results[name] = err.Error()
				allOK = false
			} else {
				results[name] = "ok"
			}
		}
		status := http.StatusOK
		if !allOK {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"checks": results})
	})
}
```

- [ ] **Step 5: Run the http tests, expect pass**

```bash
cd go/ai-service && go test ./internal/http/... -v
```

Expected: `PASS` for `TestChatHandler_StreamsEventsAsSSE` and `TestChatHandler_BadBody`.

- [ ] **Step 6: Commit**

```bash
git add go/ai-service/internal/http/
git commit -m "feat(ai-service): add POST /chat SSE handler and health/ready endpoints"
```

---

## Task 9: Wire everything together in `cmd/server/main.go`

**Files:**
- Modify: `go/ai-service/cmd/server/main.go`

- [ ] **Step 1: Replace `main.go` with the full wiring**

Full contents of `go/ai-service/cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/agent"
	apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

func main() {
	port := getenv("PORT", "8093")
	ollamaURL := getenv("OLLAMA_URL", "http://ollama:11434")
	ollamaModel := getenv("OLLAMA_MODEL", "qwen2.5:14b")
	ecommerceURL := getenv("ECOMMERCE_URL", "http://ecommerce-service:8092")

	// LLM client
	llmc := llm.NewOllamaClient(ollamaURL, ollamaModel)

	// Tool registry
	ecomClient := clients.NewEcommerceClient(ecommerceURL)
	registry := tools.NewMemRegistry()
	registry.Register(tools.NewSearchProductsTool(ecomClient))
	registry.Register(tools.NewGetProductTool(ecomClient))
	registry.Register(tools.NewCheckInventoryTool(ecomClient))

	// Agent
	a := agent.New(llmc, registry, 8, 30*time.Second)

	// HTTP
	router := gin.New()
	router.Use(gin.Recovery())

	apphttp.RegisterHealthRoutes(router, map[string]apphttp.ReadyCheck{
		"ollama": func() error {
			req, _ := http.NewRequest(http.MethodGet, ollamaURL+"/api/tags", nil)
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		},
		"ecommerce": func() error {
			req, _ := http.NewRequest(http.MethodGet, ecommerceURL+"/health", nil)
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		},
	})
	apphttp.RegisterChatRoutes(router, a)

	srv := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		slog.Info("ai-service starting",
			"port", port,
			"ollama_url", ollamaURL,
			"ollama_model", ollamaModel,
			"ecommerce_url", ecommerceURL,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 2: Verify `*agent.Agent` satisfies `apphttp.Runner`**

```bash
cd go/ai-service && go build ./...
```

Expected: clean build. If there's a compile error about `*agent.Agent` not implementing `apphttp.Runner`, the `Run` signatures have drifted — reconcile.

- [ ] **Step 3: Run all tests**

```bash
cd go/ai-service && go test ./... -v
```

Expected: all packages PASS.

- [ ] **Step 4: Commit**

```bash
git add go/ai-service/cmd/server/main.go
git commit -m "feat(ai-service): wire agent, tools, and HTTP handlers in main"
```

---

## Task 10: Dockerfile, docker-compose integration, and end-to-end curl check

**Files:**
- Create: `go/ai-service/Dockerfile`
- Modify: `docker-compose.yml`
- Modify: `Makefile`

- [ ] **Step 1: Write the Dockerfile**

Contents of `go/ai-service/Dockerfile`:

```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ai-service ./cmd/server

FROM alpine:3.19

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 1001 appuser

COPY --from=builder /ai-service /ai-service

USER appuser

EXPOSE 8093
ENTRYPOINT ["/ai-service"]
```

- [ ] **Step 2: Add `ai-service` to `docker-compose.yml`**

Open `docker-compose.yml` and add a new service entry. Exact shape (add alongside the other Go services, matching their style):

```yaml
  ai-service:
    build:
      context: ./go/ai-service
      dockerfile: Dockerfile
    container_name: ai-service
    ports:
      - "8093:8093"
    environment:
      PORT: "8093"
      OLLAMA_URL: "http://host.docker.internal:11434"
      OLLAMA_MODEL: "qwen2.5:14b"
      ECOMMERCE_URL: "http://ecommerce-service:8092"
    depends_on:
      - ecommerce-service
    restart: unless-stopped
```

If `host.docker.internal` is not how Ollama is reached elsewhere in this compose file, mirror whatever pattern the existing Python chat/debug services use to reach Ollama on the host.

- [ ] **Step 3: Add `preflight-ai-service` to the Makefile**

Append to `Makefile`:

```makefile
.PHONY: preflight-ai-service
preflight-ai-service:
	cd go/ai-service && go vet ./...
	cd go/ai-service && go test ./... -count=1
```

And extend `preflight-go` (or the Go preflight target) to depend on `preflight-ai-service`.

- [ ] **Step 4: Run the preflight locally**

```bash
make preflight-ai-service
```

Expected: vet clean, all tests pass.

- [ ] **Step 5: Bring the stack up and run a manual end-to-end check**

```bash
colima start || true
docker compose up -d --build ai-service ecommerce-service
sleep 5
curl -s localhost:8093/health
curl -s localhost:8093/ready
```

Expected: `/health` returns `{"status":"ok"}`. `/ready` returns `200` with `ollama` and `ecommerce` checks either `"ok"` or an honest error (acceptable if Ollama isn't running locally; the service is still up).

- [ ] **Step 6: Manual SSE smoke test against `/chat`**

```bash
curl -N -s -X POST localhost:8093/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"find me a jacket under 150 dollars"}]}'
```

Expected: a stream of SSE events — at minimum a `final` event, and (if Ollama + a seeded ecommerce product set are reachable) at least one `tool_call` event for `search_products` followed by a `tool_result` event and a `final` event.

If Ollama is not reachable from Mac in this compose layout, this smoke test is allowed to return an `error` event sourced from `ollama chat: ...`. That's still proof the HTTP, routing, and agent wiring are correct. The full real-Ollama end-to-end path is re-verified in Plan 5 against Minikube.

- [ ] **Step 7: Tear down and commit**

```bash
docker compose down
git add go/ai-service/Dockerfile docker-compose.yml Makefile
git commit -m "feat(ai-service): containerize and add to docker-compose with preflight target"
```

---

## Done criteria for Plan 1

- `go test ./go/ai-service/...` runs fully offline and passes.
- `make preflight-ai-service` passes.
- `docker compose up ai-service` brings up a healthy service that serves `/health`, `/ready`, and `/chat`.
- With Ollama reachable and at least one matching product in `ecommerce-service`, a `curl` POST to `/chat` returns an SSE stream that includes a `tool_call` for `search_products`, a `tool_result`, and a `final` event.
- No auth, no user-scoped tools, no Redis, no metrics, no frontend — those are Plans 2–5.
