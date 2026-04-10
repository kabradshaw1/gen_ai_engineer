"""Prometheus metrics for the ingestion service."""

from prometheus_client import Counter, Histogram
from prometheus_fastapi_instrumentator import Instrumentator

SERVICE = "ingestion"

# --- Instrumentator (auto RED metrics on /metrics) ---
instrumentator = Instrumentator(
    should_group_status_codes=False,
    excluded_handlers=["/health", "/metrics"],
)

# --- Ollama embedding metrics ---
EMBEDDING_DURATION = Histogram(
    "embedding_duration_seconds",
    "Time spent calling Ollama /api/embed",
    ["service", "model"],
    buckets=(0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0),
)

# --- Qdrant metrics ---
QDRANT_OPERATION_DURATION = Histogram(
    "qdrant_operation_duration_seconds",
    "Time spent on Qdrant operations",
    ["service", "operation"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 5.0),
)

# --- Pipeline metrics ---
CHUNKS_CREATED = Counter(
    "ingestion_chunks_created_total",
    "Total number of chunks created during ingestion",
    ["service"],
)
