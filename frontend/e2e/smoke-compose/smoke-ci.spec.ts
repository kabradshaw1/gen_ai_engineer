import { test, expect } from "@playwright/test";
import path from "path";

// Target the local docker-compose stack rather than production. The compose
// stack uses mock-ollama, so the chat response is a known canned string.
const API_URL = process.env.SMOKE_API_URL || "http://localhost:8000";

// This string must match what services/mock-ollama/main.py emits in
// _chat_stream(). If the mock changes, update both places together.
const EXPECTED_CHAT_SUBSTRING = "mock response";

test.describe("compose-smoke CI tests", () => {
  test("backend health checks pass", async ({ request }) => {
    for (const svc of ["chat", "ingestion", "debug"]) {
      const res = await request.get(`${API_URL}/${svc}/health`);
      expect(res.ok(), `${svc}/health should return 2xx`).toBeTruthy();
    }
  });

  test("RAG happy path: upload → ask → streamed mock response", async ({
    request,
  }) => {
    const collection = "ci-smoke";

    // 1. Upload the fixture PDF to a dedicated collection.
    const pdfPath = path.join(__dirname, "..", "fixtures", "test.pdf");
    const fs = await import("fs");
    const pdfBuffer = fs.readFileSync(pdfPath);

    const uploadRes = await request.post(
      `${API_URL}/ingestion/ingest?collection=${collection}`,
      {
        multipart: {
          file: {
            name: "test.pdf",
            mimeType: "application/pdf",
            buffer: pdfBuffer,
          },
        },
      }
    );
    expect(uploadRes.ok(), "upload should succeed").toBeTruthy();

    // 2. Ask a question; response is a streamed NDJSON body from chat service.
    const chatRes = await request.post(`${API_URL}/chat/chat`, {
      data: {
        question: "What is this document about?",
        collection,
      },
    });
    expect(chatRes.ok(), "chat/chat should return 2xx").toBeTruthy();

    const body = await chatRes.text();
    expect(
      body.toLowerCase().includes(EXPECTED_CHAT_SUBSTRING),
      `chat response should contain "${EXPECTED_CHAT_SUBSTRING}"; got: ${body.slice(
        0,
        500
      )}`
    ).toBeTruthy();

    // 3. Cleanup: delete the collection so CI runs are idempotent.
    const delRes = await request.delete(
      `${API_URL}/ingestion/collections/${collection}`
    );
    expect(delRes.ok(), "collection delete should succeed").toBeTruthy();
  });
});
