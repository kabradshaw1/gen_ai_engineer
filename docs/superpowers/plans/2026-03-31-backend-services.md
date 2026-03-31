# Backend Services Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build two FastAPI microservices (Ingestion + Chat) orchestrated via Docker Compose with Qdrant and Ollama, implementing a complete RAG pipeline for PDF document Q&A.

**Architecture:** Two independent FastAPI services share a Qdrant vector database. The Ingestion API parses PDFs, chunks text, embeds via Ollama, and stores in Qdrant. The Chat API embeds user questions, retrieves relevant chunks from Qdrant, and streams LLM responses via SSE. Ollama runs on the host (GPU access), services connect to it over the Docker network.

**Tech Stack:** Python 3.11, FastAPI, LangChain, PyPDF2, Qdrant (Docker), Ollama (host), Docker Compose, pytest

---

## File Structure

```
services/
├── ingestion/
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py          # FastAPI app, /ingest + /documents endpoints
│   │   ├── config.py         # Settings via pydantic-settings
│   │   ├── pdf_parser.py     # PDF text extraction (PyPDF2)
│   │   ├── chunker.py        # Text splitting (LangChain)
│   │   ├── embedder.py       # Ollama embedding calls
│   │   └── store.py          # Qdrant upsert operations
│   └── tests/
│       ├── __init__.py
│       ├── conftest.py
│       ├── test_pdf_parser.py
│       ├── test_chunker.py
│       ├── test_embedder.py
│       ├── test_store.py
│       └── test_main.py
├── chat/
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py           # FastAPI app, /chat + /health endpoints
│   │   ├── config.py          # Settings via pydantic-settings
│   │   ├── retriever.py       # Qdrant search logic
│   │   ├── prompt.py          # Prompt templates
│   │   └── chain.py           # LangChain RAG chain + streaming
│   └── tests/
│       ├── __init__.py
│       ├── conftest.py
│       ├── test_retriever.py
│       ├── test_prompt.py
│       ├── test_chain.py
│       └── test_main.py
docker-compose.yml
.env.example
```

---

### Task 1: Project Scaffolding + Docker Compose

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`
- Create: `services/ingestion/Dockerfile`
- Create: `services/ingestion/requirements.txt`
- Create: `services/ingestion/app/__init__.py`
- Create: `services/ingestion/app/config.py`
- Create: `services/ingestion/app/main.py`
- Create: `services/chat/Dockerfile`
- Create: `services/chat/requirements.txt`
- Create: `services/chat/app/__init__.py`
- Create: `services/chat/app/config.py`
- Create: `services/chat/app/main.py`

- [ ] **Step 1: Create .env.example**

```env
# Ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
CHAT_MODEL=mistral
EMBEDDING_MODEL=nomic-embed-text

# Qdrant
QDRANT_HOST=qdrant
QDRANT_PORT=6333
COLLECTION_NAME=documents

# Ingestion
CHUNK_SIZE=1000
CHUNK_OVERLAP=200
MAX_FILE_SIZE_MB=50
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
services:
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6333/readyz"]
      interval: 10s
      timeout: 5s
      retries: 5

  ingestion:
    build:
      context: ./services/ingestion
    ports:
      - "8001:8000"
    env_file: .env
    depends_on:
      qdrant:
        condition: service_healthy
    extra_hosts:
      - "host.docker.internal:host-gateway"

  chat:
    build:
      context: ./services/chat
    ports:
      - "8002:8000"
    env_file: .env
    depends_on:
      qdrant:
        condition: service_healthy
    extra_hosts:
      - "host.docker.internal:host-gateway"

volumes:
  qdrant_data:
```

- [ ] **Step 3: Create ingestion service scaffold**

`services/ingestion/requirements.txt`:
```
fastapi==0.115.0
uvicorn[standard]==0.30.0
python-multipart==0.0.9
pypdf2==3.0.1
langchain-text-splitters==0.2.0
langchain-community==0.2.0
qdrant-client==1.9.0
httpx==0.27.0
pydantic-settings==2.3.0
pytest==8.2.0
pytest-asyncio==0.23.0
```

`services/ingestion/Dockerfile`:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

`services/ingestion/app/__init__.py`: empty file

`services/ingestion/app/config.py`:
```python
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ollama_base_url: str = "http://host.docker.internal:11434"
    embedding_model: str = "nomic-embed-text"
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    collection_name: str = "documents"
    chunk_size: int = 1000
    chunk_overlap: int = 200
    max_file_size_mb: int = 50


settings = Settings()
```

`services/ingestion/app/main.py`:
```python
from fastapi import FastAPI

app = FastAPI(title="Ingestion API")


@app.get("/health")
async def health():
    return {"status": "ok"}
```

- [ ] **Step 4: Create chat service scaffold**

`services/chat/requirements.txt`:
```
fastapi==0.115.0
uvicorn[standard]==0.30.0
langchain-community==0.2.0
qdrant-client==1.9.0
httpx==0.27.0
pydantic-settings==2.3.0
sse-starlette==2.1.0
pytest==8.2.0
pytest-asyncio==0.23.0
```

`services/chat/Dockerfile`:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

`services/chat/app/__init__.py`: empty file

`services/chat/app/config.py`:
```python
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ollama_base_url: str = "http://host.docker.internal:11434"
    chat_model: str = "mistral"
    embedding_model: str = "nomic-embed-text"
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    collection_name: str = "documents"


