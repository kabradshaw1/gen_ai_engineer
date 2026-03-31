# Lesson Notebooks Design Spec

## Overview

7 Jupyter notebooks that guide Kyle through rebuilding both Python backend services (ingestion + chat) from scratch. Each lesson builds one module, explains why things are done the way they are, and fills in knowledge gaps for a developer coming from Go/TypeScript. By lesson 7, both services are fully rebuilt.

## Goals

- Rebuild the actual app code yourself, module by module
- Understand *why* each design decision was made, not just *what* the code does
- Learn what each package/library does and why it was chosen — build vocabulary to discuss these tools and design decisions with Claude agents and in interviews
- Bridge Go/TS knowledge to Python patterns — brief comparisons, not language tutorials
- Each notebook is self-contained and runnable independently
- Produce working code that is functionally equivalent to `services/ingestion/` and `services/chat/` (same behavior, may differ in variable names or minor details)

## Audience

Kyle — strong Go and TypeScript developer, experienced with Docker/K8s/web services, limited Python production experience. Familiar with Ollama. Knows what RAG is conceptually but hasn't built one.

## Notebook Structure

Every notebook follows this format:

1. **Intro** — What you're building, why it matters, how it fits in the app (2-3 paragraphs)
2. **Prerequisites** — pip installs, connectivity checks for Ollama/Qdrant where needed
3. **Package Introductions** — Before using a new package, explain what it is, what problem it solves, why it was chosen over alternatives, and the key APIs you'll use. For example: "PyPDF2 is a pure-Python PDF library. We're using it because it has no system dependencies (unlike pdfminer or poppler). The main API is `PdfReader` — you give it a file-like object and iterate `.pages`." This gives you the vocabulary to discuss these tools confidently.
4. **Go/TS Comparison** — Brief sidebar mapping the core concept to Go/TS equivalents (1-2 paragraphs, not a language tutorial)
5. **Build It** — Multiple code cells, each building one piece. Markdown cells between explain the *why* behind each decision. This is the bulk of the notebook.
6. **Experiment** — 2-3 "try this" cells where you tweak parameters and observe effects
7. **Check Your Understanding** — 2-3 reflection prompts ("In your own words, why does chunk overlap matter?")

## Claude Code Pitfall Callouts

Sprinkled throughout lessons where relevant (not a standalone lesson). These are short callout blocks warning about common mistakes AI assistants make:

- Generating sync code where async is required (Lesson 01)
- Using deprecated LangChain APIs (Lesson 02)
- Hardcoded model names instead of config-driven (Lesson 03)
- Incorrect Qdrant collection schemas (Lesson 04)
- Missing error handling on Ollama connection failures (Lesson 06)

Format: A markdown cell with a bold "Pitfall" header and 2-3 sentences.

## Lesson Outlines

### Lesson 01: Python & FastAPI Basics

**What you build:** A FastAPI app with `/health` endpoint, Pydantic request/response models, async route handlers, and config via pydantic-settings.

**Key concepts:**
- Decorator-based routing (`@app.get`) vs explicit router registration in Go/chi or Express
- Pydantic models as Go structs with built-in validation
- `async def` vs `def` — when and why (compared to goroutines)
- Type hints — Python's approach vs Go's static typing
- `pydantic-settings` for environment-driven config (like `envconfig` in Go)

**Dependencies:** fastapi, uvicorn, pydantic-settings

**Requires running:** Nothing external

### Lesson 02: PDF Parsing & Chunking

**What you build:** `extract_pages()` function using PyPDF2, `chunk_pages()` function using LangChain's RecursiveCharacterTextSplitter.

**Key concepts:**
- BytesIO as Python's in-memory file buffer (like Go's `bytes.Buffer` / `io.Reader`)
- Why you chunk text (LLM context windows, embedding quality)
- Why overlap matters (context continuity across chunk boundaries)
- RecursiveCharacterTextSplitter — what "recursive" means (tries multiple separators)

**Experiment cells:**
- Change chunk_size from 100 to 2000, observe chunk count and content
- Change chunk_overlap from 0 to 500, see how chunks share content
- Try a PDF with tables or complex formatting, see what PyPDF2 misses

**Dependencies:** PyPDF2, langchain-text-splitters, fpdf2 (for creating test PDFs)

**Requires running:** Nothing external

### Lesson 03: Embeddings & Vector Spaces

**What you build:** `embed_texts()` async function using httpx to call Ollama's `/api/embed` endpoint.

**Key concepts:**
- What embeddings are — text → dense vectors that capture semantic meaning
- Why 768 dimensions (nomic-embed-text's output size)
- Cosine similarity — why it works for comparing meaning
- httpx as Python's async HTTP client (like Go's `http.Client` but async)
- async/await for I/O-bound operations

**Experiment cells:**
- Embed similar sentences, compute cosine similarity with numpy
- Embed dissimilar sentences, compare scores
- Embed a question and a passage that answers it — see high similarity

**Dependencies:** httpx, numpy

