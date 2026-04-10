# Agent Harness in Go — `go/ai-service`

## Abstract

`go/ai-service` is a Go microservice that wraps an LLM agent loop around the existing ecommerce backend.
A user opens the **AI Shopping Assistant** drawer in the frontend, types a message, and the service
orchestrates one or more Ollama (Qwen 2.5) tool calls against ecommerce-service, streams events back
over SSE, and returns a final natural-language answer. The service exists because the portfolio needed
a concrete example of **agent design in Go** — typed tool dispatch, structured observability, and
deliberate API boundaries — as opposed to adding yet more RAG on the Python side. The full rationale
for the pivot is in `docs/adr/rag-reevaluation-2026-04.md`: RAG is commodity in 2026, agentic tool-use
is the scarce skill, and Go hiring managers should see Go at the center of the AI work.

## Architecture at a Glance

```
Browser
  └─ Frontend drawer (Next.js)
       └─ POST /go-ai/chat   (SSE response)
            └─ ai-service (Go)
                 ├─ OllamaClient  ──►  Ollama /api/chat  (Qwen 2.5 14B, Windows GPU)
                 ├─ tools.Registry
                 │    ├─ catalog tools  ──►  ecommerce-service  (products, inventory)
                 │    ├─ cart / order tools  ──►  ecommerce-service  (authed)
                 │    └─ returns tool   ──►  ecommerce-service  (authed)
                 ├─ cache.Cache  ──►  Redis  (TTL-backed result cache)
                 └─ metrics.Recorder  ──►  Prometheus  →  Grafana
```

The HTTP handler validates the JWT, extracts `userID`, builds a `Turn` struct, calls `agent.Run`,
and forwards each `Event` to the SSE stream. The agent loop is unaware of HTTP; the only I/O it
performs is calling `llm.Client.Chat` and invoking tools through the registry.

Traffic enters through the shared NGINX Ingress at `api.kylebradshaw.dev/go-ai/*` and is forwarded
to the `ai-service` pod in the `go-ecommerce` namespace. The Ollama server lives on the Windows host
and is reached from Kubernetes via an ExternalName service — the same pattern used by every other
service in the cluster that needs GPU inference.

## The Agent Loop

The `Run` method in `internal/agent/agent.go` is the core of the service. Every agent turn passes
through this function exactly once. The loop is intentionally synchronous and linear: call the LLM,
check for tool calls, dispatch tools, append results, repeat — or stop.

```go
// From go/ai-service/internal/agent/agent.go

// Run executes the loop. The emit callback receives every event in order.
// Infrastructure failures (LLM unreachable, ctx cancelled, max steps) are returned as errors.
// Tool-level failures are fed back into the conversation as tool results and do not return an error.
func (a *Agent) Run(ctx context.Context, turn Turn, emit func(Event)) error {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	turnID := uuid.NewString()
	startTime := time.Now()
	stepsCompleted := 0
	turn.Messages = guardrails.TruncateHistory(turn.Messages, guardrails.DefaultMaxHistory)
	messages := append([]llm.Message(nil), turn.Messages...)
	var toolsCalled []string

	for step := 0; step < a.maxSteps; step++ {
		resp, err := a.llm.Chat(ctx, messages, a.registry.Schemas())
		if err != nil {
			emit(Event{Error: &ErrorEvent{Reason: err.Error()}})
			a.rec.RecordTurn("error", stepsCompleted, time.Since(startTime))
			// ... structured log ...
			return fmt.Errorf("llm chat: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			outcome := "final"
			if guardrails.IsRefusal(resp.Content) {
				outcome = "refused"
			}
			a.rec.RecordTurn(outcome, stepsCompleted+1, time.Since(startTime))
			// ... structured log ...
			emit(Event{Final: &FinalEvent{Text: resp.Content}})
			return nil
		}

		stepsCompleted++
		messages = append(messages, llm.Message{
			Role: llm.RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls,
		})

		for _, call := range resp.ToolCalls {
			emit(Event{ToolCall: &ToolCallEvent{Name: call.Name, Args: call.Args}})
			tool, ok := a.registry.Get(call.Name)
			if !ok {
				errMsg := "unknown tool: " + call.Name
				emit(Event{ToolError: &ToolErrorEvent{Name: call.Name, Error: errMsg}})
				msg, _ := llm.ToolResultMessage(call.ID, call.Name, map[string]string{"error": errMsg})
				messages = append(messages, msg)
				a.rec.RecordTool(call.Name, "unknown", 0)
				continue
			}
			toolStart := time.Now()
			result, toolErr := safeCall(ctx, tool, call.Args, turn.UserID)
			// ... record metrics, emit event, append tool result message ...
		}
	}

	emit(Event{Error: &ErrorEvent{Reason: ErrMaxSteps.Error()}})
	a.rec.RecordTurn("max_steps", a.maxSteps, time.Since(startTime))
	return ErrMaxSteps
}
```

