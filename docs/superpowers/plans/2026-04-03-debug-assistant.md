# Debug Assistant Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an agentic debug assistant that indexes Python codebases and diagnoses bugs through a multi-step tool-calling loop, with a real-time agent timeline UI.

**Architecture:** New `services/debug/` FastAPI service with custom agent loop calling Ollama's tool-use API. Four tools (vector search, file read, grep, test runner) execute in a loop until diagnosis. SSE streams each step to a Next.js frontend at `/ai/debug`.

**Tech Stack:** FastAPI, Ollama (Qwen 2.5 14B + nomic-embed-text), Qdrant, LangChain text splitters, httpx, sse-starlette, Next.js, shadcn/ui, TypeScript

---

## File Structure

### New Files: `services/debug/`

```
services/debug/
├── Dockerfile
├── requirements.txt
├── app/
│   ├── __init__.py
│   ├── main.py          # FastAPI app: /health, /index, /debug endpoints
│   ├── config.py        # Settings from env vars (pydantic-settings)
│   ├── indexer.py        # Walk Python project, chunk with Language.PYTHON splitter, embed, store
│   ├── agent.py          # Agent loop: build prompt → call Ollama → parse tool call → execute → repeat
│   ├── tools.py          # Tool definitions (JSON schemas) + execute functions
│   └── prompts.py        # System prompt and tool-use templates
└── tests/
    ├── __init__.py
    ├── conftest.py
    ├── test_config.py
    ├── test_indexer.py
    ├── test_tools.py
    ├── test_agent.py
    ├── test_prompts.py
    └── test_main.py
```

### New Files: Frontend

```
frontend/src/
├── app/ai/debug/
│   └── page.tsx          # Debug assistant page
└── components/
    ├── DebugForm.tsx      # Bug description + error output input form
    ├── AgentTimeline.tsx   # Real-time step-by-step execution display
    ├── ToolCallCard.tsx    # Collapsible card for tool calls and results
    └── DiagnosisBanner.tsx # Final diagnosis display
```

### Modified Files

```
frontend/src/app/ai/page.tsx        # Add Debug Assistant section + Mermaid diagram
docker-compose.yml                   # Add debug service
.env.example                         # Add DEBUG_MODEL setting
.github/workflows/ci.yml             # Add debug service to matrices
```

---

### Task 1: Service Scaffold — Config and Health Check

**Files:**
- Create: `services/debug/app/__init__.py`
- Create: `services/debug/app/config.py`
- Create: `services/debug/app/main.py`
- Create: `services/debug/requirements.txt`
- Create: `services/debug/Dockerfile`
- Create: `services/debug/tests/__init__.py`
- Create: `services/debug/tests/conftest.py`
- Test: `services/debug/tests/test_config.py`
- Test: `services/debug/tests/test_main.py`

- [ ] **Step 1: Create requirements.txt**

```
# services/debug/requirements.txt
fastapi==0.135.3
uvicorn[standard]==0.30.0
langchain-text-splitters==0.2.4
langchain-community==0.2.19
qdrant-client==1.9.0
httpx==0.28.1
pydantic-settings==2.3.0
sse-starlette==2.1.0
pytest==8.2.0
pytest-asyncio==0.25.3
pytest-cov==7.1.0
```

- [ ] **Step 2: Create config.py**

```python
# services/debug/app/config.py
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ollama_base_url: str = "http://host.docker.internal:11434"
    chat_model: str = "qwen2.5:14b"
    embedding_model: str = "nomic-embed-text"
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    max_agent_steps: int = 10
    max_file_lines: int = 100
    max_grep_matches: int = 20
    test_timeout_seconds: int = 30
    allowed_origins: str = "https://kylebradshaw.dev"


settings = Settings()
```

- [ ] **Step 3: Create empty __init__.py files**

Create empty files at:
- `services/debug/app/__init__.py`
- `services/debug/tests/__init__.py`

- [ ] **Step 4: Create conftest.py**

```python
# services/debug/tests/conftest.py
```

Empty for now — fixtures will be added in later tasks.

- [ ] **Step 5: Write failing test for config**

```python
# services/debug/tests/test_config.py
from app.config import Settings


def test_default_settings():
    s = Settings()
    assert s.ollama_base_url == "http://host.docker.internal:11434"
    assert s.chat_model == "qwen2.5:14b"
    assert s.embedding_model == "nomic-embed-text"
    assert s.qdrant_host == "qdrant"
    assert s.qdrant_port == 6333
    assert s.max_agent_steps == 10
    assert s.max_file_lines == 100
    assert s.max_grep_matches == 20
    assert s.test_timeout_seconds == 30


def test_settings_from_env(monkeypatch):
    monkeypatch.setenv("CHAT_MODEL", "mistral")
    monkeypatch.setenv("MAX_AGENT_STEPS", "5")
    s = Settings()
    assert s.chat_model == "mistral"
    assert s.max_agent_steps == 5
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd services/debug && python -m pytest tests/test_config.py -v`
Expected: 2 passed

- [ ] **Step 7: Write failing test for health endpoint**

```python
# services/debug/tests/test_main.py
import json
from unittest.mock import AsyncMock, MagicMock, patch

from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


@patch("app.main.httpx.AsyncClient")
@patch("app.main.QdrantClient")
def test_health_ok(mock_qdrant_cls, mock_httpx_cls):
    mock_qdrant = MagicMock()
    mock_qdrant.get_collections.return_value = True
    mock_qdrant_cls.return_value = mock_qdrant

    mock_client = AsyncMock()
    mock_response = AsyncMock(status_code=200)
    mock_client.get.return_value = mock_response
    mock_httpx_cls.return_value.__aenter__ = AsyncMock(return_value=mock_client)
    mock_httpx_cls.return_value.__aexit__ = AsyncMock(return_value=False)

    response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


@patch("app.main.httpx.AsyncClient")
@patch("app.main.QdrantClient")
def test_health_qdrant_down(mock_qdrant_cls, mock_httpx_cls):
    mock_qdrant = MagicMock()
    mock_qdrant.get_collections.side_effect = Exception("connection refused")
    mock_qdrant_cls.return_value = mock_qdrant

    response = client.get("/health")
    assert response.status_code == 503
```

- [ ] **Step 8: Run test to verify it fails**

Run: `cd services/debug && python -m pytest tests/test_main.py -v`
Expected: FAIL — `app.main` does not exist yet

- [ ] **Step 9: Implement main.py with health endpoint**

```python
# services/debug/app/main.py
import logging

import httpx
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from qdrant_client import QdrantClient

from app.config import settings

logger = logging.getLogger(__name__)

app = FastAPI(title="Debug Assistant API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins.split(","),
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)


@app.get("/health")
async def health():
    try:
        qd = QdrantClient(host=settings.qdrant_host, port=settings.qdrant_port)
        qd.get_collections()
    except Exception:
        logger.error("Qdrant health check failed", exc_info=True)
        return JSONResponse(
            status_code=503,
            content={"status": "unhealthy", "detail": "Qdrant unavailable"},
        )

    try:
        async with httpx.AsyncClient() as client:
            resp = await client.get(f"{settings.ollama_base_url}/api/tags", timeout=5.0)
            resp.raise_for_status()
    except Exception:
        logger.error("Ollama health check failed", exc_info=True)
        return JSONResponse(
            status_code=503,
            content={"status": "unhealthy", "detail": "Ollama unavailable"},
        )

    return {"status": "ok"}
```

- [ ] **Step 10: Run tests to verify they pass**

Run: `cd services/debug && python -m pytest tests/test_main.py -v`
Expected: 2 passed

- [ ] **Step 11: Create Dockerfile**

```dockerfile
# services/debug/Dockerfile
FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

RUN useradd --create-home appuser
USER appuser

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

- [ ] **Step 12: Run ruff checks**

Run: `cd services/debug && ruff check app/ tests/ && ruff format --check app/ tests/`
Expected: No errors

- [ ] **Step 13: Commit**

```bash
git add services/debug/
git commit -m "feat(debug): scaffold service with config and health endpoint"
```

---

### Task 2: Code Indexer

**Files:**
- Create: `services/debug/app/indexer.py`
- Test: `services/debug/tests/test_indexer.py`

- [ ] **Step 1: Write failing tests for indexer**

```python
# services/debug/tests/test_indexer.py
import os
import tempfile
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from app.indexer import collect_python_files, chunk_code_files, index_project


def test_collect_python_files_finds_py_files():
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create some Python files
        open(os.path.join(tmpdir, "main.py"), "w").write("def hello(): pass")
        open(os.path.join(tmpdir, "util.py"), "w").write("x = 1")
        # Create non-Python files (should be ignored)
        open(os.path.join(tmpdir, "readme.md"), "w").write("# readme")
        open(os.path.join(tmpdir, "data.json"), "w").write("{}")

        files = collect_python_files(tmpdir)
        assert len(files) == 2
        assert all(f.endswith(".py") for f in files)