settings = Settings()
```

`services/chat/app/main.py`:
```python
from fastapi import FastAPI

app = FastAPI(title="Chat API")


@app.get("/health")
async def health():
    return {"status": "ok"}
```

- [ ] **Step 5: Verify scaffold builds**

Run: `docker compose build`
Expected: Both images build successfully, no errors.

- [ ] **Step 6: Verify services start**

Run: `docker compose up -d && sleep 5 && curl http://localhost:8001/health && curl http://localhost:8002/health && docker compose down`
Expected: Both return `{"status":"ok"}`

- [ ] **Step 7: Commit**

```bash
git add docker-compose.yml .env.example services/
git commit -m "scaffold: FastAPI services + Docker Compose with Qdrant"
```

---

### Task 2: PDF Parser Module

**Files:**
- Create: `services/ingestion/tests/__init__.py`
- Create: `services/ingestion/tests/conftest.py`
- Create: `services/ingestion/tests/test_pdf_parser.py`
- Create: `services/ingestion/app/pdf_parser.py`

- [ ] **Step 1: Create test fixtures**

`services/ingestion/tests/__init__.py`: empty file

`services/ingestion/tests/conftest.py`:
```python
import io
import pytest
from pypdf2 import PdfWriter


@pytest.fixture
def sample_pdf_bytes() -> bytes:
    """Create a simple 2-page PDF in memory."""
    writer = PdfWriter()

    # Page 1
    writer.add_blank_page(width=612, height=792)
    page1 = writer.pages[0]
    # We'll use a real PDF with text for integration tests.
    # For unit tests, we test the interface contract.

    # Page 2
    writer.add_blank_page(width=612, height=792)

    buffer = io.BytesIO()
    writer.write(buffer)
    buffer.seek(0)
    return buffer.read()


@pytest.fixture
def empty_bytes() -> bytes:
    return b""
```

- [ ] **Step 2: Write failing tests for pdf_parser**

`services/ingestion/tests/test_pdf_parser.py`:
```python
import io
import pytest
from app.pdf_parser import extract_pages


def test_extract_pages_returns_list_of_dicts(sample_pdf_bytes):
    pages = extract_pages(io.BytesIO(sample_pdf_bytes))
    assert isinstance(pages, list)
    assert len(pages) == 2
    for page in pages:
        assert "page_number" in page
        assert "text" in page


def test_extract_pages_page_numbers_are_1_indexed(sample_pdf_bytes):
    pages = extract_pages(io.BytesIO(sample_pdf_bytes))
    assert pages[0]["page_number"] == 1
    assert pages[1]["page_number"] == 2


def test_extract_pages_empty_bytes_raises():
    with pytest.raises(ValueError, match="empty or invalid"):
        extract_pages(io.BytesIO(b""))


def test_extract_pages_invalid_pdf_raises():
    with pytest.raises(ValueError, match="empty or invalid"):
        extract_pages(io.BytesIO(b"not a pdf"))
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd services/ingestion && python -m pytest tests/test_pdf_parser.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.pdf_parser'`

- [ ] **Step 4: Implement pdf_parser.py**

`services/ingestion/app/pdf_parser.py`:
```python
from io import BytesIO
from PyPDF2 import PdfReader


def extract_pages(pdf_file: BytesIO) -> list[dict]:
    """Extract text from each page of a PDF.

    Returns a list of dicts with 'page_number' (1-indexed) and 'text' keys.
    Raises ValueError if the file is empty or not a valid PDF.
    """
    try:
        content = pdf_file.read()
        if not content:
            raise ValueError("empty or invalid PDF")
        pdf_file.seek(0)
        reader = PdfReader(pdf_file)
    except Exception as e:
        if "empty or invalid" in str(e):
            raise
        raise ValueError(f"empty or invalid PDF: {e}")

    pages = []
    for i, page in enumerate(reader.pages):
        text = page.extract_text() or ""
        pages.append({"page_number": i + 1, "text": text})

    return pages
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd services/ingestion && python -m pytest tests/test_pdf_parser.py -v`
Expected: All 4 tests PASS

- [ ] **Step 6: Commit**

```bash
git add services/ingestion/app/pdf_parser.py services/ingestion/tests/
git commit -m "feat: add PDF text extraction module"
```

---

### Task 3: Chunker Module

**Files:**
- Create: `services/ingestion/tests/test_chunker.py`
- Create: `services/ingestion/app/chunker.py`

- [ ] **Step 1: Write failing tests for chunker**

