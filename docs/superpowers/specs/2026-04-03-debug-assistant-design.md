# Debug Assistant Design

**Date:** 2026-04-03
**Status:** Approved

## Summary

An agentic debug assistant that takes a bug description (and optional error output), searches a pre-indexed Python codebase, forms hypotheses, runs tools to verify them, and streams a step-by-step diagnosis to the user. Builds on the Document Q&A project by reusing embeddings, vector search, and chunking patterns while adding a custom agent loop with tool calling.

## Goals

- Demonstrate agentic AI concepts: tool use, planning loops, multi-step reasoning
- Build a custom agent loop from scratch (no framework agent abstractions)
- Reuse LangChain only for utilities (text splitting, embeddings, Qdrant integration)
- Extend the existing portfolio frontend with a real-time agent execution timeline
- Run locally against Ollama on the Windows PC (RTX 3090)

## Architecture

### New Service: `services/debug/`

```
services/debug/
├── app/
│   ├── main.py          # FastAPI app, SSE streaming endpoint
│   ├── config.py        # Ollama URL, Qdrant settings, model names
│   ├── indexer.py        # Walk repo, chunk code files, embed → Qdrant
│   ├── agent.py          # Agent loop: observe → think → act
│   ├── tools.py          # Tool definitions + execution
│   └── prompts.py        # System prompt, tool-use prompt templates
└── tests/
    ├── test_indexer.py
    ├── test_tools.py
    ├── test_agent.py
    └── test_prompts.py
```

Sits alongside `services/ingestion/` and `services/chat/`. Follows the same FastAPI patterns, Docker packaging, and CI integration.

### Model Change

Upgrade from Mistral 7B to **Qwen 2.5 14B** (`qwen2.5:14b`) for both the existing chat service and the new debug service. Reasons:

- 14B is significantly better at multi-step tool-calling reasoning than 7B
- Qwen 2.5 was specifically trained for function/tool calling
- Fits on RTX 3090 (24GB VRAM): ~11 GB for Qwen 14B + ~270 MB for nomic-embed-text
- Single model for both services avoids VRAM swap latency
- Set `OLLAMA_KEEP_ALIVE=-1` so the model stays loaded for instant portfolio responses

Embeddings remain **nomic-embed-text** (unchanged).

## Agent Loop

The core of the debug service. A custom observe-think-act loop:

1. Build prompt: system prompt + bug description + optional error output + conversation history
2. Call Ollama `/api/chat` with tool definitions
3. Parse response:
   - If tool call → execute the tool, append result to history, go to step 1
   - If text response → stream as final diagnosis, exit loop
4. Stream each step to the frontend as a typed SSE event

### Loop Guardrails

- **Max iterations:** 10 steps. Hard cap to prevent runaway loops with 14B models.
- **Max tool output:** Truncate results to keep context manageable (100 lines for file reads, 20 matches for grep, etc.)
- **Duplicate detection:** If the agent calls the same tool with identical arguments, inject a nudge to try a different approach.

### Conversation History Format

Standard multi-turn tool-use format for Ollama's `/api/chat`:

```
[system, user, assistant (tool_call), tool (result), assistant (tool_call), tool (result), ..., assistant (diagnosis)]
```

## Tools

Four tools for v1. Each has a name, description (LLM reads this), JSON schema parameters, and an execute function.

### Tool 1: `search_code`

Vector similarity search over the indexed codebase.

- **Input:** `{ "query": "PDF text extraction error handling" }`
- **Output:** Top 5 matching code chunks with file path, line numbers, and content
- **Implementation:** Embed query with nomic-embed-text → search Qdrant → return results
- **Reuses:** Same retriever pattern from the Document Q&A chat service

### Tool 2: `read_file`

Read a specific file or line range.

- **Input:** `{ "path": "app/pdf_parser.py", "start_line": 30, "end_line": 60 }`
- **Output:** File contents with line numbers (capped at 100 lines)
- **Why:** After vector search finds a chunk, the agent often needs surrounding context

### Tool 3: `grep`

Keyword/regex search across the project.

- **Input:** `{ "pattern": "raise HTTPException", "file_glob": "*.py" }`
- **Output:** Matching lines with file paths and line numbers (capped at 20 matches)
- **Why:** Vector search finds semantically similar code; grep finds exact strings (error messages, function names, imports). The agent needs both.

### Tool 4: `run_tests`

Run pytest on a specific file or test. Executes in the indexed project's root directory (stored as metadata on the Qdrant collection during indexing).

- **Input:** `{ "target": "tests/test_pdf_parser.py", "test_name": "test_empty_pdf" }`
- **Output:** Pass/fail status, failure message, stdout (truncated)
- **Why:** The agent can form a hypothesis and verify it by running relevant tests
- **Security:** Runs in a subprocess with a timeout (30s default). Only executes within the indexed project directory.

### Potential v2 Tools

- **Static analyzer** — run ruff and parse results
- **Git log/blame** — check recent changes to suspicious files

## Code Indexing

