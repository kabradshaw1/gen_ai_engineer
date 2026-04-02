import Link from "next/link";
import { MermaidDiagram } from "@/components/MermaidDiagram";

const architectureDiagram = `flowchart LR
  subgraph Ingestion["Document Ingestion"]
    direction LR
    A[PDF Upload] --> B[Parse\nPyMuPDF]
    B --> C[Chunk\nLangChain]
    C --> D[Embed\nnomic-embed-text]
    D --> E[(Qdrant)]
  end

  subgraph Query["Question Answering"]
    direction LR
    F[User Question] --> G[Embed\nnomic-embed-text]
    G --> H[Vector Search\nQdrant]
    H --> I[Build RAG Prompt]
    I --> J[Stream Response\nMistral 7B]
  end
`;

export default function AISection() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="mx-auto max-w-3xl px-6 py-12">
        {/* Navigation */}
        <Link
          href="/"
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          &larr; Home
        </Link>

        {/* Header */}
        <h1 className="mt-8 text-3xl font-bold">AI / Gen AI Engineer</h1>

        {/* Bio */}
        <section className="mt-8">
          <p className="text-muted-foreground leading-relaxed">
            [Placeholder: AI-focused bio. Describe your interest in generative
            AI, RAG architectures, and building intelligent systems. Highlight
            relevant experience and what excites you about the field.]
          </p>
        </section>

        {/* Project Explanation */}
        <section className="mt-12">
          <h2 className="text-2xl font-semibold">Document Q&A Assistant</h2>
          <p className="mt-4 text-muted-foreground leading-relaxed">
            A full-stack Retrieval-Augmented Generation (RAG) application that
            lets users upload PDF documents and ask questions about their
            content. The system parses, chunks, and embeds documents into a
            vector database, then retrieves relevant context to generate
            accurate, grounded answers using a local LLM.
          </p>

          <h3 className="mt-6 text-lg font-medium">Tech Stack</h3>
          <ul className="mt-2 list-disc pl-6 text-muted-foreground space-y-1">
            <li>FastAPI microservices (ingestion + chat)</li>
            <li>Qdrant vector database</li>
            <li>Ollama with Mistral 7B (chat) and nomic-embed-text (embeddings)</li>
            <li>Next.js + TypeScript + shadcn/ui frontend</li>
            <li>Docker Compose orchestration</li>
            <li>CI/CD with GitHub Actions, security scanning, E2E tests</li>
          </ul>
        </section>

        {/* Architecture Diagram */}
        <section className="mt-12">
          <h2 className="text-2xl font-semibold">How It Works</h2>
          <div className="mt-6 rounded-xl border border-foreground/10 bg-card p-6">
            <MermaidDiagram chart={architectureDiagram} />
          </div>
        </section>

        {/* Demo Link */}
        <section className="mt-12">
          <Link
            href="/ai/rag"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Try the Demo &rarr;
          </Link>
        </section>
      </div>
    </div>
  );
}
