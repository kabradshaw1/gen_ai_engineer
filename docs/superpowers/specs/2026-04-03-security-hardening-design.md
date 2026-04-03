# Security Hardening — Application Input Layer

**Date:** 2026-04-03
**Status:** Proposed
**Scope:** Input validation, error sanitization, XSS protection, CORS tightening

## Context

The project's infrastructure-level security is solid: CI runs Bandit, pip-audit, npm audit, Gitleaks, and Hadolint. CORS origins are restricted via environment variable. Dockerfiles run as non-root. However, the application input layer has gaps — user-supplied parameters lack validation, error messages leak internal details, and the frontend has an unsanitized `innerHTML` assignment.

This spec addresses 5 targeted fixes identified by a security audit. Authentication is intentionally excluded — this is a portfolio project where frictionless demo access matters, and Cloudflare Tunnel provides edge-level protection.

## Changes

### 1. Collection Name Validation

**Problem:** Both services accept a `collection` parameter with no format or length constraints. Arbitrary strings are passed directly to Qdrant.

**Fix:** Validate against `^[a-zA-Z0-9_-]{1,100}$`. Return 422 if invalid.

- **Ingestion service (`/ingest` endpoint):** Validate the `collection` query parameter early in the handler, before any processing.
- **Chat service (`ChatRequest` model):** Add `Field(pattern=r"^[a-zA-Z0-9_-]{1,100}$")` to the `collection` field on the Pydantic model.

### 2. Question Length Limit

**Problem:** The chat `question` field has no length limit. Arbitrarily large prompts can be sent to Ollama.

**Fix:** Add `Field(max_length=2000)` to the `question` field on `ChatRequest`. FastAPI returns 422 automatically if exceeded.

### 3. Error Message Sanitization

**Problem:** Exception details are interpolated into HTTP responses and SSE events, leaking internal information (hostnames, ports, library errors).

**Affected locations:**
- `services/ingestion/app/main.py`: `f"Embedding service unavailable: {e}"` and `f"Vector store unavailable: {e}"`
- `services/chat/app/main.py`: `f"Backend service unavailable: {e}"` and `f"Internal error: {e}"`

**Fix:** Add `logging` to both services. Log the full error server-side with `logger.error(...)`. Return generic messages to clients (same text, just without `{e}`). Status codes remain unchanged.

### 4. DOMPurify for MermaidDiagram

**Problem:** `frontend/src/components/MermaidDiagram.tsx` directly sets DOM content from mermaid's render output without sanitization. While mermaid's output is safe today, this is vulnerable to supply chain compromise of the mermaid package.

**Fix:** Install `dompurify` and `@types/dompurify`. Wrap the mermaid SVG output in `DOMPurify.sanitize()` before assigning it to the DOM element.

### 5. Restrict CORS `allow_methods`

**Problem:** Both services use `allow_methods=["*"]`, allowing all HTTP methods including those not served by any endpoint.

**Fix:** Restrict to explicit method lists:
- **Ingestion service:** `["GET", "POST", "DELETE"]`
- **Chat service:** `["GET", "POST"]`

`allow_headers=["*"]` stays — this is standard for API services and accepted by the existing devsecops spec.

## Testing

- Add unit tests for collection name validation (valid names, invalid characters, too-long names)
- Add unit test for question length limit (2001-char question returns 422)
- Verify existing CORS tests still pass
- Verify existing endpoint tests still pass
- Run `ruff check` and `ruff format --check` on services
- Run `npx tsc --noEmit` on frontend

## Out of Scope

- Authentication/authorization (portfolio project, Cloudflare Tunnel provides edge protection)
- Rate limiting (defer to Cloudflare's built-in rate limiting)
- Prompt injection defenses (Ollama's context window is the natural limit)
- Shared validation module (not worth the coupling for 2 simple validators)