Before debugging, the codebase must be embedded into Qdrant.

### Process

1. User provides a path to a Python project
2. `indexer.py` walks the directory, skipping non-code files (`.pyc`, `__pycache__`, `.git`, `node_modules`, `venv`, etc.)
3. Python files are chunked using LangChain's `Language.PYTHON` splitter — splits on logical boundaries (functions, classes, top-level blocks) rather than arbitrary token counts
4. Each chunk gets metadata: `file_path`, `start_line`, `end_line`, `type` (function/class/module)
5. Chunks are embedded with nomic-embed-text and stored in a Qdrant collection named after the project

### Key Difference from Document Q&A

- Document Q&A chunks prose (paragraphs)
- Debug assistant chunks code (functions and classes)
- Code-aware chunking matters more because a function split in half is much harder to reason about

### Re-indexing

Drop and recreate the collection on re-index. Incremental indexing is a v2 concern.

### Scope

Python files only (`.py`) for v1. Architecture supports adding more languages by swapping the LangChain language splitter.

## API

The debug service runs on port **8003**.

### Endpoints

```
POST /index     — Index a Python project
  Body: { "path": "/path/to/project" }
  Response: { "collection": "project-name", "files_indexed": 12, "chunks": 87 }

POST /debug     — Start a debug session (SSE stream)
  Body: { "collection": "project-name", "description": "...", "error_output": "..." }
  Response: Server-Sent Events stream

GET  /health    — Health check
```

### SSE Event Types

```
event: thinking
data: {"step": 1, "content": "The error mentions pdf_parser.py..."}

event: tool_call
data: {"step": 2, "tool": "search_code", "args": {"query": "PDF text extraction"}}

event: tool_result
data: {"step": 2, "tool": "search_code", "result": "Found 3 matches...", "truncated": false}

event: thinking
data: {"step": 3, "content": "The parse_pdf function doesn't handle empty text..."}

event: tool_call
data: {"step": 4, "tool": "run_tests", "args": {"target": "tests/test_pdf_parser.py"}}

event: tool_result
data: {"step": 4, "tool": "run_tests", "result": "1 failed: test_empty_pdf..."}

event: diagnosis
data: {"content": "The bug is in parse_pdf()...", "files": ["app/pdf_parser.py"], "suggestion": "Add guard clause..."}

event: done
data: {}
```

## Frontend

### Routes

- `/ai` — Existing hub page. Add a second project section for the Debug Assistant with description, workflow diagram (Mermaid), tech stack, and "Try the Demo" link to `/ai/debug`
- `/ai/debug` — The debug tool interface

### `/ai/debug` Layout

**Left panel:** Input form
- Project path text field
- Bug description textarea
- Error output textarea (optional)
- "Debug" button

**Right panel:** Agent execution timeline (streams in real-time)

### New Components

- `DebugForm` — input form for bug description and project path
- `AgentTimeline` — step-by-step execution display
- `ToolCallCard` — collapsible card showing tool name, arguments, and result
- `DiagnosisBanner` — highlighted final result block

### Reused Patterns

- SSE streaming connection (same pattern as `ChatWindow`)
- shadcn/ui components (Card, Collapsible, Badge, Textarea, Button)
- MermaidDiagram component for the `/ai` page workflow diagram
- Layout shell and navigation

### Updates to `/ai` Page

- Add Debug Assistant project section below Document Q&A
- Update model reference from "Mistral 7B" to "Qwen 2.5 14B"

## Testing

### Unit Tests

- `test_indexer.py` — chunking logic, file filtering, metadata extraction
- `test_tools.py` — each tool executes correctly, handles edge cases (missing files, no matches, test failures)
- `test_agent.py` — mock Ollama responses, verify loop terminates, handles tool errors gracefully
- `test_prompts.py` — prompt templates render correctly

### E2E Tests

- Mocked Playwright test for `/ai/debug` — submit a bug description, verify the timeline renders steps and a diagnosis

## Deployment

- Add debug service to `docker-compose.yml` alongside ingestion and chat (port 8003)
- Add Cloudflare Tunnel route: `api-debug.kylebradshaw.dev` → Windows PC :8003
- Pull `qwen2.5:14b` on Windows machine via Ollama
- Update existing services' configs to reference the new model
- Set `OLLAMA_KEEP_ALIVE=-1` as system environment variable on Windows
- CI pipeline picks up the new service automatically (same patterns)

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Agent framework | Custom loop, LangChain for utilities only | Own the core learning, avoid volatile agent abstractions |
| Model | Qwen 2.5 14B | Best tool-calling at 14B size, fits on 3090 |
| Code chunking | LangChain Language.PYTHON splitter | Splits on logical boundaries (functions/classes) |
| Streaming | SSE with typed events | Same pattern as existing chat service, richer event types |
| Scope | Python-only for v1 | Keeps tooling simple, extensible to other languages later |
| Frontend | Extend existing Next.js app at /ai/debug | Portfolio hub pattern, reuse existing infrastructure |