### Design decisions baked into this loop

1. **Tool errors become tool results, not loop errors.** A tool that returns an error gets its error
   message serialized into a `tool` role message and fed back into the conversation. The LLM can
   observe the failure and try a different tool or explain the problem. Only infrastructure failures
   (LLM unreachable, `ctx` cancelled) return as Go errors and abort the loop.

2. **Hard step cap + wall-clock timeout — two independent bounds.** `maxSteps` caps the number of
   LLM calls regardless of how fast they are. `timeout` caps total wall-clock time regardless of
   how few steps have run. Either can fire first. Neither alone is sufficient: a slow model can burn
   the timeout before hitting the step cap; a fast model with a bad prompt can burn the step cap
   before the timeout fires.

3. **Sequential tool dispatch in v1.** The loop calls tools one at a time in the order the LLM
   requested them. This matches Ollama's sequential tool-call behavior and keeps the history clean
   without coordination. Parallel dispatch is a v2 concern.

4. **Registry passed in, not global.** The `Agent` holds a `tools.Registry` interface injected at
   construction. Tests swap in a fake registry with zero-dependency tools. The agent package never
   imports a concrete tool implementation.

5. **`emit` is a function, not a channel.** The caller decides what to do with events (write to SSE,
   collect into a slice for tests, discard). A channel would force the caller to drain concurrently.
   A callback keeps the loop sequential and the caller in control.

6. **`userID` is an explicit parameter on `Tool.Call`, not a context value.** The JWT itself IS
   carried on context (see Section 7), because it must cross package boundaries invisibly. But
   `userID` is extracted once at the HTTP boundary and passed explicitly so every tool's signature
   makes the auth dependency visible and tests can set it without constructing a fake context.

## The Tool Interface and Registry

Every capability the agent can invoke implements the four-method `Tool` interface in
`internal/tools/registry.go`. The interface is intentionally minimal — anything that satisfies
it can be registered, including a future MCP adapter.

```go
// From go/ai-service/internal/tools/registry.go

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
```

`Result` has two fields because the LLM and the frontend want different representations of the same
data. `Content` is compact, JSON-serializable, and goes into the `tool` role message the LLM reads
next (e.g., `{"id": "abc", "name": "Widget", "price": 2999, "stock": 14}`). `Display` is an
optional richer payload forwarded to the frontend as a `tool_result` SSE event — for example, a
`{"kind": "product_card", "product": {...}}` object the drawer renders as a card. Keeping them
separate avoids polluting the LLM's context with UI concerns.

**How a future MCP adapter slots in.** An MCP client adapter needs exactly one new file:

```go
// hypothetical: internal/tools/mcpadapter/mcp.go
type ClientTool struct {
    client *mcp.Client
    spec   mcp.ToolSpec
}

func (t *ClientTool) Name() string             { return t.spec.Name }
func (t *ClientTool) Description() string      { return t.spec.Description }
func (t *ClientTool) Schema() json.RawMessage  { return t.spec.InputSchema }
func (t *ClientTool) Call(ctx context.Context, args json.RawMessage, userID string) (tools.Result, error) {
    raw, err := t.client.CallTool(ctx, t.spec.Name, args)
    // ... wrap in tools.Result ...
}
```

No changes to `agent.go`, no changes to the registry, no changes to any existing tool. Registration
is just `registry.Register(&mcpadapter.ClientTool{...})`.

## Ollama Tool-Calling Contract

`OllamaClient.Chat` in `internal/llm/ollama.go` speaks Ollama's `/api/chat` endpoint using the
OpenAI-compatible tool-calling shape that Qwen 2.5 supports. `stream` is always `false` — the agent
loop needs the complete response (including any tool calls) before it can dispatch.

```go
// Request body sent to POST /api/chat
{
  "model": "qwen2.5:14b",
  "stream": false,
  "messages": [
    {"role": "system", "content": "You are a helpful shopping assistant..."},
    {"role": "user",   "content": "Do you have any wireless headphones under $100?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "search_products",
        "description": "Search the product catalog by free-text query...",
        "parameters": {
          "type": "object",
          "properties": {
            "query":     {"type": "string"},
            "max_price": {"type": "number"},
            "limit":     {"type": "integer"}
          },
          "required": ["query"]
        }
      }
    }
    // ... remaining 8 tools ...
  ]
}

// Response when the model wants to call a tool
{
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [
      {
        "function": {
          "name": "search_products",
          "arguments": {"query": "wireless headphones", "max_price": 100}
        }
      }
    ]
  },
  "done": true
}
```