def test_collect_python_files_skips_hidden_and_venv():
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create files in directories that should be skipped
        os.makedirs(os.path.join(tmpdir, ".git"))
        open(os.path.join(tmpdir, ".git", "config.py"), "w").write("")
        os.makedirs(os.path.join(tmpdir, "__pycache__"))
        open(os.path.join(tmpdir, "__pycache__", "mod.py"), "w").write("")
        os.makedirs(os.path.join(tmpdir, "venv", "lib"), exist_ok=True)
        open(os.path.join(tmpdir, "venv", "lib", "site.py"), "w").write("")
        # Create a valid file
        open(os.path.join(tmpdir, "app.py"), "w").write("x = 1")

        files = collect_python_files(tmpdir)
        assert len(files) == 1
        assert files[0].endswith("app.py")


def test_chunk_code_files_splits_on_functions():
    with tempfile.TemporaryDirectory() as tmpdir:
        code = '''def foo():
    """Do foo."""
    return 1


def bar():
    """Do bar."""
    return 2


class Baz:
    """A class."""

    def method(self):
        return 3
'''
        filepath = os.path.join(tmpdir, "module.py")
        open(filepath, "w").write(code)

        chunks = chunk_code_files([filepath], project_root=tmpdir)
        assert len(chunks) >= 1
        for chunk in chunks:
            assert "text" in chunk
            assert "file_path" in chunk
            assert "start_line" in chunk
            assert "end_line" in chunk
            assert chunk["file_path"] == "module.py"  # Relative path


def test_chunk_code_files_skips_empty_files():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "empty.py")
        open(filepath, "w").write("")

        chunks = chunk_code_files([filepath], project_root=tmpdir)
        assert len(chunks) == 0


@pytest.mark.asyncio
@patch("app.indexer.QdrantClient")
@patch("app.indexer.embed_texts", new_callable=AsyncMock)
async def test_index_project_returns_stats(mock_embed, mock_qdrant_cls):
    mock_embed.return_value = [[0.1] * 768]
    mock_qdrant = MagicMock()
    mock_qdrant_cls.return_value = mock_qdrant

    with tempfile.TemporaryDirectory() as tmpdir:
        open(os.path.join(tmpdir, "app.py"), "w").write("def hello(): pass")

        result = await index_project(
            project_path=tmpdir,
            ollama_base_url="http://localhost:11434",
            embedding_model="nomic-embed-text",
            qdrant_host="localhost",
            qdrant_port=6333,
        )

    assert result["files_indexed"] == 1
    assert result["chunks"] >= 1
    assert "collection" in result
    mock_qdrant.delete_collection.assert_called_once()
    mock_qdrant.create_collection.assert_called_once()
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/debug && python -m pytest tests/test_indexer.py -v`
Expected: FAIL — `app.indexer` does not exist

- [ ] **Step 3: Implement indexer.py**

```python
# services/debug/app/indexer.py
import os
import uuid

import httpx
from langchain_text_splitters import Language, RecursiveCharacterTextSplitter
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, PointStruct, VectorParams

SKIP_DIRS = {
    ".git",
    "__pycache__",
    ".mypy_cache",
    ".pytest_cache",
    ".ruff_cache",
    "venv",
    ".venv",
    "env",
    ".env",
    "node_modules",
    ".tox",
    "dist",
    "build",
    "egg-info",
}


def collect_python_files(project_path: str) -> list[str]:
    """Walk project directory and return absolute paths to .py files."""
    py_files = []
    for root, dirs, files in os.walk(project_path):
        dirs[:] = [d for d in dirs if d not in SKIP_DIRS and not d.startswith(".")]
        for f in files:
            if f.endswith(".py"):
                py_files.append(os.path.join(root, f))
    return sorted(py_files)


def chunk_code_files(
    file_paths: list[str],
    project_root: str,
    chunk_size: int = 1500,
    chunk_overlap: int = 200,
) -> list[dict]:
    """Chunk Python files using language-aware splitting."""
    splitter = RecursiveCharacterTextSplitter.from_language(
        language=Language.PYTHON,
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap,
    )

    chunks = []
    for filepath in file_paths:
        content = open(filepath).read()
        if not content.strip():
            continue

        rel_path = os.path.relpath(filepath, project_root)
        splits = splitter.split_text(content)

        # Track line numbers for each split
        lines = content.split("\n")
        search_from = 0
        for split_text in splits:
            first_line = split_text.split("\n")[0]
            start_line = 1
            for i in range(search_from, len(lines)):
                if lines[i] == first_line:
                    start_line = i + 1  # 1-indexed
                    search_from = i + 1
                    break

            end_line = start_line + split_text.count("\n")

            chunks.append({
                "text": split_text,
                "file_path": rel_path,
                "start_line": start_line,
                "end_line": end_line,
            })

    return chunks


async def embed_texts(
    texts: list[str],
    ollama_base_url: str,
    model: str,
) -> list[list[float]]:
    """Embed texts using Ollama's /api/embed endpoint."""
    if not texts:
        return []

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{ollama_base_url}/api/embed",
            json={"model": model, "input": texts},
            timeout=120.0,
        )
        response.raise_for_status()
        return response.json()["embeddings"]


async def index_project(
    project_path: str,
    ollama_base_url: str,
    embedding_model: str,
    qdrant_host: str,
    qdrant_port: int,
) -> dict:
    """Index a Python project: collect files, chunk, embed, store in Qdrant."""
    project_name = os.path.basename(os.path.normpath(project_path))
    collection_name = f"debug-{project_name}"

    # Collect and chunk
    py_files = collect_python_files(project_path)
    chunks = chunk_code_files(py_files, project_root=project_path)

    if not chunks:
        return {"collection": collection_name, "files_indexed": 0, "chunks": 0}

    # Embed
    texts = [c["text"] for c in chunks]
    vectors = await embed_texts(texts, ollama_base_url, embedding_model)

    # Store in Qdrant (drop and recreate)
    qd = QdrantClient(host=qdrant_host, port=qdrant_port)
    if qd.collection_exists(collection_name):
        qd.delete_collection(collection_name)

    qd.create_collection(
        collection_name=collection_name,
        vectors_config=VectorParams(size=768, distance=Distance.COSINE),
    )

    points = [
        PointStruct(
            id=str(uuid.uuid4()),
            vector=vector,
            payload={
                "text": chunk["text"],
                "file_path": chunk["file_path"],
                "start_line": chunk["start_line"],
                "end_line": chunk["end_line"],
                "project_path": project_path,
            },
        )
        for chunk, vector in zip(chunks, vectors)
    ]
    qd.upsert(collection_name=collection_name, points=points)

    return {
        "collection": collection_name,
        "files_indexed": len(py_files),
        "chunks": len(chunks),
    }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/debug && python -m pytest tests/test_indexer.py -v`
Expected: 5 passed

- [ ] **Step 5: Run ruff checks**

Run: `cd services/debug && ruff check app/indexer.py tests/test_indexer.py && ruff format --check app/indexer.py tests/test_indexer.py`

- [ ] **Step 6: Commit**

```bash
git add services/debug/app/indexer.py services/debug/tests/test_indexer.py
git commit -m "feat(debug): add code indexer with language-aware Python chunking"
```

---

### Task 3: Tool Definitions and Execution

**Files:**
- Create: `services/debug/app/tools.py`
- Test: `services/debug/tests/test_tools.py`

- [ ] **Step 1: Write failing tests for tools**

```python
# services/debug/tests/test_tools.py
import os
import tempfile
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from app.tools import (
    TOOL_DEFINITIONS,
    execute_tool,
    tool_grep,
    tool_read_file,
    tool_run_tests,
    tool_search_code,
)


def test_tool_definitions_has_four_tools():
    assert len(TOOL_DEFINITIONS) == 4
    names = {t["function"]["name"] for t in TOOL_DEFINITIONS}
    assert names == {"search_code", "read_file", "grep", "run_tests"}


def test_tool_definitions_have_required_fields():
    for tool in TOOL_DEFINITIONS:
        assert tool["type"] == "function"
        func = tool["function"]
        assert "name" in func
        assert "description" in func
        assert "parameters" in func


def test_read_file_returns_content():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "test.py")
        open(filepath, "w").write("line1\nline2\nline3\nline4\nline5\n")

        result = tool_read_file(
            project_path=tmpdir,
            path="test.py",
            start_line=2,
            end_line=4,
        )
        assert "line2" in result
        assert "line3" in result
        assert "line4" in result
        assert "line1" not in result


def test_read_file_full_file():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "test.py")
        open(filepath, "w").write("line1\nline2\n")

        result = tool_read_file(project_path=tmpdir, path="test.py")
        assert "line1" in result
        assert "line2" in result