`services/ingestion/tests/test_chunker.py`:
```python
import pytest
from app.chunker import chunk_pages


def test_chunk_pages_returns_list_of_dicts():
    pages = [
        {"page_number": 1, "text": "Hello world. " * 100},
        {"page_number": 2, "text": "Goodbye world. " * 100},
    ]
    chunks = chunk_pages(pages, chunk_size=200, chunk_overlap=50)
    assert isinstance(chunks, list)
    assert len(chunks) > 2  # Should produce multiple chunks
    for chunk in chunks:
        assert "text" in chunk
        assert "page_number" in chunk
        assert "chunk_index" in chunk


def test_chunk_pages_preserves_page_numbers():
    pages = [
        {"page_number": 1, "text": "Short text on page one."},
        {"page_number": 3, "text": "Short text on page three."},
    ]
    chunks = chunk_pages(pages, chunk_size=1000, chunk_overlap=0)
    page_numbers = [c["page_number"] for c in chunks]
    assert 1 in page_numbers
    assert 3 in page_numbers


def test_chunk_pages_respects_chunk_size():
    pages = [{"page_number": 1, "text": "word " * 500}]
    chunks = chunk_pages(pages, chunk_size=100, chunk_overlap=20)
    for chunk in chunks:
        # Allow some overflow due to word boundaries
        assert len(chunk["text"]) <= 150


def test_chunk_pages_empty_pages_skipped():
    pages = [
        {"page_number": 1, "text": ""},
        {"page_number": 2, "text": "Has content."},
    ]
    chunks = chunk_pages(pages, chunk_size=1000, chunk_overlap=0)
    assert all(c["page_number"] == 2 for c in chunks)


def test_chunk_pages_sequential_chunk_index():
    pages = [{"page_number": 1, "text": "Hello world. " * 100}]
    chunks = chunk_pages(pages, chunk_size=100, chunk_overlap=20)
    indices = [c["chunk_index"] for c in chunks]
    assert indices == list(range(len(indices)))
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/ingestion && python -m pytest tests/test_chunker.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.chunker'`

- [ ] **Step 3: Implement chunker.py**

`services/ingestion/app/chunker.py`:
```python
from langchain_text_splitters import RecursiveCharacterTextSplitter


def chunk_pages(
    pages: list[dict],
    chunk_size: int = 1000,
    chunk_overlap: int = 200,
) -> list[dict]:
    """Split page text into overlapping chunks.

    Returns list of dicts with 'text', 'page_number', and 'chunk_index' keys.
    Empty pages are skipped.
    """
    splitter = RecursiveCharacterTextSplitter(
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap,
        length_function=len,
    )

    chunks = []
    index = 0
    for page in pages:
        text = page["text"].strip()
        if not text:
            continue

        splits = splitter.split_text(text)
        for split in splits:
            chunks.append({
                "text": split,
                "page_number": page["page_number"],
                "chunk_index": index,
            })
            index += 1

    return chunks
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/ingestion && python -m pytest tests/test_chunker.py -v`
Expected: All 5 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/app/chunker.py services/ingestion/tests/test_chunker.py
git commit -m "feat: add text chunking module with overlap"
```

---

### Task 4: Embedder Module

**Files:**
- Create: `services/ingestion/tests/test_embedder.py`
- Create: `services/ingestion/app/embedder.py`

- [ ] **Step 1: Write failing tests for embedder**

`services/ingestion/tests/test_embedder.py`:
```python
import pytest
from unittest.mock import AsyncMock, patch
from app.embedder import embed_texts


@pytest.mark.asyncio
async def test_embed_texts_returns_list_of_vectors():
    mock_response = {
        "embeddings": [[0.1] * 768, [0.2] * 768]
    }
    with patch("app.embedder.httpx.AsyncClient") as MockClient:
        client_instance = AsyncMock()
        client_instance.post.return_value = AsyncMock(
            status_code=200,
            json=lambda: mock_response,
            raise_for_status=lambda: None,
        )
        MockClient.return_value.__aenter__ = AsyncMock(return_value=client_instance)
        MockClient.return_value.__aexit__ = AsyncMock(return_value=False)

        vectors = await embed_texts(
            texts=["hello", "world"],
            ollama_base_url="http://localhost:11434",
            model="nomic-embed-text",
        )

    assert len(vectors) == 2
    assert len(vectors[0]) == 768


@pytest.mark.asyncio
async def test_embed_texts_calls_ollama_api():
    mock_response = {"embeddings": [[0.1] * 768]}
    with patch("app.embedder.httpx.AsyncClient") as MockClient:
        client_instance = AsyncMock()
        client_instance.post.return_value = AsyncMock(
            status_code=200,
            json=lambda: mock_response,
            raise_for_status=lambda: None,
        )
        MockClient.return_value.__aenter__ = AsyncMock(return_value=client_instance)
        MockClient.return_value.__aexit__ = AsyncMock(return_value=False)

        await embed_texts(
            texts=["hello"],
            ollama_base_url="http://localhost:11434",
            model="nomic-embed-text",
        )

    client_instance.post.assert_called_once_with(
        "http://localhost:11434/api/embed",
        json={"model": "nomic-embed-text", "input": ["hello"]},
        timeout=120.0,
    )


@pytest.mark.asyncio
async def test_embed_texts_empty_list_returns_empty():
    vectors = await embed_texts(
        texts=[],
        ollama_base_url="http://localhost:11434",
        model="nomic-embed-text",
    )
    assert vectors == []
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/ingestion && python -m pytest tests/test_embedder.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.embedder'`

- [ ] **Step 3: Implement embedder.py**

