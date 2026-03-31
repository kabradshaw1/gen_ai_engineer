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
