# CLAUDE.md

## Project Intent

Portfolio project for a Gen AI Engineer job application — a Document Q&A Assistant demonstrating RAG architecture, prompt engineering, and Python API development.

## Environment

- **Python 3.11** via **Miniconda** (conda env: `py-refresher`)
- **Ollama** served locally on RTX 3090 (24GB VRAM)
- **Docker Compose** for service orchestration
- **Next.js** frontend (local dev + Vercel deployment)

## Tech Stack

- FastAPI (Python backend microservices)
- Qdrant (vector database, Docker container)
- Ollama (mistral 7B for chat, nomic-embed-text for embeddings)
- LangChain (RAG orchestration)
- Next.js + TypeScript (frontend)
- Docker Compose (deployment)

## Kyle's Background

- Strong in Go and TypeScript (full-stack web apps)
- Experienced with Docker, Kubernetes, GitHub Actions, SQL/NoSQL
- Has used Ollama and built web services to interact with it
- Limited hands-on experience with Python data processing, LLM workflows, LangChain, RAG, prompt engineering
- Has written Python for Django, taken tutorials, but limited production Python experience

## Lesson Notebooks

The `lessons/` directory contains build-along Jupyter notebooks. Each lesson teaches a concept by building a piece of the portfolio app. Format per concept:

1. **Go/TS comparison** — frame the concept using what Kyle already knows
2. **Code cell — example** — complete, runnable code
3. **Code cell — experiment** — variation or "try this" prompt
4. **Markdown cell — "In your own words"** — reflection prompt

## Design Spec

See `docs/superpowers/specs/2026-03-31-document-qa-assistant-design.md` for the full architecture and design.