def test_read_file_missing_file():
    with tempfile.TemporaryDirectory() as tmpdir:
        result = tool_read_file(project_path=tmpdir, path="nonexistent.py")
        assert "not found" in result.lower() or "error" in result.lower()


def test_read_file_rejects_path_traversal():
    with tempfile.TemporaryDirectory() as tmpdir:
        result = tool_read_file(project_path=tmpdir, path="../../../etc/passwd")
        assert "error" in result.lower()


def test_grep_finds_matches():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "app.py")
        open(filepath, "w").write("def foo():\n    raise ValueError('bad')\n\ndef bar():\n    pass\n")

        result = tool_grep(project_path=tmpdir, pattern="raise", file_glob="*.py")
        assert "ValueError" in result
        assert "app.py" in result


def test_grep_no_matches():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "app.py")
        open(filepath, "w").write("def foo(): pass\n")

        result = tool_grep(project_path=tmpdir, pattern="nonexistent_string")
        assert "no matches" in result.lower()


def test_run_tests_success():
    with tempfile.TemporaryDirectory() as tmpdir:
        test_file = os.path.join(tmpdir, "test_example.py")
        open(test_file, "w").write("def test_pass():\n    assert True\n")

        result = tool_run_tests(project_path=tmpdir, target="test_example.py")
        assert "passed" in result.lower()


def test_run_tests_failure():
    with tempfile.TemporaryDirectory() as tmpdir:
        test_file = os.path.join(tmpdir, "test_example.py")
        open(test_file, "w").write("def test_fail():\n    assert False, 'intentional'\n")

        result = tool_run_tests(project_path=tmpdir, target="test_example.py")
        assert "failed" in result.lower()


@pytest.mark.asyncio
@patch("app.tools.QdrantClient")
@patch("app.tools.embed_texts", new_callable=AsyncMock)
async def test_search_code_returns_results(mock_embed, mock_qdrant_cls):
    mock_embed.return_value = [[0.1] * 768]

    mock_hit = MagicMock()
    mock_hit.payload = {
        "text": "def foo(): pass",
        "file_path": "app.py",
        "start_line": 1,
        "end_line": 1,
    }
    mock_hit.score = 0.95

    mock_qdrant = MagicMock()
    mock_qdrant.search.return_value = [mock_hit]
    mock_qdrant_cls.return_value = mock_qdrant

    result = await tool_search_code(
        query="foo function",
        collection="debug-myproject",
        ollama_base_url="http://localhost:11434",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
    )
    assert "app.py" in result
    assert "foo" in result


@pytest.mark.asyncio
async def test_execute_tool_dispatches_correctly():
    with tempfile.TemporaryDirectory() as tmpdir:
        filepath = os.path.join(tmpdir, "hello.py")
        open(filepath, "w").write("print('hello')\n")

        result = await execute_tool(
            tool_name="read_file",
            arguments={"path": "hello.py"},
            project_path=tmpdir,
            collection="debug-test",
            ollama_base_url="http://localhost:11434",
            embedding_model="nomic-embed-text",
            qdrant_host="localhost",
            qdrant_port=6333,
        )
        assert "hello" in result


@pytest.mark.asyncio
async def test_execute_tool_unknown_tool():
    result = await execute_tool(
        tool_name="unknown_tool",
        arguments={},
        project_path="/tmp",
        collection="test",
        ollama_base_url="http://localhost:11434",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
    )
    assert "unknown tool" in result.lower()
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/debug && python -m pytest tests/test_tools.py -v`
Expected: FAIL — `app.tools` does not exist

- [ ] **Step 3: Implement tools.py**

```python
# services/debug/app/tools.py
import os
import re
import subprocess

import httpx
from qdrant_client import QdrantClient

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "search_code",
            "description": (
                "Search the indexed codebase for code semantically similar to the query. "
                "Use this to find relevant functions, classes, or code patterns related to the bug."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Natural language description of the code to search for",
                    },
                },
                "required": ["query"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": (
                "Read a specific file from the project. Optionally specify a line range. "
                "Use this when you need to see more context around a code chunk found by search."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Relative path to the file from the project root",
                    },
                    "start_line": {
                        "type": "integer",
                        "description": "First line to read (1-indexed, optional)",
                    },
                    "end_line": {
                        "type": "integer",
                        "description": "Last line to read (1-indexed, optional)",
                    },
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "grep",
            "description": (
                "Search for a pattern (regex or literal string) across project files. "
                "Use this to find exact error messages, function calls, imports, or variable names."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "pattern": {
                        "type": "string",
                        "description": "Regex pattern or literal string to search for",
                    },
                    "file_glob": {
                        "type": "string",
                        "description": "Glob pattern to filter files (e.g. '*.py'). Defaults to '*.py'",
                    },
                },
                "required": ["pattern"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "run_tests",
            "description": (
                "Run pytest on a specific test file or test function. "
                "Use this to verify a hypothesis about the bug by running relevant tests."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "target": {
                        "type": "string",
                        "description": "Test file path relative to project root (e.g. 'tests/test_parser.py')",
                    },
                    "test_name": {
                        "type": "string",
                        "description": "Specific test function name (optional, e.g. 'test_empty_pdf')",
                    },
                },
                "required": ["target"],
            },
        },
    },
]


async def embed_texts(
    texts: list[str],
    ollama_base_url: str,
    model: str,
) -> list[list[float]]:
    """Embed texts using Ollama's /api/embed endpoint."""
    if not texts:
        return []

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{ollama_base_url}/api/embed",
            json={"model": model, "input": texts},
            timeout=120.0,
        )
        response.raise_for_status()
        return response.json()["embeddings"]


async def tool_search_code(
    query: str,
    collection: str,
    ollama_base_url: str,
    embedding_model: str,
    qdrant_host: str,
    qdrant_port: int,
    top_k: int = 5,
) -> str:
    """Vector similarity search over indexed codebase."""
    vectors = await embed_texts([query], ollama_base_url, embedding_model)
    query_vector = vectors[0]

    qd = QdrantClient(host=qdrant_host, port=qdrant_port)
    results = qd.search(
        collection_name=collection,
        query_vector=query_vector,
        limit=top_k,
    )

    if not results:
        return "No matching code found."

    parts = []
    for hit in results:
        p = hit.payload
        parts.append(
            f"--- {p['file_path']} (lines {p['start_line']}-{p['end_line']}, "
            f"score: {hit.score:.2f}) ---\n{p['text']}"
        )
    return "\n\n".join(parts)


def tool_read_file(
    project_path: str,
    path: str,
    start_line: int | None = None,
    end_line: int | None = None,
    max_lines: int = 100,
) -> str:
    """Read a file from the project directory."""
    # Prevent path traversal
    full_path = os.path.normpath(os.path.join(project_path, path))
    if not full_path.startswith(os.path.normpath(project_path)):
        return "Error: path traversal is not allowed."

    if not os.path.isfile(full_path):
        return f"Error: file not found: {path}"

    lines = open(full_path).readlines()

    if start_line is not None and end_line is not None:
        # Convert to 0-indexed
        selected = lines[start_line - 1 : end_line]
    else:
        selected = lines[:max_lines]

    if not selected:
        return f"Error: no content in specified range for {path}"

    numbered = []
    offset = (start_line or 1) - 1
    for i, line in enumerate(selected):
        numbered.append(f"{offset + i + 1:4d} | {line.rstrip()}")

    result = f"--- {path} ---\n" + "\n".join(numbered)
    if len(lines) > max_lines and start_line is None:
        result += f"\n... (truncated, {len(lines)} total lines)"
    return result


def tool_grep(
    project_path: str,
    pattern: str,
    file_glob: str = "*.py",
    max_matches: int = 20,
) -> str:
    """Search for a pattern across project files."""
    matches = []
    try:
        compiled = re.compile(pattern)
    except re.error:
        return f"Error: invalid regex pattern: {pattern}"

    for root, dirs, files in os.walk(project_path):
        # Skip hidden and venv directories
        dirs[:] = [
            d
            for d in dirs
            if not d.startswith(".") and d not in {"__pycache__", "venv", ".venv", "node_modules"}
        ]
        for filename in files:
            # Simple glob matching
            if file_glob == "*.py" and not filename.endswith(".py"):
                continue
            if file_glob and file_glob != "*.py":
                from fnmatch import fnmatch

                if not fnmatch(filename, file_glob):
                    continue

            filepath = os.path.join(root, filename)
            rel_path = os.path.relpath(filepath, project_path)

            try:
                with open(filepath) as f:
                    for line_num, line in enumerate(f, 1):
                        if compiled.search(line):
                            matches.append(f"{rel_path}:{line_num}: {line.rstrip()}")
                            if len(matches) >= max_matches:
                                break
            except (OSError, UnicodeDecodeError):
                continue

            if len(matches) >= max_matches:
                break
        if len(matches) >= max_matches:
            break

    if not matches:
        return f"No matches found for pattern: {pattern}"

    result = "\n".join(matches)
    if len(matches) >= max_matches:
        result += f"\n... (truncated at {max_matches} matches)"
    return result