`services/ingestion/app/embedder.py`:
```python
import httpx


async def embed_texts(
    texts: list[str],
    ollama_base_url: str,
    model: str,
) -> list[list[float]]:
    """Embed a list of texts using Ollama's /api/embed endpoint.

    Returns a list of embedding vectors (list of floats).
    """
    if not texts:
        return []

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{ollama_base_url}/api/embed",
            json={"model": model, "input": texts},
            timeout=120.0,
        )
        response.raise_for_status()
        data = response.json()

    return data["embeddings"]
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/ingestion && python -m pytest tests/test_embedder.py -v`
Expected: All 3 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/app/embedder.py services/ingestion/tests/test_embedder.py
git commit -m "feat: add Ollama embedding module"
```

---

### Task 5: Qdrant Store Module

**Files:**
- Create: `services/ingestion/tests/test_store.py`
- Create: `services/ingestion/app/store.py`

- [ ] **Step 1: Write failing tests for store**

`services/ingestion/tests/test_store.py`:
```python
import pytest
from unittest.mock import MagicMock, patch, AsyncMock
from app.store import QdrantStore


@pytest.fixture
def mock_qdrant_client():
    with patch("app.store.QdrantClient") as MockClient:
        client = MagicMock()
        MockClient.return_value = client
        yield client


def test_store_init_creates_collection_if_not_exists(mock_qdrant_client):
    mock_qdrant_client.collection_exists.return_value = False
    store = QdrantStore(host="localhost", port=6333, collection_name="test")
    mock_qdrant_client.create_collection.assert_called_once()


def test_store_init_skips_creation_if_exists(mock_qdrant_client):
    mock_qdrant_client.collection_exists.return_value = True
    store = QdrantStore(host="localhost", port=6333, collection_name="test")
    mock_qdrant_client.create_collection.assert_not_called()


def test_upsert_vectors(mock_qdrant_client):
    mock_qdrant_client.collection_exists.return_value = True
    store = QdrantStore(host="localhost", port=6333, collection_name="test")

    chunks = [
        {"text": "hello", "page_number": 1, "chunk_index": 0},
        {"text": "world", "page_number": 1, "chunk_index": 1},
    ]
    vectors = [[0.1] * 768, [0.2] * 768]

    store.upsert(
        chunks=chunks,
        vectors=vectors,
        document_id="doc-123",
        filename="test.pdf",
    )

    mock_qdrant_client.upsert.assert_called_once()
    call_args = mock_qdrant_client.upsert.call_args
    assert call_args.kwargs["collection_name"] == "test"
    points = call_args.kwargs["points"]
    assert len(points) == 2
    assert points[0].payload["filename"] == "test.pdf"
    assert points[0].payload["document_id"] == "doc-123"
    assert points[0].payload["page_number"] == 1
    assert points[0].payload["text"] == "hello"


def test_list_documents(mock_qdrant_client):
    mock_qdrant_client.collection_exists.return_value = True
    store = QdrantStore(host="localhost", port=6333, collection_name="test")

    mock_qdrant_client.scroll.return_value = (
        [
            MagicMock(payload={
                "document_id": "doc-1",
                "filename": "a.pdf",
                "page_number": 1,
                "chunk_index": 0,
            }),
            MagicMock(payload={
                "document_id": "doc-1",
                "filename": "a.pdf",
                "page_number": 1,
                "chunk_index": 1,
            }),
            MagicMock(payload={
                "document_id": "doc-2",
                "filename": "b.pdf",
                "page_number": 1,
                "chunk_index": 0,
            }),
        ],
        None,
    )

    docs = store.list_documents()
    assert len(docs) == 2
    assert docs[0]["document_id"] == "doc-1"
    assert docs[0]["chunks"] == 2
    assert docs[1]["document_id"] == "doc-2"
    assert docs[1]["chunks"] == 1
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/ingestion && python -m pytest tests/test_store.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.store'`

- [ ] **Step 3: Implement store.py**

`services/ingestion/app/store.py`:
```python
import uuid
from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance,
    PointStruct,
    VectorParams,
)


class QdrantStore:
    def __init__(self, host: str, port: int, collection_name: str):
        self.client = QdrantClient(host=host, port=port)
        self.collection_name = collection_name
        self._ensure_collection()

    def _ensure_collection(self):
        if not self.client.collection_exists(self.collection_name):
            self.client.create_collection(
                collection_name=self.collection_name,
                vectors_config=VectorParams(
                    size=768,
                    distance=Distance.COSINE,
                ),
            )

    def upsert(
        self,
        chunks: list[dict],
        vectors: list[list[float]],
        document_id: str,
        filename: str,
    ) -> None:
        points = [
            PointStruct(
                id=str(uuid.uuid4()),
                vector=vector,
                payload={
                    "text": chunk["text"],
                    "page_number": chunk["page_number"],
                    "chunk_index": chunk["chunk_index"],
                    "document_id": document_id,
                    "filename": filename,
                },
            )
            for chunk, vector in zip(chunks, vectors)
        ]
        self.client.upsert(
            collection_name=self.collection_name,
            points=points,
        )

    def list_documents(self) -> list[dict]:
        records, _ = self.client.scroll(
            collection_name=self.collection_name,
            limit=10000,
            with_payload=True,
            with_vectors=False,
        )

        docs: dict[str, dict] = {}
        for record in records:
            doc_id = record.payload["document_id"]
            if doc_id not in docs:
                docs[doc_id] = {
                    "document_id": doc_id,
                    "filename": record.payload["filename"],
                    "chunks": 0,
                }
            docs[doc_id]["chunks"] += 1

        return list(docs.values())
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/ingestion && python -m pytest tests/test_store.py -v`
Expected: All 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/app/store.py services/ingestion/tests/test_store.py
git commit -m "feat: add Qdrant vector store module"
```

---

### Task 6: Ingestion API Endpoint

**Files:**
- Create: `services/ingestion/tests/test_main.py`
- Modify: `services/ingestion/app/main.py`

- [ ] **Step 1: Write failing tests for /ingest and /documents endpoints**

`services/ingestion/tests/test_main.py`:
```python
import io
import pytest
from unittest.mock import patch, AsyncMock, MagicMock
from fastapi.testclient import TestClient
from app.main import app


