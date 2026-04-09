"""Mock Ollama server for CI.

Stubs the endpoints the Python AI services actually call:
- POST /api/embed       — batch API: {model, input: [list]} → {embeddings: [[...]]}.
                          Used by ingestion, chat, and debug embed calls.
- POST /api/embeddings  — legacy singular API: {model, prompt} → {embedding: [...]}.
                          Kept for completeness; not currently used by the services.
- POST /api/generate    — NDJSON stream with {response, done}. Used by
                          services/chat/app/chain.py:stream_ollama_response.
- POST /api/chat        — NDJSON stream with {message:{role,content}, done}. Used
                          by services/debug/app/agent.py (non-streaming via the
                          debug agent loop, but exposed here for completeness).
- GET  /api/tags        — returns an empty model list (used by health check).

The real Ollama response schemas are pinned here verbatim so a future
version bump in Ollama will break these tests loudly instead of drifting
silently. If Ollama's API changes, update both the real client code and
this stub together.
"""

import json

from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from pydantic import BaseModel

app = FastAPI(title="mock-ollama")

# 768 matches the nomic-embed-text embedding size. Changing this will
# cause Qdrant to reject inserts from ingestion because the collection
# schema pins a fixed vector size.
EMBEDDING_DIM = 768


def _deterministic_vector(text: str) -> list[float]:
    """Return a reproducible EMBEDDING_DIM-length vector derived from text."""
    seed = sum(ord(c) for c in text) or 1
    return [((seed * (i + 1)) % 2000 - 1000) / 1000.0 for i in range(EMBEDDING_DIM)]


class EmbedRequest(BaseModel):
    model: str
    # Ollama's /api/embed accepts either a string or list of strings.
    input: str | list[str]


@app.post("/api/embed")
def embed(req: EmbedRequest) -> dict:
    """Batch embed endpoint used by ingestion/chat/debug."""
    texts = [req.input] if isinstance(req.input, str) else req.input
    return {"embeddings": [_deterministic_vector(t) for t in texts]}


class EmbeddingsRequest(BaseModel):
    model: str
    prompt: str


@app.post("/api/embeddings")
def embeddings(req: EmbeddingsRequest) -> dict:
    """Legacy singular embed endpoint. Not currently used by services."""
    return {"embedding": _deterministic_vector(req.prompt)}


class GenerateRequest(BaseModel):
    model: str
    prompt: str
    system: str | None = None
    stream: bool = True


def _generate_stream() -> bytes:
    # NDJSON stream used by services/chat/app/chain.py — reads
    # data["response"] per line and breaks when data["done"] is true.
    chunks = [
        {"response": "This is ", "done": False},
        {"response": "a mock response.", "done": False},
        {"response": "", "done": True},
    ]
    return ("\n".join(json.dumps(c) for c in chunks) + "\n").encode()


@app.post("/api/generate")
def generate(_req: GenerateRequest) -> StreamingResponse:
    return StreamingResponse(
        iter([_generate_stream()]),
        media_type="application/x-ndjson",
    )


class ChatMessage(BaseModel):
    role: str
    content: str


class ChatRequest(BaseModel):
    model: str
    messages: list[ChatMessage]
    stream: bool = True


def _chat_stream() -> bytes:
    chunks = [
        {"message": {"role": "assistant", "content": "This is "}, "done": False},
        {
            "message": {"role": "assistant", "content": "a mock response."},
            "done": False,
        },
        {"message": {"role": "assistant", "content": ""}, "done": True},
    ]
    return ("\n".join(json.dumps(c) for c in chunks) + "\n").encode()


@app.post("/api/chat")
def chat(req: ChatRequest):
    # Real Ollama returns a single JSON object when stream=false and an
    # NDJSON stream when stream=true. Debug agent uses stream=false; chat
    # service uses /api/generate instead. Handle both for parity.
    if not req.stream:
        return {
            "message": {
                "role": "assistant",
                "content": "mock response",
            },
            "done": True,
        }
    return StreamingResponse(
        iter([_chat_stream()]),
        media_type="application/x-ndjson",
    )


@app.get("/api/tags")
def tags() -> dict:
    return {"models": []}


@app.get("/health")
def health() -> dict:
    return {"status": "healthy"}