def tool_run_tests(
    project_path: str,
    target: str,
    test_name: str | None = None,
    timeout: int = 30,
) -> str:
    """Run pytest on a specific test file or test function."""
    cmd = ["python", "-m", "pytest", target, "-v", "--tb=short", "--no-header"]
    if test_name:
        cmd[3] = f"{target}::{test_name}"

    try:
        result = subprocess.run(
            cmd,
            cwd=project_path,
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        output = result.stdout + result.stderr
        # Truncate long output
        lines = output.split("\n")
        if len(lines) > 50:
            output = "\n".join(lines[:50]) + f"\n... (truncated, {len(lines)} total lines)"
        return output
    except subprocess.TimeoutExpired:
        return f"Error: test timed out after {timeout}s"
    except FileNotFoundError:
        return "Error: pytest not found. Is it installed?"


async def execute_tool(
    tool_name: str,
    arguments: dict,
    project_path: str,
    collection: str,
    ollama_base_url: str,
    embedding_model: str,
    qdrant_host: str,
    qdrant_port: int,
) -> str:
    """Dispatch a tool call to the appropriate function."""
    if tool_name == "search_code":
        return await tool_search_code(
            query=arguments["query"],
            collection=collection,
            ollama_base_url=ollama_base_url,
            embedding_model=embedding_model,
            qdrant_host=qdrant_host,
            qdrant_port=qdrant_port,
        )
    elif tool_name == "read_file":
        return tool_read_file(
            project_path=project_path,
            path=arguments["path"],
            start_line=arguments.get("start_line"),
            end_line=arguments.get("end_line"),
        )
    elif tool_name == "grep":
        return tool_grep(
            project_path=project_path,
            pattern=arguments["pattern"],
            file_glob=arguments.get("file_glob", "*.py"),
        )
    elif tool_name == "run_tests":
        return tool_run_tests(
            project_path=project_path,
            target=arguments["target"],
            test_name=arguments.get("test_name"),
        )
    else:
        return f"Error: unknown tool '{tool_name}'"
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/debug && python -m pytest tests/test_tools.py -v`
Expected: All passed

- [ ] **Step 5: Run ruff checks**

Run: `cd services/debug && ruff check app/tools.py tests/test_tools.py && ruff format --check app/tools.py tests/test_tools.py`

- [ ] **Step 6: Commit**

```bash
git add services/debug/app/tools.py services/debug/tests/test_tools.py
git commit -m "feat(debug): add four agent tools — search, read, grep, test runner"
```

---

### Task 4: Prompt Templates

**Files:**
- Create: `services/debug/app/prompts.py`
- Test: `services/debug/tests/test_prompts.py`

- [ ] **Step 1: Write failing tests for prompts**

```python
# services/debug/tests/test_prompts.py
from app.prompts import SYSTEM_PROMPT, build_user_prompt, build_duplicate_nudge


def test_system_prompt_mentions_tools():
    assert "search_code" in SYSTEM_PROMPT
    assert "read_file" in SYSTEM_PROMPT
    assert "grep" in SYSTEM_PROMPT
    assert "run_tests" in SYSTEM_PROMPT


def test_system_prompt_mentions_diagnosis():
    assert "diagnosis" in SYSTEM_PROMPT.lower() or "diagnose" in SYSTEM_PROMPT.lower()


def test_build_user_prompt_with_error():
    prompt = build_user_prompt(
        description="upload returns 500",
        error_output="Traceback: ValueError in parser.py",
    )
    assert "upload returns 500" in prompt
    assert "Traceback" in prompt


def test_build_user_prompt_without_error():
    prompt = build_user_prompt(
        description="upload returns 500",
        error_output=None,
    )
    assert "upload returns 500" in prompt
    assert "error output" not in prompt.lower() or "none" in prompt.lower() or "no error" in prompt.lower()


def test_build_duplicate_nudge():
    nudge = build_duplicate_nudge("search_code", '{"query": "foo"}')
    assert "search_code" in nudge
    assert "different" in nudge.lower() or "another" in nudge.lower()
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/debug && python -m pytest tests/test_prompts.py -v`
Expected: FAIL — `app.prompts` does not exist

- [ ] **Step 3: Implement prompts.py**

```python
# services/debug/app/prompts.py

SYSTEM_PROMPT = """\
You are a debugging assistant. Your job is to diagnose bugs in Python codebases.

You have access to these tools:
- search_code: Semantic search over the indexed codebase. Use this first to find relevant code.
- read_file: Read a specific file or line range. Use after search to see more context.
- grep: Search for exact strings or regex patterns across the project. Use for error messages, \
function names, and imports.
- run_tests: Run pytest on a specific test file or function. Use to verify hypotheses.

Debugging process:
1. Analyze the bug description and any error output
2. Use search_code to find relevant code
3. Use read_file and grep to understand the code in context
4. Form a hypothesis about the root cause
5. Use run_tests to verify your hypothesis if applicable
6. When confident, provide your diagnosis

When you are ready to give your final diagnosis, respond with text (not a tool call) that includes:
- Root cause: What is causing the bug
- Location: Which file(s) and line(s) are involved
- Suggestion: How to fix it

Be methodical. Use tools to gather evidence before concluding. Do not guess without evidence."""

USER_PROMPT_TEMPLATE = """\
Bug description: {description}

Error output:
{error_output}

Please investigate this bug. Start by searching for relevant code."""

USER_PROMPT_NO_ERROR_TEMPLATE = """\
Bug description: {description}

No error output was provided. Please investigate this bug. \
Start by searching for relevant code."""


def build_user_prompt(description: str, error_output: str | None = None) -> str:
    """Build the initial user message for the agent."""
    if error_output and error_output.strip():
        return USER_PROMPT_TEMPLATE.format(
            description=description,
            error_output=error_output,
        )
    return USER_PROMPT_NO_ERROR_TEMPLATE.format(description=description)


def build_duplicate_nudge(tool_name: str, arguments: str) -> str:
    """Build a message nudging the agent to try a different approach."""
    return (
        f"You already called {tool_name} with these same arguments: {arguments}. "
        "Try a different tool or different arguments to make progress."
    )
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/debug && python -m pytest tests/test_prompts.py -v`
Expected: 5 passed

- [ ] **Step 5: Run ruff checks**

Run: `cd services/debug && ruff check app/prompts.py tests/test_prompts.py && ruff format --check app/prompts.py tests/test_prompts.py`

- [ ] **Step 6: Commit**

```bash
git add services/debug/app/prompts.py services/debug/tests/test_prompts.py
git commit -m "feat(debug): add agent system prompt and user prompt templates"
```

---

### Task 5: Agent Loop

**Files:**
- Create: `services/debug/app/agent.py`
- Test: `services/debug/tests/test_agent.py`

- [ ] **Step 1: Write failing tests for the agent loop**

```python
# services/debug/tests/test_agent.py
import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from app.agent import run_agent_loop


@pytest.mark.asyncio
@patch("app.agent.execute_tool", new_callable=AsyncMock)
@patch("app.agent.call_ollama", new_callable=AsyncMock)
async def test_agent_returns_diagnosis_on_text_response(mock_ollama, mock_tool):
    """Agent should yield a diagnosis event when Ollama returns plain text."""
    mock_ollama.return_value = {
        "message": {
            "role": "assistant",
            "content": "Root cause: missing null check in parser.py line 42",
        }
    }

    events = []
    async for event in run_agent_loop(
        description="upload fails",
        error_output=None,
        collection="debug-test",
        project_path="/tmp/test",
        ollama_base_url="http://localhost:11434",
        chat_model="qwen2.5:14b",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        max_steps=10,
    ):
        events.append(event)

    event_types = [e["event"] for e in events]
    assert "diagnosis" in event_types
    assert "done" in event_types
    diagnosis = next(e for e in events if e["event"] == "diagnosis")
    assert "null check" in diagnosis["data"]["content"]


@pytest.mark.asyncio
@patch("app.agent.execute_tool", new_callable=AsyncMock)
@patch("app.agent.call_ollama", new_callable=AsyncMock)
async def test_agent_executes_tool_calls(mock_ollama, mock_tool):
    """Agent should execute tool calls and loop back."""
    # First call: tool call. Second call: diagnosis.
    mock_ollama.side_effect = [
        {
            "message": {
                "role": "assistant",
                "content": "",
                "tool_calls": [
                    {
                        "function": {
                            "name": "search_code",
                            "arguments": {"query": "upload handler"},
                        }
                    }
                ],
            }
        },
        {
            "message": {
                "role": "assistant",
                "content": "Root cause: the upload handler doesn't validate file type",
            }
        },
    ]
    mock_tool.return_value = "Found: app/upload.py def handle_upload()..."

    events = []
    async for event in run_agent_loop(
        description="upload fails",
        error_output=None,
        collection="debug-test",
        project_path="/tmp/test",
        ollama_base_url="http://localhost:11434",
        chat_model="qwen2.5:14b",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        max_steps=10,
    ):
        events.append(event)

    event_types = [e["event"] for e in events]
    assert "tool_call" in event_types
    assert "tool_result" in event_types
    assert "diagnosis" in event_types
    mock_tool.assert_called_once()


@pytest.mark.asyncio
@patch("app.agent.execute_tool", new_callable=AsyncMock)
@patch("app.agent.call_ollama", new_callable=AsyncMock)
async def test_agent_stops_at_max_steps(mock_ollama, mock_tool):
    """Agent should stop after max_steps and emit a diagnosis."""
    # Always return tool calls — should hit the max step limit
    mock_ollama.return_value = {
        "message": {
            "role": "assistant",
            "content": "",
            "tool_calls": [
                {
                    "function": {
                        "name": "grep",
                        "arguments": {"pattern": "error"},
                    }
                }
            ],
        }
    }
    mock_tool.return_value = "app.py:10: raise ValueError('error')"

    events = []
    async for event in run_agent_loop(
        description="bug",
        error_output=None,
        collection="debug-test",
        project_path="/tmp/test",
        ollama_base_url="http://localhost:11434",
        chat_model="qwen2.5:14b",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        max_steps=3,
    ):
        events.append(event)

    event_types = [e["event"] for e in events]
    assert "done" in event_types
    # Should have at most 3 tool call events
    tool_calls = [e for e in events if e["event"] == "tool_call"]
    assert len(tool_calls) <= 3


@pytest.mark.asyncio
@patch("app.agent.execute_tool", new_callable=AsyncMock)
@patch("app.agent.call_ollama", new_callable=AsyncMock)
async def test_agent_detects_duplicate_tool_calls(mock_ollama, mock_tool):
    """Agent should nudge when the same tool+args are called twice."""
    # Return same tool call twice, then diagnosis
    mock_ollama.side_effect = [
        {
            "message": {
                "role": "assistant",
                "content": "",
                "tool_calls": [
                    {"function": {"name": "grep", "arguments": {"pattern": "error"}}}
                ],
            }
        },
        {
            "message": {
                "role": "assistant",
                "content": "",
                "tool_calls": [
                    {"function": {"name": "grep", "arguments": {"pattern": "error"}}}
                ],
            }
        },
        {
            "message": {
                "role": "assistant",
                "content": "Root cause: found the issue",
            }
        },
    ]
    mock_tool.return_value = "app.py:10: raise ValueError"

    events = []
    async for event in run_agent_loop(
        description="bug",
        error_output=None,
        collection="debug-test",
        project_path="/tmp/test",
        ollama_base_url="http://localhost:11434",
        chat_model="qwen2.5:14b",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        max_steps=10,
    ):
        events.append(event)

    # The agent should still complete (nudge doesn't block)
    event_types = [e["event"] for e in events]
    assert "diagnosis" in event_types
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/debug && python -m pytest tests/test_agent.py -v`
Expected: FAIL — `app.agent` does not exist

- [ ] **Step 3: Implement agent.py**

```python
# services/debug/app/agent.py
import json
import logging
from collections.abc import AsyncGenerator

import httpx

from app.prompts import SYSTEM_PROMPT, build_duplicate_nudge, build_user_prompt
from app.tools import TOOL_DEFINITIONS, execute_tool

logger = logging.getLogger(__name__)


async def call_ollama(
    messages: list[dict],
    model: str,
    base_url: str,
    tools: list[dict] | None = None,
) -> dict:
    """Make a single call to Ollama's /api/chat endpoint."""
    payload = {
        "model": model,
        "messages": messages,
        "stream": False,
    }
    if tools:
        payload["tools"] = tools

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{base_url}/api/chat",
            json=payload,
            timeout=300.0,
        )
        response.raise_for_status()
        return response.json()


