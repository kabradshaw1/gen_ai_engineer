# Portfolio Update Design

## Context

The frontend is currently a single-page Document Q&A Assistant at `/`. Kyle wants to restructure it into a portfolio site where `/` is a general bio with links to job-specific sections. The first section showcases this repo's AI/RAG project. The existing Q&A demo moves to a nested route.

## Route Structure

| Route     | Purpose                                              |
|-----------|------------------------------------------------------|
| `/`       | General bio landing page + links to job sections     |
| `/ai`     | AI/Gen AI Engineer section: bio, project explanation, architecture diagram |
| `/ai/rag` | Document Q&A Assistant demo (current home page content) |

## Page Designs

### `/` — Portfolio Landing Page

- Kyle's name as heading
- Placeholder bio text (Kyle will write the real content later)
- Section links rendered as cards — each card links to a job-specific section
- First card: "AI / Gen AI Engineer" → `/ai`
- Extensible: adding a future section (e.g., `/fullstack`) is just adding another card
- Uses existing dark theme and shadcn/ui components (Card, Button)

### `/ai` — AI Section

- Short AI-focused bio (placeholder text)
- Project explanation: what the Document Q&A Assistant does, the tech stack, and why it was built
- Mermaid.js architecture diagram showing two flows:
  - **Ingestion flow:** PDF Upload → Parse (PyMuPDF) → Chunk (LangChain) → Embed (nomic-embed-text) → Store (Qdrant)
  - **Query flow:** Question → Embed → Vector Search (Qdrant) → Build RAG Prompt → Stream Response (Mistral 7B via Ollama)
- "Try the Demo" link/button → `/ai/rag`
- Uses existing dark theme and shadcn/ui components

### `/ai/rag` — Demo Page

- Exact current home page content moved here unchanged
- Components: ChatWindow, FileUpload, DocumentList, MessageInput, SourceBadge
- API calls to ingestion (8001) and chat (8002) services remain the same
- No changes to component logic

## Technical Approach

### Next.js App Router Structure

```
frontend/src/app/
├── layout.tsx          # Existing root layout (unchanged)
├── globals.css         # Existing styles (unchanged)
├── page.tsx            # NEW: Portfolio landing page
├── ai/
│   ├── page.tsx        # NEW: AI section page
│   └── rag/
│       └── page.tsx    # MOVED: Current home page content
```

### Dependencies

- `mermaid` npm package for rendering the architecture diagram on `/ai`
- No other new dependencies

### Navigation

- Next.js `Link` component for all internal navigation
- Back links on `/ai` → `/` and `/ai/rag` → `/ai` for easy navigation
- No global nav bar needed yet (can be added later as more sections are created)

### Existing Components

All existing components (ChatWindow, MessageInput, FileUpload, DocumentList, SourceBadge) are reused as-is on `/ai/rag`. No modifications needed.

### Mermaid Diagram

Client-side rendered Mermaid component wrapping a diagram definition. The diagram shows the two main flows (ingestion and query) with labeled nodes for each step.

## What Changes

1. `frontend/src/app/page.tsx` — rewritten from Q&A app to portfolio landing page
2. `frontend/src/app/ai/page.tsx` — new AI section page
3. `frontend/src/app/ai/rag/page.tsx` — current home page content moved here
4. New Mermaid wrapper component for diagram rendering
5. `mermaid` added to package.json

## What Doesn't Change

- All existing React components (ChatWindow, FileUpload, DocumentList, MessageInput, SourceBadge)
- All UI components (button, card, input, badge, scroll-area, popover)
- Backend services, API endpoints, Docker setup
- E2E tests (will need path updates)
- Root layout, globals.css, theme

## Verification

1. `npm run dev` — verify all three routes render correctly
2. `/` shows bio and AI section link
3. `/ai` shows project explanation and Mermaid diagram
4. `/ai/rag` has the full Q&A demo working (chat, upload, document management)
5. Navigation between pages works via Next.js Link
6. `npx tsc --noEmit` passes
7. Existing E2E tests updated for new paths if needed
