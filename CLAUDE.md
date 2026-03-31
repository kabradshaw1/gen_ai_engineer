# CLAUDE.md

## Project Intent

Portfolio project for a Gen AI Engineer job application — a Document Q&A Assistant demonstrating RAG architecture, prompt engineering, and Python API development.

## Tech Stack

- FastAPI (Python backend microservices)
- Qdrant (vector database, Docker container)
- Ollama (mistral 7B for chat, nomic-embed-text for embeddings)
- LangChain text splitters (chunking only — not using the full LangChain framework)
- Next.js + TypeScript + shadcn/ui (frontend)
- Docker Compose (backend orchestration)

## Infrastructure

- **Mac (dev machine):** Code editing, frontend dev server, no GPU
- **Windows (PC@100.79.113.84 via Tailscale):** Ollama (RTX 3090), Docker Compose (Qdrant + backend services)
- **SSH:** `ssh PC@100.79.113.84` — key-based auth configured
- **Local dev:** SSH tunnels forward `localhost:8001` and `localhost:8002` to Windows backend
  ```bash
  ssh -f -N -L 8001:localhost:8001 -L 8002:localhost:8002 PC@100.79.113.84
  ```
- **Frontend:** `npm run dev` in `frontend/`, points to `localhost:8001`/`8002` via tunnels
- **Production:** Frontend on Vercel (`https://kylebradshaw.dev`), backend via Cloudflare Tunnel:
  - `https://api-chat.kylebradshaw.dev` → Windows PC :8002
  - `https://api-ingestion.kylebradshaw.dev` → Windows PC :8001
  - Cloudflared installed as Windows service (auto-starts on boot)

## Project Structure

```
services/
├── ingestion/          # FastAPI — PDF upload, parse, chunk, embed, store
│   ├── app/            # main.py, pdf_parser.py, chunker.py, embedder.py, store.py, config.py
│   └── tests/          # 20 unit tests
├── chat/               # FastAPI — question embed, search, RAG prompt, stream
│   ├── app/            # main.py, retriever.py, prompt.py, chain.py, config.py
│   └── tests/          # 12 unit tests
frontend/               # Next.js + shadcn/ui — chat UI, PDF upload, SSE streaming
lessons/                # 7 Jupyter notebooks rebuilding the services from scratch
docker-compose.yml      # Qdrant + ingestion + chat services
.env.example            # Config template
```

## Kyle's Background

- Strong in Go and TypeScript (full-stack web apps)
- Experienced with Docker, Kubernetes, GitHub Actions, SQL/NoSQL
- Has used Ollama and built web services to interact with it
- Limited hands-on experience with Python data processing, LLM workflows, RAG, prompt engineering
- Has written Python for Django, taken tutorials, but limited production Python experience

## Lesson Notebooks

7 notebooks in `lessons/` that rebuild both Python services from scratch. Each notebook follows this format:

1. **Intro** — what you're building and why
2. **Prerequisites** — pip installs, connectivity checks
3. **Package Introductions** — what each library does, why chosen, key APIs
4. **Go/TS Comparison** — map concepts to what Kyle already knows
5. **Build It** — code cells with explanatory markdown between
6. **Experiment** — tweak parameters, observe effects
7. **Check Your Understanding** — reflection prompts

## Design Specs

- `docs/superpowers/specs/2026-03-31-document-qa-assistant-design.md` — full system architecture
- `docs/superpowers/specs/2026-03-31-frontend-design.md` — frontend design
- `docs/superpowers/specs/2026-03-31-lesson-notebooks-design.md` — lesson notebook design

## Current State

- **Branch:** `feat/backend-services` — all work lives here, not yet merged to main
- **Backend:** Complete — 32 tests passing, Docker Compose tested on Windows
- **Frontend:** Complete — builds clean, tested against live backend via SSH tunnels
- **Lessons:** Complete — 7 notebooks created
- **Deployed:** Frontend on Vercel, backend via Cloudflare Tunnel (cloudflared Windows service)
- **Remaining:** None — all features complete