async def run_agent_loop(
    description: str,
    error_output: str | None,
    collection: str,
    project_path: str,
    ollama_base_url: str,
    chat_model: str,
    embedding_model: str,
    qdrant_host: str,
    qdrant_port: int,
    max_steps: int = 10,
) -> AsyncGenerator[dict, None]:
    """Run the agent loop: call Ollama → parse → execute tool → repeat."""
    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": build_user_prompt(description, error_output)},
    ]

    seen_calls: set[str] = set()
    step = 0

    for _ in range(max_steps):
        step += 1

        try:
            response = await call_ollama(
                messages=messages,
                model=chat_model,
                base_url=ollama_base_url,
                tools=TOOL_DEFINITIONS,
            )
        except Exception as e:
            logger.error("Ollama call failed: %s", e, exc_info=True)
            yield {
                "event": "diagnosis",
                "data": {"step": step, "content": f"Error communicating with LLM: {e}"},
            }
            yield {"event": "done", "data": {}}
            return

        msg = response["message"]
        tool_calls = msg.get("tool_calls")

        # No tool calls = final diagnosis
        if not tool_calls:
            content = msg.get("content", "Unable to determine the root cause.")
            yield {
                "event": "diagnosis",
                "data": {"step": step, "content": content},
            }
            yield {"event": "done", "data": {}}
            return

        # Process tool call
        tool_call = tool_calls[0]
        func = tool_call["function"]
        tool_name = func["name"]
        arguments = func["arguments"]
        if isinstance(arguments, str):
            arguments = json.loads(arguments)

        # Emit thinking if assistant included content
        if msg.get("content"):
            yield {
                "event": "thinking",
                "data": {"step": step, "content": msg["content"]},
            }

        # Check for duplicate calls
        call_key = json.dumps({"name": tool_name, "args": arguments}, sort_keys=True)
        if call_key in seen_calls:
            nudge = build_duplicate_nudge(tool_name, json.dumps(arguments))
            messages.append(msg)
            messages.append({"role": "user", "content": nudge})
            continue
        seen_calls.add(call_key)

        # Emit tool call event
        yield {
            "event": "tool_call",
            "data": {"step": step, "tool": tool_name, "args": arguments},
        }

        # Execute the tool
        result = await execute_tool(
            tool_name=tool_name,
            arguments=arguments,
            project_path=project_path,
            collection=collection,
            ollama_base_url=ollama_base_url,
            embedding_model=embedding_model,
            qdrant_host=qdrant_host,
            qdrant_port=qdrant_port,
        )

        # Emit tool result event
        truncated = len(result) > 2000
        display_result = result[:2000] + "..." if truncated else result
        yield {
            "event": "tool_result",
            "data": {
                "step": step,
                "tool": tool_name,
                "result": display_result,
                "truncated": truncated,
            },
        }

        # Add to conversation history
        messages.append(msg)
        messages.append({"role": "tool", "content": result})

    # Max steps reached — ask for final diagnosis without tools
    messages.append({
        "role": "user",
        "content": (
            "You have reached the maximum number of tool calls. "
            "Based on everything you have found so far, provide your diagnosis now."
        ),
    })

    try:
        response = await call_ollama(
            messages=messages,
            model=chat_model,
            base_url=ollama_base_url,
            tools=None,
        )
        content = response["message"].get("content", "Unable to determine the root cause.")
    except Exception as e:
        content = f"Error getting final diagnosis: {e}"

    yield {
        "event": "diagnosis",
        "data": {"step": step + 1, "content": content},
    }
    yield {"event": "done", "data": {}}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/debug && python -m pytest tests/test_agent.py -v`
Expected: 4 passed

- [ ] **Step 5: Run ruff checks**

Run: `cd services/debug && ruff check app/agent.py tests/test_agent.py && ruff format --check app/agent.py tests/test_agent.py`

- [ ] **Step 6: Commit**

```bash
git add services/debug/app/agent.py services/debug/tests/test_agent.py
git commit -m "feat(debug): implement custom agent loop with tool calling and guardrails"
```

---

### Task 6: API Endpoints — /index and /debug

**Files:**
- Modify: `services/debug/app/main.py`
- Modify: `services/debug/tests/test_main.py`

- [ ] **Step 1: Write failing tests for /index and /debug endpoints**

Add the following tests to `services/debug/tests/test_main.py`:

```python
# Append to services/debug/tests/test_main.py

@patch("app.main.index_project", new_callable=AsyncMock)
def test_index_success(mock_index):
    mock_index.return_value = {
        "collection": "debug-myproject",
        "files_indexed": 5,
        "chunks": 42,
    }

    response = client.post("/index", json={"path": "/tmp/myproject"})
    assert response.status_code == 200
    data = response.json()
    assert data["collection"] == "debug-myproject"
    assert data["files_indexed"] == 5
    assert data["chunks"] == 42


