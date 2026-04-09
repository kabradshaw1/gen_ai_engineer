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
