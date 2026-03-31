# Document Q&A Assistant — Design Spec

## Overview

A portfolio project demonstrating RAG architecture, prompt engineering, agentic AI workflows, and Python API development for a Gen AI Engineer role. Users upload PDFs, ask questions, and receive AI-generated answers grounded in document content with source citations.

## Target Job Requirements Covered

- Python API development (FastAPI)
- Data processing and AI-enabled workflows
- LLM, prompt engineering, RAG architecture
- Hybrid prompting technique
- LangChain, Hugging Face (tokenizers)
- Docker containerization
- Git version control
- NLP fundamentals

## Architecture

### System Overview

```
Vercel (Cloud)                        Docker Compose (Local)
┌─────────────────┐    /ingest     ┌──────────────────┐
│  Next.js        │───────────────▶│  Ingestion API   │
│  Frontend       │                │  (FastAPI)        │
│                 │    /chat       ├──────────────────┤
│  - Chat UI      │───────────────▶│  Chat API        │
│  - PDF upload   │◀── SSE ───────│  (FastAPI)        │
│  - Source cites  │               └────┬────────┬────┘
└─────────────────┘                     │        │
                                        ▼        ▼
                                  ┌──────────┐ ┌──────────┐
                                  │  Qdrant  │ │  Ollama  │
                                  │(vectors) │ │  (LLM)   │
                                  └──────────┘ └──────────┘
```

### Services

**Ingestion API (FastAPI)**
- Receives PDF uploads via multipart form
- Extracts text using PyPDF2
- Chunks text with LangChain RecursiveCharacterTextSplitter (1000 chars, 200 overlap)
- Embeds chunks via Ollama nomic-embed-text (768 dimensions)
- Stores vectors + metadata (filename, page number) in Qdrant

**Chat API (FastAPI)**
- Receives user question
- Embeds question via Ollama nomic-embed-text
- Searches Qdrant for top-5 most similar chunks
- Builds RAG prompt with retrieved context
- Streams response from Ollama mistral via SSE
- Returns source citations (filename + page) with final token

**Qdrant (Vector DB)**
- Runs as Docker container
- Stores document embeddings in collections
- Cosine similarity search
- REST API for both Python services

**Ollama (Local LLM)**
- mistral 7B for chat/completion (~4GB VRAM, RTX 3090)
- nomic-embed-text for embeddings (768 dimensions)
- Runs on host with GPU passthrough or as container

**Next.js Frontend (Vercel)**
- Minimal, functional UI
- Single-page chat interface
- PDF upload button in header
- Streaming token-by-token message display
- Source citations below each answer
- Document list sidebar

## API Design

### Ingestion API

**POST /ingest**
```
Request: multipart/form-data
  file: PDF binary
  collection: string (optional, default: "default")

Response: 200
{
  "status": "success",
  "document_id": "uuid",
  "chunks_created": 42,
  "filename": "report.pdf"
}
```

**GET /documents**
```
Response: 200
{
  "documents": [
    {
      "id": "uuid",
      "filename": "report.pdf",
      "chunks": 42,
      "ingested_at": "2026-03-31T..."
    }
  ]
}
```

### Chat API

**POST /chat**
```
Request: application/json
{
  "question": "What are the key findings?",
  "collection": "default"
}

Response: text/event-stream
data: {"token": "The"}
data: {"token": " report"}
...
data: {"done": true, "sources": [{"file": "report.pdf", "page": 3}]}
```

**GET /health**
```
Response: 200
{
  "status": "healthy",
  "qdrant": "connected",
  "ollama": "connected"
}
```

## Project Structure

```
gen_ai_engineer/
├── docker-compose.yml
├── .env.example
├── services/
│   ├── ingestion/
│   │   ├── Dockerfile
│   │   ├── requirements.txt
│   │   ├── app/
│   │   │   ├── main.py          # FastAPI app + /ingest endpoint
│   │   │   ├── pdf_parser.py    # PDF text extraction
│   │   │   ├── chunker.py       # Text splitting logic
│   │   │   ├── embedder.py      # Ollama embedding calls
│   │   │   └── store.py         # Qdrant storage operations
│   │   └── tests/
│   └── chat/
│       ├── Dockerfile
│       ├── requirements.txt
│       ├── app/
│       │   ├── main.py          # FastAPI app + /chat endpoint
│       │   ├── retriever.py     # Qdrant search logic
│       │   ├── prompt.py        # Prompt templates
│       │   └── chain.py         # LangChain RAG chain
│       └── tests/
├── frontend/
│   ├── package.json
│   ├── src/
│   │   ├── app/
│   │   │   ├── page.tsx         # Main chat page
│   │   │   └── layout.tsx
│   │   └── components/
│   │       ├── ChatWindow.tsx   # Message display
│   │       ├── MessageInput.tsx # Text input + send
│   │       └── FileUpload.tsx   # PDF upload widget
│   └── next.config.js
└── lessons/
    ├── 01_python_fastapi_basics.ipynb
    ├── 02_pdf_parsing_and_chunking.ipynb
    ├── 03_embeddings_and_vectors.ipynb
    ├── 04_qdrant_vector_storage.ipynb
    ├── 05_langchain_rag_chain.ipynb
    ├── 06_prompt_engineering.ipynb
    ├── 07_streaming_and_sse.ipynb
    ├── 08_docker_compose_deployment.ipynb
    └── 09_putting_it_all_together.ipynb
```