def test_index_missing_path():
    response = client.post("/index", json={})
    assert response.status_code == 422


@patch("app.main.index_project", new_callable=AsyncMock)
def test_index_nonexistent_path(mock_index):
    mock_index.side_effect = FileNotFoundError("path not found")

    response = client.post("/index", json={"path": "/nonexistent"})
    assert response.status_code == 400


@patch("app.main.run_agent_loop")
def test_debug_streams_sse_events(mock_agent):
    async def fake_events(*args, **kwargs):
        yield {"event": "thinking", "data": {"step": 1, "content": "Analyzing..."}}
        yield {
            "event": "diagnosis",
            "data": {"step": 2, "content": "Root cause: missing check"},
        }
        yield {"event": "done", "data": {}}

    mock_agent.return_value = fake_events()

    response = client.post(
        "/debug",
        json={
            "collection": "debug-test",
            "description": "upload fails",
        },
    )

    assert response.status_code == 200
    assert "text/event-stream" in response.headers["content-type"]

    events = []
    for line in response.text.strip().split("\n"):
        if line.startswith("data: "):
            events.append(json.loads(line[6:]))

    assert any("thinking" in str(e) or "Analyzing" in str(e) for e in events)


def test_debug_missing_collection():
    response = client.post("/debug", json={"description": "bug"})
    assert response.status_code == 422


def test_debug_missing_description():
    response = client.post("/debug", json={"collection": "test"})
    assert response.status_code == 422
```

- [ ] **Step 2: Run tests to verify new tests fail**

Run: `cd services/debug && python -m pytest tests/test_main.py -v`
Expected: New tests fail — endpoints don't exist yet

- [ ] **Step 3: Update main.py with /index and /debug endpoints**

Replace the full `services/debug/app/main.py` with:

```python
# services/debug/app/main.py
import json
import logging
import os

import httpx
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from qdrant_client import QdrantClient
from sse_starlette.sse import EventSourceResponse

from app.agent import run_agent_loop
from app.config import settings
from app.indexer import index_project

logger = logging.getLogger(__name__)

app = FastAPI(title="Debug Assistant API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins.split(","),
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)


class IndexRequest(BaseModel):
    path: str


class DebugRequest(BaseModel):
    collection: str = Field(pattern=r"^[a-zA-Z0-9_-]{1,100}$")
    description: str = Field(max_length=5000)
    error_output: str | None = Field(default=None, max_length=10000)


# Store project paths by collection name for tool execution
_project_paths: dict[str, str] = {}


@app.get("/health")
async def health():
    try:
        qd = QdrantClient(host=settings.qdrant_host, port=settings.qdrant_port)
        qd.get_collections()
    except Exception:
        logger.error("Qdrant health check failed", exc_info=True)
        return JSONResponse(
            status_code=503,
            content={"status": "unhealthy", "detail": "Qdrant unavailable"},
        )

    try:
        async with httpx.AsyncClient() as client:
            resp = await client.get(f"{settings.ollama_base_url}/api/tags", timeout=5.0)
            resp.raise_for_status()
    except Exception:
        logger.error("Ollama health check failed", exc_info=True)
        return JSONResponse(
            status_code=503,
            content={"status": "unhealthy", "detail": "Ollama unavailable"},
        )

    return {"status": "ok"}