**Requires running:** Ollama with nomic-embed-text

### Lesson 04: Qdrant Vector Storage

**What you build:** `QdrantStore` class (upsert, list_documents) and `QdrantRetriever` class (search).

**Key concepts:**
- Vector databases — what they are, why you can't just use Postgres for this
- Collections, points, payloads — Qdrant's data model
- Cosine distance for similarity search
- Python classes vs Go structs+methods — `__init__` is the constructor, `self` is the receiver
- Why you store metadata (filename, page) alongside vectors — for citations

**Experiment cells:**
- Upsert vectors, then search with a query vector
- Change top_k and observe how many results come back
- Search with a vector that has no good matches — see low scores

**Dependencies:** qdrant-client

**Requires running:** Qdrant (Docker container)

### Lesson 05: RAG Chain & Prompt Engineering

**What you build:** `build_rag_prompt()` function, the full retrieve→prompt→generate pipeline.

**Key concepts:**
- RAG pattern — retrieve context, build prompt, generate response
- Grounding — telling the LLM to only use provided context (prevents hallucination)
- Prompt templates — system prompt vs user prompt, few-shot patterns
- What happens without grounding (hallucination demo — ask about something not in context)
- Source attribution in the prompt ("cite filename and page")

**Experiment cells:**
- Build a prompt with context, send to Ollama, see grounded response
- Build a prompt WITHOUT context instructions, see hallucination
- Modify the system prompt to change the assistant's tone/behavior

**Dependencies:** httpx

**Requires running:** Ollama with mistral, Qdrant with data from Lesson 04

### Lesson 06: Streaming & SSE

**What you build:** `stream_ollama_response()` async generator function.

**Key concepts:**
- Async generators (`async def` + `yield`) — Python's equivalent of Go channels for streaming
- Why stream (user experience — tokens appear immediately vs waiting for full response)
- SSE format — `data: {...}\n\n` lines, parsed client-side
- httpx streaming with `client.stream()` and `response.aiter_lines()`
- The `done` event pattern — final event carries metadata (sources)

**Experiment cells:**
- Stream a response token by token, printing each as it arrives
- Compare streamed vs non-streamed response times (perceived latency)

**Dependencies:** httpx

**Requires running:** Ollama with mistral

### Lesson 07: Wiring the Endpoints

**What you build:** Complete `/ingest` and `/chat` FastAPI endpoints, assembling all previous pieces.

**Key concepts:**
- File upload handling with `UploadFile` (like multipart parsing in Go)
- CORS middleware — what it is, why the frontend needs it
- Lazy singleton pattern for database connections (`get_store()`)
- SSE response via `sse-starlette` (wrapping the async generator in an EventSourceResponse)
- Error handling patterns — HTTPException for client errors, validation flow
- The full data flow: PDF → pages → chunks → embeddings → Qdrant → search → prompt → stream

**This lesson ties everything together.** By the end, you have both services rebuilt and can compare your code to `services/`.

**Dependencies:** fastapi, uvicorn, python-multipart, PyPDF2, langchain-text-splitters, httpx, qdrant-client, pydantic-settings, sse-starlette

**Requires running:** Ollama (mistral + nomic-embed-text), Qdrant

## File Structure

```
lessons/
├── 01_python_fastapi_basics.ipynb
├── 02_pdf_parsing_and_chunking.ipynb
├── 03_embeddings_and_vectors.ipynb
├── 04_qdrant_vector_storage.ipynb
├── 05_rag_chain_and_prompts.ipynb
├── 06_streaming_and_sse.ipynb
├── 07_wiring_the_endpoints.ipynb
└── requirements.txt
```

A single `requirements.txt` covers all lessons:
```
fastapi==0.115.0
uvicorn[standard]==0.30.0
python-multipart==0.0.9
pypdf2==3.0.1
langchain-text-splitters==0.2.0
qdrant-client==1.9.0
httpx==0.27.0
pydantic-settings==2.3.0
sse-starlette==2.1.0
numpy
fpdf2
jupyter
```

## Prerequisites Per Lesson

| Lesson | Ollama | Qdrant | Notes |
|--------|--------|--------|-------|
| 01 | No | No | Pure Python, no external services |
| 02 | No | No | Creates test PDFs in-notebook |
| 03 | Yes (nomic-embed-text) | No | Calls Ollama embedding API |
| 04 | No | Yes | Connects to Qdrant, creates collections |
| 05 | Yes (mistral + nomic-embed-text) | Yes | Full RAG pipeline |
| 06 | Yes (mistral) | No | Streams from Ollama |
| 07 | Yes (mistral + nomic-embed-text) | Yes | Full stack |

Each notebook starts with a prerequisites cell that checks connectivity and prints a clear error if a service is unavailable.

## Out of Scope

- Docker/deployment lessons (Kyle already knows Docker well)
- Integration testing lesson (already done live)
- Frontend lessons (not Python, Kyle knows TypeScript)
- Unit testing walkthrough (tests exist in services/ as reference)