The `OllamaClient` translates the raw Ollama shapes into the internal `llm.ChatResponse` type.
The agent loop never sees Ollama-specific structs — it depends only on `llm.Client`, the
one-method interface in `internal/llm/client.go`:

```go
// From go/ai-service/internal/llm/client.go
type Client interface {
    Chat(ctx context.Context, messages []Message, tools []ToolSchema) (ChatResponse, error)
}
```

This makes swapping Ollama for any other model provider (OpenAI, Anthropic, a local stub) a
one-file change with zero agent-loop edits.

## The Nine-Tool Catalog

| name | file | user-scoped | description |
|---|---|---|---|
| `search_products` | catalog.go | no | Free-text search via ecommerce-service; `max_price` filter is in dollars at the LLM surface, converted to cents before the API call |
| `get_product` | catalog.go | no | Fetch one product by ID |
| `check_inventory` | catalog.go | no | Stock count + boolean; reuses GetProduct under the hood |
| `list_orders` | orders.go | yes | Last 20 orders for the authenticated user |
| `get_order` | orders.go | yes | One order by ID; ownership checked server-side by ecommerce-service |
| `summarize_orders` | orders.go | yes | Sub-LLM call that summarizes recent order history in plain language |
| `view_cart` | cart.go | yes | Current cart contents |
| `add_to_cart` | cart.go | yes | Add a product + quantity to cart |
| `initiate_return` | returns.go | yes | Open a return request for an order item |

> **`place_order` / checkout is deliberately NOT a tool.** The agent can browse the catalog, manage
> the cart, review orders, and open returns — but it cannot move money. This is a design choice,
> not a missing feature. "I drew the boundary here and here is why" is a stronger interview answer
> than a flashier demo with no boundaries at all.

## Auth Model

The service shares the same HS256 JWT secret as `go/auth-service`. The chat handler validates the
token, extracts `userID`, and stores two things: the string `userID` on `Turn.UserID` (passed
explicitly through the agent loop to every tool), and the raw bearer token on the request
`context.Context` via `internal/jwtctx`.

Tools that call ecommerce-service authenticated endpoints retrieve the token from context via
`jwtctx.FromContext` and pass it in the `Authorization` header through `clients.EcommerceClient`.
This means ecommerce-service remains the sole authorization authority — ai-service never checks
resource ownership itself, it just forwards credentials.

```go
// From go/ai-service/internal/jwtctx/jwtctx.go

// WithJWT returns a new context that carries the user's bearer token.
func WithJWT(ctx context.Context, jwt string) context.Context {
	return context.WithValue(ctx, jwtKey, jwt)
}

// FromContext returns the bearer token attached by WithJWT, or "".
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(jwtKey).(string)
	return v
}
```

## Python vs. Go Agent Loop

The Python Debug Assistant in `services/debug/` has an agent loop that is structurally identical:
call the LLM, check for tool calls, dispatch, append results, repeat, enforce a step cap, propagate
a context timeout. The two implementations share the same conceptual shape.

**What Go makes easier:** Typed tool arguments are validated by `json.Unmarshal` into concrete
structs — a missing required field produces a compile-time-visible error path rather than a
KeyError at runtime. The `Tool` interface is statically checked by the compiler; a bad
implementation is a build error, not a test failure. The `emit` callback is a plain function
value with no `asyncio` event loop required — the loop stays synchronous and its execution order
is obvious. There is no GIL, no `await` at every I/O call, and no `asyncio.gather` needed even
when you want concurrency.

**What is harder in Go:** There is no native pattern matching on discriminated union types.
Agent events (`ToolCallEvent`, `ToolResultEvent`, `FinalEvent`, `ErrorEvent`) are represented
as a tagged struct (`agent.Event`) with optional pointer fields — the caller checks which pointer
is non-nil. Python's match/case or a union type makes event dispatch more readable. JSON shaping
for the LLM wire format is also more verbose in Go: each request/response shape requires an
explicit struct definition, whereas Python libraries like LangChain handle this generically.

## Operating It Like a System

- **Cache:** A `cache.Cache` interface with a `NopCache` fallback means caching is opt-in per
  tool. Catalog tools (`search_products`, `get_product`, `check_inventory`) are wrapped with
  `tools.Cached` at a 60-second TTL. Order tools use a shorter 10-second TTL (data changes
  more often). Write tools (`add_to_cart`, `initiate_return`) are never cached. Redis errors
  fall through to the NopCache so the service keeps working without Redis.

- **Metrics:** Six Prometheus metrics exposed at `/metrics`, visible on the existing Grafana
  `system-overview` dashboard under the "AI Service" row:
  - `ai_agent_turns_total{outcome}` — counter, labeled `final` / `refused` / `error` / `max_steps`
  - `ai_agent_steps_per_turn` — histogram, LLM calls per turn
  - `ai_agent_turn_duration_seconds` — histogram, end-to-end wall-clock
  - `ai_tool_calls_total{name,outcome}` — counter per tool
  - `ai_tool_duration_seconds{name}` — histogram per tool
  - `ai_cache_events_total{cache,event}` — counter, labeled `hit` / `miss` / `set` / `error`