client = TestClient(app)


def test_health():
    response = client.get("/health")
    assert response.status_code == 200


@patch("app.main.get_store")
@patch("app.main.embed_texts", new_callable=AsyncMock)
@patch("app.main.extract_pages")
def test_ingest_pdf_success(mock_extract, mock_embed, mock_get_store):
    mock_extract.return_value = [
        {"page_number": 1, "text": "Hello world. " * 100},
    ]
    mock_embed.return_value = [[0.1] * 768] * 2
    mock_store = MagicMock()
    mock_get_store.return_value = mock_store

    pdf_content = b"%PDF-1.4 fake content"
    response = client.post(
        "/ingest",
        files={"file": ("test.pdf", io.BytesIO(pdf_content), "application/pdf")},
    )

    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["filename"] == "test.pdf"
    assert "document_id" in data
    assert "chunks_created" in data


@patch("app.main.get_store")
def test_ingest_rejects_non_pdf(mock_get_store):
    response = client.post(
        "/ingest",
        files={"file": ("test.txt", io.BytesIO(b"hello"), "text/plain")},
    )
    assert response.status_code == 422


@patch("app.main.get_store")
def test_documents_list(mock_get_store):
    mock_store = MagicMock()
    mock_store.list_documents.return_value = [
        {"document_id": "abc", "filename": "test.pdf", "chunks": 5},
    ]
    mock_get_store.return_value = mock_store

    response = client.get("/documents")
    assert response.status_code == 200
    data = response.json()
    assert len(data["documents"]) == 1
    assert data["documents"][0]["filename"] == "test.pdf"
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/ingestion && python -m pytest tests/test_main.py -v`
Expected: FAIL — missing imports and endpoints

- [ ] **Step 3: Implement full ingestion main.py**

`services/ingestion/app/main.py`:
```python
import uuid
from io import BytesIO

from fastapi import FastAPI, File, HTTPException, UploadFile

from app.chunker import chunk_pages
from app.config import settings
from app.embedder import embed_texts
from app.pdf_parser import extract_pages
from app.store import QdrantStore

app = FastAPI(title="Ingestion API")

_store: QdrantStore | None = None


def get_store() -> QdrantStore:
    global _store
    if _store is None:
        _store = QdrantStore(
            host=settings.qdrant_host,
            port=settings.qdrant_port,
            collection_name=settings.collection_name,
        )
    return _store


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.post("/ingest")
async def ingest(file: UploadFile = File(...)):
    if not file.filename or not file.filename.lower().endswith(".pdf"):
        raise HTTPException(status_code=422, detail="Only PDF files are accepted")

    content = await file.read()
    max_bytes = settings.max_file_size_mb * 1024 * 1024
    if len(content) > max_bytes:
        raise HTTPException(
            status_code=422,
            detail=f"File exceeds {settings.max_file_size_mb}MB limit",
        )

    try:
        pages = extract_pages(BytesIO(content))
    except ValueError as e:
        raise HTTPException(status_code=422, detail=str(e))

    chunks = chunk_pages(
        pages,
        chunk_size=settings.chunk_size,
        chunk_overlap=settings.chunk_overlap,
    )

    if not chunks:
        raise HTTPException(status_code=422, detail="No text content found in PDF")

    texts = [c["text"] for c in chunks]
    vectors = await embed_texts(
        texts=texts,
        ollama_base_url=settings.ollama_base_url,
        model=settings.embedding_model,
    )

    document_id = str(uuid.uuid4())
    store = get_store()
    store.upsert(
        chunks=chunks,
        vectors=vectors,
        document_id=document_id,
        filename=file.filename,
    )

    return {
        "status": "success",
        "document_id": document_id,
        "chunks_created": len(chunks),
        "filename": file.filename,
    }


@app.get("/documents")
async def list_documents():
    store = get_store()
    return {"documents": store.list_documents()}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/ingestion && python -m pytest tests/test_main.py -v`
Expected: All 4 tests PASS

- [ ] **Step 5: Run all ingestion tests**

Run: `cd services/ingestion && python -m pytest tests/ -v`
Expected: All 16 tests PASS

- [ ] **Step 6: Commit**

```bash
git add services/ingestion/app/main.py services/ingestion/tests/test_main.py
git commit -m "feat: add /ingest and /documents API endpoints"
```

---

### Task 7: Chat Retriever Module

**Files:**
- Create: `services/chat/tests/__init__.py`
- Create: `services/chat/tests/conftest.py`
- Create: `services/chat/tests/test_retriever.py`
- Create: `services/chat/app/retriever.py`

- [ ] **Step 1: Create test scaffolding**

`services/chat/tests/__init__.py`: empty file

`services/chat/tests/conftest.py`:
```python
# Shared fixtures for chat service tests
```

- [ ] **Step 2: Write failing tests for retriever**

`services/chat/tests/test_retriever.py`:
```python
import pytest
from unittest.mock import MagicMock, patch
from app.retriever import QdrantRetriever