## Data Flow

### Ingestion Flow
1. User uploads PDF via Next.js UI
2. Frontend sends multipart POST to Ingestion API
3. PyPDF2 extracts text, preserving page boundaries
4. RecursiveCharacterTextSplitter chunks text (1000 chars, 200 overlap)
5. Each chunk embedded via Ollama nomic-embed-text → 768d vector
6. Vectors + metadata (filename, page number, chunk index) upserted to Qdrant

### Chat Flow
1. User types question in Next.js UI
2. Frontend sends POST to Chat API
3. Question embedded via Ollama nomic-embed-text → 768d vector
4. Qdrant cosine similarity search returns top-5 chunks
5. RAG prompt assembled: system prompt + context chunks + user question
6. Prompt sent to Ollama mistral, response streamed via SSE
7. Final SSE event includes source citations (filename + page)

## Models

| Model | Purpose | Size | VRAM |
|-------|---------|------|------|
| mistral 7B | Chat/completion | ~4GB | Comfortable on RTX 3090 |
| nomic-embed-text | Embeddings (768d) | ~275MB | Minimal |

## Lesson Series

9 build-along Jupyter notebooks. Each lesson builds a piece of the actual portfolio app. Format follows CLAUDE.md conventions:

1. **Go/TS comparison** — frame each concept in terms Kyle already knows
2. **Runnable code example** — complete, executable demonstration
3. **Experiment prompt** — variation to try
4. **"In your own words"** — reflection prompt

### Lesson Outline

| # | Topic | Builds |
|---|-------|--------|
| 01 | Python & FastAPI Basics | Service scaffolding, routes, Pydantic models |
| 02 | PDF Parsing & Chunking | pdf_parser.py, chunker.py |
| 03 | Embeddings & Vector Spaces | embedder.py, understanding similarity |
| 04 | Qdrant Vector Storage | store.py, retriever.py |
| 05 | LangChain RAG Chain | chain.py, document loaders, retrievers |
| 06 | Prompt Engineering | prompt.py, hybrid prompting, hallucination handling |
| 07 | Streaming & SSE | FastAPI StreamingResponse, Next.js consumption |
| 08 | Docker Compose Deployment | Dockerfiles, compose config, GPU passthrough |
| 09 | Putting It All Together | Integration testing, edge cases, Claude Code pitfalls |

### Claude Code Pitfall Callouts

Each lesson includes warnings about where Claude Code may make incorrect assumptions when given vague instructions. Examples:
- Generating sync code where async is required
- Using deprecated LangChain APIs
- Incorrect Qdrant collection schemas
- Missing error handling on Ollama connection failures
- Hardcoded model names instead of config-driven

## Testing Strategy

**Unit tests (pytest)**
- Each module tested independently: chunker, embedder, retriever, prompt builder
- Mock external services (Ollama, Qdrant) for fast unit tests

**Integration tests**
- Docker Compose test profile spins up real Qdrant + Ollama
- Test full ingestion flow: PDF → vectors in Qdrant
- Test full chat flow: question → streamed answer with sources

**Health checks**
- Docker Compose health checks ensure startup order
- Chat API /health verifies Qdrant + Ollama connectivity

**Manual smoke test**
- Upload a known PDF, ask specific questions, verify correct page citations

## Deployment Model

**Local development (primary):** All services run locally via Docker Compose. The Next.js frontend runs via `npm run dev` and calls `localhost` APIs. This is the main demo mode.

**Vercel deployment:** The Next.js frontend can be deployed to Vercel as a static portfolio piece. When deployed, it needs a publicly accessible backend — this is out of scope for the initial build. The Vercel deployment demonstrates Next.js/Vercel skills; the full RAG demo runs locally.

## Error Handling

- **Ollama unreachable:** Both services return 503 with clear error message. Health check endpoint reports status.
- **Corrupt/unreadable PDF:** Ingestion API returns 422 with details about which pages failed. Partial ingestion of readable pages is supported.
- **Qdrant down:** Chat API returns 503. Health check reports Qdrant status.
- **Empty search results:** Chat API responds honestly ("I don't have enough context to answer that") rather than hallucinating.
- **Oversized PDF:** Ingestion API enforces a configurable max file size (default 50MB).

## Job Requirements Out of Scope

The job posting mentions several technologies that are not included in this initial build:
- **Hugging Face** — could be added later as an alternative embedding provider
- **LlamaIndex** — LangChain was chosen instead; both serve similar purposes
- **Gemini LLM** — using local Ollama models instead; Gemini integration could be a future enhancement
- **PyTorch/TensorFlow/Keras** — ML framework knowledge is demonstrated conceptually in the lessons but not used directly in the app

These are noted as potential future enhancements, not gaps.

## Environment

- Python 3.11 via Miniconda (conda env: `py-refresher`)
- Ollama served locally on RTX 3090 (24GB VRAM)
- Docker Compose for service orchestration
- Next.js frontend (local dev + Vercel deployment)
- Qdrant as Docker container