- **Guardrails:** History is truncated to the last 20 messages before each turn
  (`guardrails.TruncateHistory`). Refusal detection inspects the final LLM response and tags
  the turn outcome as `refused` rather than `final`. A Redis token-bucket rate limiter enforces
  20 requests/minute/IP; Redis errors fail open so the service stays available.

- **Logging:** One structured JSON line per turn via `log/slog` with fields `turn_id`, `user_id`,
  `steps`, `tools_called`, `duration_ms`, `outcome`. Every outcome path (final, refused, error,
  max_steps) emits a log line before returning.

- **Evals:** A build-tagged (`//go:build eval`) package runs three offline cases against a
  mocked LLM: product search, order lookup, and a multi-step cart interaction. Triggered via
  `make preflight-ai-service-evals`. Real-LLM nightly evals are deferred.

## Prometheus Counter Declaration (example)

```go
// From go/ai-service/internal/metrics/metrics.go

var (
	TurnsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_agent_turns_total",
		Help: "Agent turns by outcome.",
	}, []string{"outcome"})

	StepsPerTurn = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ai_agent_steps_per_turn",
		Help:    "Number of LLM calls per turn.",
		Buckets: []float64{1, 2, 3, 4, 5, 6, 8, 10},
	})

	TurnDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ai_agent_turn_duration_seconds",
		Help:    "End-to-end agent turn duration.",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	ToolCallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_tool_calls_total",
		Help: "Tool invocations by name and outcome.",
	}, []string{"name", "outcome"})

	CacheEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_cache_events_total",
		Help: "Cache events by cache and event type.",
	}, []string{"cache", "event"})
)
```

## What's Deliberately Out of Scope

- **No `place_order` / checkout.** The agent assists — it does not transact. See the nine-tool
  catalog note above.
- **No embedding cache / semantic search.** All product discovery goes through ecommerce-service's
  text search endpoint. Qdrant is available in the cluster for future semantic ranking, but there
  is no embedding pipeline in this service yet.
- **No MCP adapter in v1.** The `Tool` interface is designed to admit one in a single new file
  (see Section 4). Building the MCP adapter itself is a follow-up.
- **No conversation persistence.** Each HTTP request is stateless. Message history lives entirely
  on the client (the frontend drawer). The service never writes session state.
- **No PII scrub / jailbreak / content moderation.** These are full projects in their own right.
  The refusal-detection guardrail is a lightweight label on an outcome that already happened, not
  a prevention layer.

## Check Your Understanding

1. **If the registry `Register`s two tools with the same name, which one wins and why?**
   `MemRegistry.Register` does `r.tools[t.Name()] = t` — a plain map assignment. The second
   registration silently overwrites the first. There is no error and no warning. This is a
   conscious simplicity choice for v1; a production registry would panic or error on duplicate
   names.

2. **What does the agent loop do when a tool returns an error? What does it do when the LLM call
   returns an error?**
   A tool error is serialized into a `{"error": "..."}` map, written into a `tool` role message,
   appended to the history, and the loop continues to the next tool call (or to the next LLM
   call). The loop does NOT return an error. An LLM call error emits an `ErrorEvent`, records
   the turn as `"error"`, and returns the Go error immediately — the loop aborts.

3. **Why is `userID` an explicit parameter on `Tool.Call` instead of being read from
   `context.Context`, when the JWT itself IS read from context?**
   `userID` is extracted once at the HTTP boundary and threaded explicitly so that every tool's
   type signature makes the auth dependency visible — you can read it in the interface definition
   without knowing how the context is shaped. Test code can pass any string. The JWT, by contrast,
   must cross several package boundaries (handler → agent loop → tool → HTTP client) invisibly;
   making it explicit would require every layer to accept and forward a string it does not
   otherwise care about. Context is the right carrier for that kind of ambient request-scoped value.

4. **What would it take to add an MCP adapter — name the files that would change and the files
   that would NOT change?**
   **New file:** `internal/tools/mcpadapter/mcp.go` — a struct implementing `tools.Tool` by
   proxying calls to an MCP server client.
   **Changed file:** `cmd/api/main.go` (or wherever tools are registered) — one additional
   `registry.Register(&mcpadapter.ClientTool{...})` call per MCP tool.
   **Unchanged:** `internal/agent/agent.go`, `internal/tools/registry.go`, all nine existing
   tool files, `internal/llm/`, `internal/metrics/`, `internal/jwtctx/`, all handler code.