@app.post("/index")
async def index(request: IndexRequest):
    if not os.path.isdir(request.path):
        raise HTTPException(status_code=400, detail=f"Directory not found: {request.path}")

    try:
        result = await index_project(
            project_path=request.path,
            ollama_base_url=settings.ollama_base_url,
            embedding_model=settings.embedding_model,
            qdrant_host=settings.qdrant_host,
            qdrant_port=settings.qdrant_port,
        )
    except Exception as e:
        logger.error("Indexing failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="Indexing failed")

    # Remember project path for debug sessions
    _project_paths[result["collection"]] = request.path

    return result


@app.post("/debug")
async def debug(request: DebugRequest):
    project_path = _project_paths.get(request.collection)
    if not project_path:
        raise HTTPException(
            status_code=400,
            detail=f"Collection '{request.collection}' not indexed. Call /index first.",
        )

    async def event_generator():
        try:
            async for event in run_agent_loop(
                description=request.description,
                error_output=request.error_output,
                collection=request.collection,
                project_path=project_path,
                ollama_base_url=settings.ollama_base_url,
                chat_model=settings.chat_model,
                embedding_model=settings.embedding_model,
                qdrant_host=settings.qdrant_host,
                qdrant_port=settings.qdrant_port,
                max_steps=settings.max_agent_steps,
            ):
                yield {
                    "event": event["event"],
                    "data": json.dumps(event["data"]),
                }
        except Exception as e:
            logger.error("Debug session error: %s", e, exc_info=True)
            yield {
                "event": "diagnosis",
                "data": json.dumps({"content": "Internal error during debug session."}),
            }
            yield {"event": "done", "data": json.dumps({})}

    return EventSourceResponse(event_generator())
```

- [ ] **Step 4: Update test_main.py to handle the /debug endpoint's project path requirement**

The `/debug` endpoint requires a previously indexed collection. Update the debug test:

```python
# Replace the test_debug_streams_sse_events test with:
@patch("app.main.run_agent_loop")
def test_debug_streams_sse_events(mock_agent):
    # Register a project path so /debug can find it
    from app.main import _project_paths
    _project_paths["debug-test"] = "/tmp/test"

    async def fake_events(*args, **kwargs):
        yield {"event": "thinking", "data": {"step": 1, "content": "Analyzing..."}}
        yield {
            "event": "diagnosis",
            "data": {"step": 2, "content": "Root cause: missing check"},
        }
        yield {"event": "done", "data": {}}

    mock_agent.return_value = fake_events()

    response = client.post(
        "/debug",
        json={
            "collection": "debug-test",
            "description": "upload fails",
        },
    )

    assert response.status_code == 200
    assert "text/event-stream" in response.headers["content-type"]

    events = []
    for line in response.text.strip().split("\n"):
        if line.startswith("data: "):
            events.append(json.loads(line[6:]))

    assert len(events) >= 2

    # Clean up
    _project_paths.pop("debug-test", None)
```

- [ ] **Step 5: Run all tests**

Run: `cd services/debug && python -m pytest tests/ -v`
Expected: All passed

- [ ] **Step 6: Run ruff checks**

Run: `cd services/debug && ruff check app/ tests/ && ruff format --check app/ tests/`

- [ ] **Step 7: Commit**

```bash
git add services/debug/app/main.py services/debug/tests/test_main.py
git commit -m "feat(debug): add /index and /debug SSE endpoints"
```

---

### Task 7: Docker Compose and Environment

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`

- [ ] **Step 1: Add debug service to docker-compose.yml**

Add after the `chat` service block:

```yaml
  debug:
    image: ghcr.io/kabradshaw1/gen_ai_engineer/debug:latest
    build:
      context: ./services/debug
    ports:
      - "8003:8000"
    env_file: .env
    depends_on:
      qdrant:
        condition: service_healthy
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

- [ ] **Step 2: Update .env.example with debug-specific settings**

Add to `.env.example`:

```
# Debug Assistant
CHAT_MODEL=qwen2.5:14b
MAX_AGENT_STEPS=10
TEST_TIMEOUT_SECONDS=30
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "feat(debug): add debug service to Docker Compose"
```

---

### Task 8: CI Pipeline Integration

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Read the current CI workflow to find the exact matrix locations**

Run: `cat .github/workflows/ci.yml` and identify:
- `backend-tests.strategy.matrix.service` — add `debug`
- `docker-build.strategy.matrix.service` — add `debug`
- `security-pip-audit.strategy.matrix.service` — add `debug`
- `security-hadolint.strategy.matrix.dockerfile` — add `services/debug/Dockerfile`
- Deploy step — add `debug` to `docker compose pull` command

- [ ] **Step 2: Add debug to all CI matrices**

Add `debug` to each matrix array and the deploy command. Follow the checklist in `services/CLAUDE.md`.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add debug service to test, build, security, and deploy matrices"
```

---

### Task 9: Frontend — Debug Form Component

**Files:**
- Create: `frontend/src/components/DebugForm.tsx`

- [ ] **Step 1: Create DebugForm component**

```tsx
// frontend/src/components/DebugForm.tsx
"use client";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { useState } from "react";

interface DebugFormProps {
  onSubmit: (data: {
    collection: string;
    description: string;
    errorOutput?: string;
  }) => void;
  isLoading: boolean;
}

export function DebugForm({ onSubmit, isLoading }: DebugFormProps) {
  const [projectPath, setProjectPath] = useState("");
  const [description, setDescription] = useState("");
  const [errorOutput, setErrorOutput] = useState("");
  const [indexing, setIndexing] = useState(false);
  const [collection, setCollection] = useState<string | null>(null);
  const [indexStatus, setIndexStatus] = useState<string | null>(null);

  const debugBaseUrl =
    process.env.NEXT_PUBLIC_DEBUG_API_URL || "http://localhost:8003";

  const handleIndex = async () => {
    setIndexing(true);
    setIndexStatus(null);

    try {
      const res = await fetch(`${debugBaseUrl}/index`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: projectPath }),
      });

      if (!res.ok) {
        const data = await res.json();
        setIndexStatus(`Error: ${data.detail || "Indexing failed"}`);
        return;
      }

      const data = await res.json();
      setCollection(data.collection);
      setIndexStatus(
        `Indexed ${data.files_indexed} files (${data.chunks} chunks)`
      );
    } catch {
      setIndexStatus("Error: Could not connect to debug service");
    } finally {
      setIndexing(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!collection || !description.trim()) return;

    onSubmit({
      collection,
      description: description.trim(),
      errorOutput: errorOutput.trim() || undefined,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="text-sm font-medium text-foreground">
          Project Path
        </label>
        <div className="mt-1 flex gap-2">
          <Input
            value={projectPath}
            onChange={(e) => setProjectPath(e.target.value)}
            placeholder="/path/to/python/project"
            disabled={indexing}
          />
          <Button
            type="button"
            variant="outline"
            onClick={handleIndex}
            disabled={indexing || !projectPath.trim()}
          >
            {indexing ? "Indexing..." : "Index"}
          </Button>
        </div>
        {indexStatus && (
          <p
            className={`mt-1 text-sm ${indexStatus.startsWith("Error") ? "text-destructive" : "text-muted-foreground"}`}
          >
            {indexStatus}
          </p>
        )}
      </div>

      <div>
        <label className="text-sm font-medium text-foreground">
          Bug Description
        </label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describe the bug you're investigating..."
          className="mt-1"
          rows={4}
          disabled={isLoading}
        />
      </div>

      <div>
        <label className="text-sm font-medium text-foreground">
          Error Output{" "}
          <span className="text-muted-foreground font-normal">(optional)</span>
        </label>
        <Textarea
          value={errorOutput}
          onChange={(e) => setErrorOutput(e.target.value)}
          placeholder="Paste stack trace, error logs, or test output..."
          className="mt-1 font-mono text-sm"
          rows={6}
          disabled={isLoading}
        />
      </div>

      <Button
        type="submit"
        disabled={isLoading || !collection || !description.trim()}
        className="w-full"
      >
        {isLoading ? "Debugging..." : "Debug"}
      </Button>
    </form>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/DebugForm.tsx
git commit -m "feat(frontend): add DebugForm component for bug description input"
```

---

### Task 10: Frontend — Agent Timeline Components

**Files:**
- Create: `frontend/src/components/ToolCallCard.tsx`
- Create: `frontend/src/components/DiagnosisBanner.tsx`
- Create: `frontend/src/components/AgentTimeline.tsx`

- [ ] **Step 1: Create ToolCallCard component**

```tsx
// frontend/src/components/ToolCallCard.tsx
"use client";

import { useState } from "react";

interface ToolCallCardProps {
  step: number;
  tool: string;
  args: Record<string, unknown>;
  result?: string;
  truncated?: boolean;
}

export function ToolCallCard({
  step,
  tool,
  args,
  result,
  truncated,
}: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="rounded-lg border border-foreground/10 bg-card">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between px-4 py-2 text-left text-sm"
      >
        <span>
          <span className="text-muted-foreground">Step {step}:</span>{" "}
          <span className="font-mono font-medium">{tool}</span>
          <span className="text-muted-foreground">
            ({Object.values(args).map(String).join(", ")})
          </span>
        </span>
        <span className="text-muted-foreground">{expanded ? "−" : "+"}</span>
      </button>

      {expanded && (
        <div className="border-t border-foreground/10 px-4 py-3 space-y-2">
          <div>
            <p className="text-xs font-medium text-muted-foreground">
              Arguments
            </p>
            <pre className="mt-1 text-xs font-mono whitespace-pre-wrap text-foreground/80">
              {JSON.stringify(args, null, 2)}
            </pre>
          </div>
          {result && (
            <div>
              <p className="text-xs font-medium text-muted-foreground">
                Result
                {truncated && (
                  <span className="text-yellow-500"> (truncated)</span>
                )}
              </p>
              <pre className="mt-1 max-h-60 overflow-y-auto text-xs font-mono whitespace-pre-wrap text-foreground/80">
                {result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Create DiagnosisBanner component**

```tsx
// frontend/src/components/DiagnosisBanner.tsx
interface DiagnosisBannerProps {
  content: string;
}

export function DiagnosisBanner({ content }: DiagnosisBannerProps) {
  return (
    <div className="rounded-lg border-2 border-green-500/30 bg-green-500/5 p-4">
      <h3 className="text-sm font-semibold text-green-400">Diagnosis</h3>
      <div className="mt-2 text-sm text-foreground whitespace-pre-wrap">
        {content}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create AgentTimeline component**

```tsx
// frontend/src/components/AgentTimeline.tsx
import { DiagnosisBanner } from "./DiagnosisBanner";
import { ToolCallCard } from "./ToolCallCard";

export interface AgentEvent {
  event: "thinking" | "tool_call" | "tool_result" | "diagnosis" | "done";
  data: {
    step?: number;
    content?: string;
    tool?: string;
    args?: Record<string, unknown>;
    result?: string;
    truncated?: boolean;
  };
}

interface AgentTimelineProps {
  events: AgentEvent[];
}

export function AgentTimeline({ events }: AgentTimelineProps) {
  if (events.length === 0) {
    return (
      <div className="flex h-full items-center justify-center text-muted-foreground">
        Submit a bug to start debugging
      </div>
    );
  }

  // Pair tool_call events with their results
  const toolResults = new Map<number, { result: string; truncated: boolean }>();
  for (const evt of events) {
    if (
      evt.event === "tool_result" &&
      evt.data.step !== undefined &&
      evt.data.result
    ) {
      toolResults.set(evt.data.step, {
        result: evt.data.result,
        truncated: evt.data.truncated ?? false,
      });
    }
  }

  return (
    <div className="space-y-3">
      {events.map((evt, i) => {
        if (evt.event === "thinking" && evt.data.content) {
          return (
            <p key={i} className="text-sm italic text-muted-foreground">
              {evt.data.content}
            </p>
          );
        }

        if (
          evt.event === "tool_call" &&
          evt.data.tool &&
          evt.data.args &&
          evt.data.step !== undefined
        ) {
          const result = toolResults.get(evt.data.step);
          return (
            <ToolCallCard
              key={i}
              step={evt.data.step}
              tool={evt.data.tool}
              args={evt.data.args}
              result={result?.result}
              truncated={result?.truncated}
            />
          );
        }

        if (evt.event === "diagnosis" && evt.data.content) {
          return <DiagnosisBanner key={i} content={evt.data.content} />;
        }

        // Skip tool_result and done events (handled inline above)
        return null;
      })}
    </div>
  );
}
```

- [ ] **Step 4: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ToolCallCard.tsx frontend/src/components/DiagnosisBanner.tsx frontend/src/components/AgentTimeline.tsx
git commit -m "feat(frontend): add AgentTimeline, ToolCallCard, and DiagnosisBanner components"
```

---

### Task 11: Frontend — Debug Page

**Files:**
- Create: `frontend/src/app/ai/debug/page.tsx`

- [ ] **Step 1: Create the debug page**

```tsx
// frontend/src/app/ai/debug/page.tsx
"use client";

import Link from "next/link";
import { useCallback, useRef, useState } from "react";
import { AgentTimeline, AgentEvent } from "@/components/AgentTimeline";
import { DebugForm } from "@/components/DebugForm";

const debugBaseUrl =
  process.env.NEXT_PUBLIC_DEBUG_API_URL || "http://localhost:8003";

export default function DebugPage() {
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const timelineRef = useRef<HTMLDivElement>(null);

  const handleSubmit = useCallback(
    async (data: {
      collection: string;
      description: string;
      errorOutput?: string;
    }) => {
      setEvents([]);
      setIsLoading(true);

      try {
        const res = await fetch(`${debugBaseUrl}/debug`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            collection: data.collection,
            description: data.description,
            error_output: data.errorOutput,
          }),
        });

        if (!res.ok) {
          const err = await res.json();
          setEvents([
            {
              event: "diagnosis",
              data: { content: `Error: ${err.detail || "Debug request failed"}` },
            },
          ]);
          return;
        }

        const reader = res.body?.getReader();
        if (!reader) return;

        const decoder = new TextDecoder();
        let buffer = "";
        let currentEventType = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            if (line.startsWith("event: ")) {
              currentEventType = line.slice(7).trim();
              continue;
            }

            if (!line.startsWith("data: ")) continue;
            const jsonStr = line.slice(6).trim();
            if (!jsonStr) continue;

            try {
              const data = JSON.parse(jsonStr);
              const eventType =
                currentEventType || inferEventType(data);

              const evt: AgentEvent = {
                event: eventType as AgentEvent["event"],
                data,
              };

              setEvents((prev) => [...prev, evt]);
              currentEventType = "";

              // Auto-scroll
              if (timelineRef.current) {
                timelineRef.current.scrollTop =
                  timelineRef.current.scrollHeight;
              }
            } catch {
              // skip malformed lines
            }
          }
        }
      } catch {
        setEvents((prev) => [
          ...prev,
          {
            event: "diagnosis",
            data: { content: "Error: Could not connect to debug service" },
          },
        ]);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <Link
          href="/ai"
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          &larr; AI Projects
        </Link>

        <h1 className="mt-8 text-3xl font-bold">Debug Assistant</h1>
        <p className="mt-2 text-muted-foreground">
          Describe a bug and watch the AI agent investigate your codebase
          step-by-step.
        </p>

        <div className="mt-8 grid grid-cols-1 gap-8 lg:grid-cols-2">
          {/* Left panel: Input form */}
          <div>
            <DebugForm onSubmit={handleSubmit} isLoading={isLoading} />
          </div>

          {/* Right panel: Agent timeline */}
          <div
            ref={timelineRef}
            className="max-h-[70vh] overflow-y-auto rounded-xl border border-foreground/10 bg-card p-4"
          >
            <AgentTimeline events={events} />
          </div>
        </div>
      </div>
    </div>
  );
}