@pytest.fixture
def mock_qdrant_client():
    with patch("app.retriever.QdrantClient") as MockClient:
        client = MagicMock()
        MockClient.return_value = client
        yield client


def test_search_returns_chunks_with_scores(mock_qdrant_client):
    mock_qdrant_client.search.return_value = [
        MagicMock(
            score=0.95,
            payload={
                "text": "relevant chunk",
                "page_number": 1,
                "filename": "doc.pdf",
                "document_id": "abc",
            },
        ),
        MagicMock(
            score=0.85,
            payload={
                "text": "another chunk",
                "page_number": 2,
                "filename": "doc.pdf",
                "document_id": "abc",
            },
        ),
    ]

    retriever = QdrantRetriever(
        host="localhost", port=6333, collection_name="test"
    )
    results = retriever.search(query_vector=[0.1] * 768, top_k=5)

    assert len(results) == 2
    assert results[0]["text"] == "relevant chunk"
    assert results[0]["score"] == 0.95
    assert results[0]["page_number"] == 1
    assert results[0]["filename"] == "doc.pdf"


def test_search_respects_top_k(mock_qdrant_client):
    mock_qdrant_client.search.return_value = []
    retriever = QdrantRetriever(
        host="localhost", port=6333, collection_name="test"
    )
    retriever.search(query_vector=[0.1] * 768, top_k=3)

    call_args = mock_qdrant_client.search.call_args
    assert call_args.kwargs["limit"] == 3


def test_search_empty_results(mock_qdrant_client):
    mock_qdrant_client.search.return_value = []
    retriever = QdrantRetriever(
        host="localhost", port=6333, collection_name="test"
    )
    results = retriever.search(query_vector=[0.1] * 768, top_k=5)
    assert results == []
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd services/chat && python -m pytest tests/test_retriever.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.retriever'`

- [ ] **Step 4: Implement retriever.py**

`services/chat/app/retriever.py`:
```python
from qdrant_client import QdrantClient


class QdrantRetriever:
    def __init__(self, host: str, port: int, collection_name: str):
        self.client = QdrantClient(host=host, port=port)
        self.collection_name = collection_name

    def search(
        self, query_vector: list[float], top_k: int = 5
    ) -> list[dict]:
        results = self.client.search(
            collection_name=self.collection_name,
            query_vector=query_vector,
            limit=top_k,
        )

        return [
            {
                "text": hit.payload["text"],
                "page_number": hit.payload["page_number"],
                "filename": hit.payload["filename"],
                "document_id": hit.payload["document_id"],
                "score": hit.score,
            }
            for hit in results
        ]
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd services/chat && python -m pytest tests/test_retriever.py -v`
Expected: All 3 tests PASS

- [ ] **Step 6: Commit**

```bash
git add services/chat/app/retriever.py services/chat/tests/
git commit -m "feat: add Qdrant retriever module for chat service"
```

---

### Task 8: Prompt Templates Module

**Files:**
- Create: `services/chat/tests/test_prompt.py`
- Create: `services/chat/app/prompt.py`

- [ ] **Step 1: Write failing tests for prompt**

`services/chat/tests/test_prompt.py`:
```python
from app.prompt import build_rag_prompt


def test_build_rag_prompt_includes_context():
    chunks = [
        {"text": "The revenue was $1M.", "filename": "report.pdf", "page_number": 3},
    ]
    prompt = build_rag_prompt(question="What was the revenue?", chunks=chunks)
    assert "The revenue was $1M." in prompt
    assert "What was the revenue?" in prompt


def test_build_rag_prompt_includes_source_attribution():
    chunks = [
        {"text": "Some fact.", "filename": "doc.pdf", "page_number": 5},
    ]
    prompt = build_rag_prompt(question="Tell me a fact.", chunks=chunks)
    assert "doc.pdf" in prompt
    assert "5" in prompt


def test_build_rag_prompt_multiple_chunks():
    chunks = [
        {"text": "First chunk.", "filename": "a.pdf", "page_number": 1},
        {"text": "Second chunk.", "filename": "b.pdf", "page_number": 2},
    ]
    prompt = build_rag_prompt(question="Summarize.", chunks=chunks)
    assert "First chunk." in prompt
    assert "Second chunk." in prompt


def test_build_rag_prompt_empty_chunks():
    prompt = build_rag_prompt(question="Anything?", chunks=[])
    assert "Anything?" in prompt
    assert "no relevant context" in prompt.lower() or "don't have" in prompt.lower()
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/chat && python -m pytest tests/test_prompt.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.prompt'`

- [ ] **Step 3: Implement prompt.py**

`services/chat/app/prompt.py`:
```python
SYSTEM_PROMPT = """You are a helpful document Q&A assistant. Answer questions based only on the provided context. If the context doesn't contain enough information to answer, say so honestly — do not make up information.

When referencing information, mention the source file and page number."""

RAG_TEMPLATE = """Context:
{context}

Question: {question}

Answer based only on the context above. Cite sources (filename, page) when possible."""

NO_CONTEXT_TEMPLATE = """The user asked: {question}

I don't have any relevant context from uploaded documents to answer this question. Please upload a relevant document first, or rephrase your question."""


def build_rag_prompt(question: str, chunks: list[dict]) -> str:
    if not chunks:
        return NO_CONTEXT_TEMPLATE.format(question=question)

    context_parts = []
    for chunk in chunks:
        source = f"[{chunk['filename']}, page {chunk['page_number']}]"
        context_parts.append(f"{source}\n{chunk['text']}")

    context = "\n\n".join(context_parts)
    return RAG_TEMPLATE.format(context=context, question=question)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/chat && python -m pytest tests/test_prompt.py -v`
