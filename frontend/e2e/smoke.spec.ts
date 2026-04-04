import { test, expect } from "@playwright/test";
import path from "path";

const FRONTEND_URL =
  process.env.SMOKE_FRONTEND_URL || "https://kylebradshaw.dev";
const API_URL =
  process.env.SMOKE_API_URL || "https://api.kylebradshaw.dev";

test.describe("Production smoke tests", () => {
  test("frontend loads", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/ai/rag`);
    await expect(
      page.locator("h1", { hasText: "Document Q&A Assistant" })
    ).toBeVisible();
    await expect(
      page.getByPlaceholder("Ask a question about your documents...")
    ).toBeVisible();
  });

  test("backend health checks pass", async ({ request }) => {
    const chatHealth = await request.get(`${API_URL}/chat/health`);
    expect(chatHealth.ok()).toBeTruthy();
    const chatData = await chatHealth.json();
    expect(chatData.status).toBe("healthy");

    const ingestionHealth = await request.get(`${API_URL}/ingestion/health`);
    expect(ingestionHealth.ok()).toBeTruthy();
    const ingestionData = await ingestionHealth.json();
    expect(ingestionData.status).toBe("healthy");
  });

  test("full E2E flow with cleanup", async ({ request }) => {
    const testCollection = "e2e-test";

    // Step 1: Upload test PDF to dedicated collection
    const pdfPath = path.join(__dirname, "fixtures", "test.pdf");
    const fs = await import("fs");
    const pdfBuffer = fs.readFileSync(pdfPath);

    const uploadResponse = await request.post(
      `${API_URL}/ingestion/ingest?collection=${testCollection}`,
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
    expect(uploadResponse.ok()).toBeTruthy();
    const uploadData = await uploadResponse.json();
    expect(uploadData.status).toBe("success");
    expect(uploadData.chunks_created).toBeGreaterThan(0);

    // Step 2: Ask a question against the test collection
    const chatResponse = await request.post(`${API_URL}/chat/chat`, {
      data: {
        question: "What is artificial intelligence?",
        collection: testCollection,
      },
    });
    expect(chatResponse.ok()).toBeTruthy();
    const chatBody = await chatResponse.text();
    expect(chatBody).toContain("data:");

    // Step 3: Cleanup — delete the test collection
    const deleteResponse = await request.delete(
      `${API_URL}/ingestion/collections/${testCollection}`
    );
    expect(deleteResponse.ok()).toBeTruthy();
    const deleteData = await deleteResponse.json();
    expect(deleteData.status).toBe("deleted");
  });
});