function inferEventType(
  data: Record<string, unknown>
): AgentEvent["event"] {
  if ("tool" in data && "args" in data) return "tool_call";
  if ("tool" in data && "result" in data) return "tool_result";
  if ("content" in data && !("tool" in data) && !("step" in data))
    return "diagnosis";
  return "thinking";
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/ai/debug/page.tsx
git commit -m "feat(frontend): add /ai/debug page with SSE streaming and agent timeline"
```

---

### Task 12: Frontend — Update /ai Hub Page

**Files:**
- Modify: `frontend/src/app/ai/page.tsx`

- [ ] **Step 1: Add Debug Assistant section and update model reference**

Update `frontend/src/app/ai/page.tsx`:

1. Update the existing architecture diagram to reference "Qwen 2.5 14B" instead of "Mistral 7B"

2. Add a new Mermaid diagram for the debug assistant workflow:

```typescript
const debugDiagram = `flowchart LR
  subgraph Index["Code Indexing"]
    direction LR
    A[Python Project] --> B[Walk Files]
    B --> C[Chunk\nLanguage.PYTHON]
    C --> D[Embed\nnomic-embed-text]
    D --> E[(Qdrant)]
  end

  subgraph Agent["Agent Loop"]
    direction LR
    F[Bug Description] --> G[Call LLM\nQwen 2.5 14B]
    G --> H{Tool Call?}
    H -->|Yes| I[Execute Tool]
    I --> G
    H -->|No| J[Stream Diagnosis]
  end
`;
```

3. Add the Debug Assistant section after the Document Q&A demo link:

```tsx
        {/* Debug Assistant Project */}
        <section className="mt-16">
          <h2 className="text-2xl font-semibold">Debug Assistant</h2>
          <p className="mt-4 text-muted-foreground leading-relaxed">
            An agentic debugging tool that investigates bugs in Python
            codebases. Given a bug description and optional error output, the
            agent searches indexed code, reads files, runs grep, and executes
            tests in a multi-step loop to diagnose the root cause — streaming
            its reasoning in real-time.
          </p>

          <h3 className="mt-6 text-lg font-medium">Tech Stack</h3>
          <ul className="mt-2 list-disc pl-6 text-muted-foreground space-y-1">
            <li>Custom agent loop with Ollama tool calling (Qwen 2.5 14B)</li>
            <li>Four tools: vector search, file reader, grep, test runner</li>
            <li>Language-aware Python code chunking via LangChain</li>
            <li>Qdrant vector database for code embeddings</li>
            <li>Server-Sent Events for real-time agent trace streaming</li>
          </ul>

          <h3 className="mt-6 text-lg font-medium">What It Demonstrates</h3>
          <ul className="mt-2 list-disc pl-6 text-muted-foreground space-y-1">
            <li>Agentic AI: multi-step tool use with an observe-think-act loop</li>
            <li>Structured output and function calling with local LLMs</li>
            <li>Building on RAG foundations (embeddings, vector search, chunking)</li>
            <li>Real-time streaming of agent reasoning to a web UI</li>
          </ul>
        </section>

        {/* Debug Assistant Diagram */}
        <section className="mt-12">
          <h2 className="text-2xl font-semibold">How It Works</h2>
          <div className="mt-6 rounded-xl border border-foreground/10 bg-card p-6">
            <MermaidDiagram chart={debugDiagram} />
          </div>
        </section>

        {/* Debug Demo Link */}
        <section className="mt-12">
          <Link
            href="/ai/debug"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Try the Debug Demo &rarr;
          </Link>
        </section>
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/ai/page.tsx
git commit -m "feat(frontend): add Debug Assistant section to /ai hub page"
```

---

### Task 13: Model Migration — Update Existing Services

**Files:**
- Modify: `services/chat/app/config.py`
- Modify: `frontend/src/app/ai/page.tsx` (already done in Task 12)

- [ ] **Step 1: Update chat service default model**

In `services/chat/app/config.py`, change:
```python
chat_model: str = "mistral"
```
to:
```python
chat_model: str = "qwen2.5:14b"
```

- [ ] **Step 2: Run chat service tests to ensure nothing breaks**

Run: `cd services/chat && python -m pytest tests/ -v`
Expected: All pass (tests mock Ollama, so model name doesn't affect them)

- [ ] **Step 3: Commit**

```bash
git add services/chat/app/config.py
git commit -m "feat(chat): upgrade default model from Mistral 7B to Qwen 2.5 14B"
```

---

### Task 14: Full Test Suite Verification

**Files:** None (verification only)

- [ ] **Step 1: Run all debug service tests**

Run: `cd services/debug && python -m pytest tests/ -v --tb=short`
Expected: All pass

- [ ] **Step 2: Run all chat service tests**

Run: `cd services/chat && python -m pytest tests/ -v --tb=short`
Expected: All pass

- [ ] **Step 3: Run all ingestion service tests**

Run: `cd services/ingestion && python -m pytest tests/ -v --tb=short`
Expected: All pass

- [ ] **Step 4: Run ruff on debug service**

Run: `cd services/debug && ruff check app/ tests/ && ruff format --check app/ tests/`
Expected: No errors

- [ ] **Step 5: Run frontend TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 6: Run frontend build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

---

### Task 15: Ollama Setup on Windows PC

**Files:** None (remote machine setup)

- [ ] **Step 1: Pull Qwen 2.5 14B model**

SSH to Windows PC and pull the model:
```bash
ssh PC@100.79.113.84 "ollama pull qwen2.5:14b"
```

- [ ] **Step 2: Set OLLAMA_KEEP_ALIVE**

On the Windows PC, set the environment variable so models stay loaded:
```
OLLAMA_KEEP_ALIVE=-1
```

This needs to be set as a system environment variable on Windows and the Ollama service restarted.

- [ ] **Step 3: Verify model loads**

```bash
ssh PC@100.79.113.84 "curl -s http://localhost:11434/api/generate -d '{\"model\": \"qwen2.5:14b\", \"prompt\": \"hello\", \"stream\": false}' | head -c 200"
```

Expected: JSON response with generated text

- [ ] **Step 4: Set up SSH tunnel for debug service**

Update the SSH tunnel command to include port 8003:
```bash
ssh -f -N -L 8001:localhost:8001 -L 8002:localhost:8002 -L 8003:localhost:8003 PC@100.79.113.84
```

---

### Task 16: Cloudflare Tunnel Configuration

- [ ] **Step 1: Add debug service route to Cloudflare Tunnel**

On the Windows PC, update the Cloudflare Tunnel config to add:
```
api-debug.kylebradshaw.dev → http://localhost:8003
```

- [ ] **Step 2: Update frontend environment for production**

Add `NEXT_PUBLIC_DEBUG_API_URL=https://api-debug.kylebradshaw.dev` to the Vercel environment variables.

- [ ] **Step 3: Commit any config changes**

```bash
git add -A
git commit -m "feat: configure debug service for production deployment"
```