Expected: All 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/chat/app/prompt.py services/chat/tests/test_prompt.py
git commit -m "feat: add RAG prompt templates"
```

---

### Task 9: RAG Chain Module

**Files:**
- Create: `services/chat/tests/test_chain.py`
- Create: `services/chat/app/chain.py`

- [ ] **Step 1: Write failing tests for chain**

`services/chat/tests/test_chain.py`:
```python
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
from app.chain import rag_query


@pytest.mark.asyncio
@patch("app.chain.stream_ollama_response")
@patch("app.chain.embed_texts", new_callable=AsyncMock)
@patch("app.chain.QdrantRetriever")
async def test_rag_query_returns_generator(
    MockRetriever, mock_embed, mock_stream
):
    mock_embed.return_value = [[0.1] * 768]
    retriever_instance = MagicMock()
    retriever_instance.search.return_value = [
        {
            "text": "The answer is 42.",
            "filename": "doc.pdf",
            "page_number": 1,
            "document_id": "abc",
            "score": 0.9,
        },
    ]
    MockRetriever.return_value = retriever_instance

    async def fake_stream(prompt, model, base_url):
        yield {"token": "The"}
        yield {"token": " answer"}
        yield {"token": " is 42."}

    mock_stream.side_effect = fake_stream

    tokens = []
    sources = None
    async for event in rag_query(
        question="What is the answer?",
        ollama_base_url="http://localhost:11434",
        chat_model="mistral",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        collection_name="documents",
    ):
        if "token" in event:
            tokens.append(event["token"])
        if "sources" in event:
            sources = event["sources"]

    assert len(tokens) == 3
    assert sources is not None
    assert sources[0]["file"] == "doc.pdf"
    assert sources[0]["page"] == 1


@pytest.mark.asyncio
@patch("app.chain.embed_texts", new_callable=AsyncMock)
@patch("app.chain.QdrantRetriever")
async def test_rag_query_no_results_still_responds(MockRetriever, mock_embed):
    mock_embed.return_value = [[0.1] * 768]
    retriever_instance = MagicMock()
    retriever_instance.search.return_value = []
    MockRetriever.return_value = retriever_instance

    events = []
    async for event in rag_query(
        question="Unknown topic?",
        ollama_base_url="http://localhost:11434",
        chat_model="mistral",
        embedding_model="nomic-embed-text",
        qdrant_host="localhost",
        qdrant_port=6333,
        collection_name="documents",
    ):
        events.append(event)

    # Should still produce token events (the "no context" response)
    assert any("token" in e for e in events)
    # Should have done/sources event
    assert any("done" in e for e in events)
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/chat && python -m pytest tests/test_chain.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.chain'`

- [ ] **Step 3: Implement chain.py**

`services/chat/app/chain.py`:
```python
from typing import AsyncGenerator

import httpx

from app.prompt import SYSTEM_PROMPT, build_rag_prompt
from app.retriever import QdrantRetriever


async def embed_texts(
    texts: list[str],
    ollama_base_url: str,
    model: str,
) -> list[list[float]]:
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


async def stream_ollama_response(
    prompt: str,
    model: str,
    base_url: str,
) -> AsyncGenerator[dict, None]:
    async with httpx.AsyncClient() as client:
        async with client.stream(
            "POST",
            f"{base_url}/api/generate",
            json={
                "model": model,
                "prompt": prompt,
                "system": SYSTEM_PROMPT,
                "stream": True,
            },
            timeout=300.0,
        ) as response:
            response.raise_for_status()
            import json

            async for line in response.aiter_lines():
                if line.strip():
                    data = json.loads(line)
                    if data.get("response"):
                        yield {"token": data["response"]}
                    if data.get("done"):
                        break


async def rag_query(
    question: str,
    ollama_base_url: str,
    chat_model: str,
    embedding_model: str,
    qdrant_host: str,
    qdrant_port: int,
    collection_name: str,
    top_k: int = 5,
) -> AsyncGenerator[dict, None]:
    # Embed the question
    vectors = await embed_texts(
        texts=[question],
        ollama_base_url=ollama_base_url,
        model=embedding_model,
    )
    query_vector = vectors[0]

    # Retrieve relevant chunks
    retriever = QdrantRetriever(
        host=qdrant_host, port=qdrant_port, collection_name=collection_name
    )
    chunks = retriever.search(query_vector=query_vector, top_k=top_k)

    # Build prompt
    prompt = build_rag_prompt(question=question, chunks=chunks)

    # Collect unique sources
    seen = set()
    sources = []
    for chunk in chunks:
        key = (chunk["filename"], chunk["page_number"])
        if key not in seen:
            seen.add(key)
            sources.append({"file": chunk["filename"], "page": chunk["page_number"]})

    # Stream response
    async for event in stream_ollama_response(
        prompt=prompt, model=chat_model, base_url=ollama_base_url
    ):
        yield event

    yield {"done": True, "sources": sources}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/chat && python -m pytest tests/test_chain.py -v`
