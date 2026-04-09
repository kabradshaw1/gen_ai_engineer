from fastapi.testclient import TestClient
from main import app

client = TestClient(app)


def test_embeddings_returns_fixed_768_vector():
    resp = client.post(
        "/api/embeddings",
        json={"model": "nomic-embed-text", "prompt": "hello"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert "embedding" in body
    assert isinstance(body["embedding"], list)
    assert len(body["embedding"]) == 768
    assert all(isinstance(x, float) for x in body["embedding"])


def test_chat_returns_ndjson_stream():
    resp = client.post(
        "/api/chat",
        json={
            "model": "qwen2.5:14b",
            "messages": [{"role": "user", "content": "hi"}],
            "stream": True,
        },
    )
    assert resp.status_code == 200
    lines = [line for line in resp.text.splitlines() if line.strip()]
    assert len(lines) >= 2
    import json

    first = json.loads(lines[0])
    assert first["message"]["role"] == "assistant"
    assert "content" in first["message"]
    last = json.loads(lines[-1])
    assert last.get("done") is True


def test_tags_endpoint_returns_empty_list():
    resp = client.get("/api/tags")
    assert resp.status_code == 200
    assert resp.json() == {"models": []}