Expected: All 2 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/chat/app/chain.py services/chat/tests/test_chain.py
git commit -m "feat: add RAG chain with streaming Ollama responses"
```

---

### Task 10: Chat API Endpoints

**Files:**
- Create: `services/chat/tests/test_main.py`
- Modify: `services/chat/app/main.py`

- [ ] **Step 1: Write failing tests for /chat and /health**

`services/chat/tests/test_main.py`:
```python
import json
import pytest
from unittest.mock import patch, AsyncMock, MagicMock
from fastapi.testclient import TestClient
from app.main import app


client = TestClient(app)


def test_health():
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy" or data["status"] == "ok"


@patch("app.main.rag_query")
def test_chat_streams_response(mock_rag_query):
    async def fake_rag_query(**kwargs):
        yield {"token": "Hello"}
        yield {"token": " world"}
        yield {"done": True, "sources": [{"file": "test.pdf", "page": 1}]}

    mock_rag_query.return_value = fake_rag_query()

    response = client.post(
        "/chat",
        json={"question": "What is this?", "collection": "default"},
    )
    assert response.status_code == 200
    assert "text/event-stream" in response.headers["content-type"]

    events = []
    for line in response.text.strip().split("\n"):
        if line.startswith("data: "):
            events.append(json.loads(line[6:]))

    tokens = [e["token"] for e in events if "token" in e]
    assert "Hello" in tokens
    done_events = [e for e in events if e.get("done")]
    assert len(done_events) == 1
    assert done_events[0]["sources"][0]["file"] == "test.pdf"


def test_chat_requires_question():
    response = client.post("/chat", json={})
    assert response.status_code == 422
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/chat && python -m pytest tests/test_main.py -v`
Expected: FAIL — missing endpoints

- [ ] **Step 3: Implement full chat main.py**

`services/chat/app/main.py`:
```python
import json

from fastapi import FastAPI
from pydantic import BaseModel
from sse_starlette.sse import EventSourceResponse

from app.chain import rag_query
from app.config import settings

app = FastAPI(title="Chat API")


class ChatRequest(BaseModel):
    question: str
    collection: str = "default"


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.post("/chat")
async def chat(request: ChatRequest):
    async def event_generator():
        async for event in rag_query(
            question=request.question,
            ollama_base_url=settings.ollama_base_url,
            chat_model=settings.chat_model,
            embedding_model=settings.embedding_model,
            qdrant_host=settings.qdrant_host,
            qdrant_port=settings.qdrant_port,
            collection_name=request.collection or settings.collection_name,
        ):
            yield {"data": json.dumps(event)}

    return EventSourceResponse(event_generator())
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/chat && python -m pytest tests/test_main.py -v`
Expected: All 3 tests PASS

- [ ] **Step 5: Run all chat tests**

Run: `cd services/chat && python -m pytest tests/ -v`
Expected: All 12 tests PASS

- [ ] **Step 6: Commit**

```bash
git add services/chat/app/main.py services/chat/tests/test_main.py
git commit -m "feat: add /chat SSE streaming endpoint and /health check"
```

---

### Task 11: Docker Compose Integration Test

**Files:**
- No new files — validates the full stack works end-to-end

- [ ] **Step 1: Ensure Ollama is running with required models**

Run: `ollama list`
Expected: Should show available models. If `mistral` and `nomic-embed-text` are missing:

Run: `ollama pull mistral && ollama pull nomic-embed-text`

- [ ] **Step 2: Create .env from example**

Run: `cp .env.example .env`

If Ollama is running on the host (not Docker), verify `OLLAMA_BASE_URL=http://host.docker.internal:11434` in `.env`.

- [ ] **Step 3: Build and start all services**

Run: `docker compose up --build -d`
Expected: All 3 services start (qdrant, ingestion, chat). Check with:

Run: `docker compose ps`
Expected: All services show "Up" or "healthy"

- [ ] **Step 4: Verify health endpoints**

Run: `curl http://localhost:8001/health`
Expected: `{"status":"ok"}`

Run: `curl http://localhost:8002/health`
Expected: `{"status":"healthy"}`

- [ ] **Step 5: Test ingestion with a real PDF**

Run: `curl -X POST http://localhost:8001/ingest -F "file=@test.pdf"`

(Use any PDF you have available, or create a simple one.)

Expected: `{"status":"success","document_id":"...","chunks_created":N,"filename":"test.pdf"}`

- [ ] **Step 6: Test chat with a question about the PDF**

Run: `curl -N -X POST http://localhost:8002/chat -H "Content-Type: application/json" -d '{"question":"What is this document about?"}'`

Expected: SSE stream of tokens, ending with a `done` event that includes sources.

- [ ] **Step 7: Test document listing**

Run: `curl http://localhost:8001/documents`

Expected: `{"documents":[{"document_id":"...","filename":"test.pdf","chunks":N}]}`

- [ ] **Step 8: Clean up**

Run: `docker compose down`

- [ ] **Step 9: Commit .env to gitignore and final cleanup**

Verify `.env` is in `.gitignore` (it should be from the existing gitignore). Then:

```bash
git add -A
git commit -m "feat: complete backend services with Docker Compose integration"
```
