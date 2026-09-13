import { OpenAPI } from "@hub-api";

const facsimilePDFURL = (facsimileId: string) =>
  `${OpenAPI.BASE.replace(/\/$/, "")}/facsimilies/${encodeURIComponent(facsimileId)}/pdf`;

const writeWindowMessage = (
  targetWindow: Window,
  title: string,
  body: string,
  options?: { loading?: boolean },
) => {
  targetWindow.document.title = title;
  targetWindow.document.body.innerHTML = `
    <main style="min-height: 100vh; display: grid; place-items: center; margin: 0; font-family: sans-serif; background: #f8fafc; color: #0f172a;">
      <section style="display: flex; flex-direction: column; align-items: center; gap: 1rem; padding: 2rem; text-align: center;">
        ${
          options?.loading
            ? '<div style="width: 2.5rem; height: 2.5rem; border: 3px solid #cbd5e1; border-top-color: #0f172a; border-radius: 9999px; animation: facsimile-spin 0.8s linear infinite;" aria-hidden="true"></div>'
            : ""
        }
        <p style="margin: 0; line-height: 1.5;">${body}</p>
      </section>
    </main>
    <style>
      @keyframes facsimile-spin {
        to { transform: rotate(360deg); }
      }
      body { margin: 0; }
    </style>
  `;
};

const CHUNK_SIZE = 25 * 1024 * 1024;
const MAX_RETRIES_PER_CHUNK = 4;
const RETRY_DELAY_MS = 1_000;

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const fetchChunk = async (
  url: string,
  bearerToken: string,
  start: number,
  end: number,
): Promise<Response> =>
  fetch(url, {
    headers: {
      Accept: "application/pdf",
      Authorization: `Bearer ${bearerToken}`,
      Range: `bytes=${start}-${end}`,
    },
  });

const fetchChunkWithRetry = async (
  url: string,
  bearerToken: string,
  start: number,
  end: number,
): Promise<Response> => {
  let lastError: unknown;
  for (let attempt = 0; attempt < MAX_RETRIES_PER_CHUNK; attempt++) {
    try {
      const response = await fetchChunk(url, bearerToken, start, end);
      if (response.ok) {
        return response;
      }
      lastError = new Error(`Opening the scan failed (${response.status}).`);
    } catch (error) {
      lastError = error;
    }
    await sleep(RETRY_DELAY_MS * (attempt + 1));
  }
  throw lastError instanceof Error
    ? lastError
    : new Error("Failed to open the scan.");
};

async function downloadPDFWithRetry(
  facsimileId: string,
  bearerToken: string,
  onProgress: (loadedBytes: number, totalBytes: number | null) => void,
): Promise<Blob> {
  const url = facsimilePDFURL(facsimileId);
  const firstChunk = await fetchChunkWithRetry(
    url,
    bearerToken,
    0,
    CHUNK_SIZE - 1,
  );
  const contentRange = firstChunk.headers.get("Content-Range");
  const totalBytes = contentRange
    ? Number(contentRange.split("/")[1])
    : Number(firstChunk.headers.get("Content-Length")) || null;
  const contentType = firstChunk.headers.get("Content-Type") || undefined;

  const parts: Blob[] = [await firstChunk.blob()];
  let loadedBytes = parts[0].size;
  onProgress(loadedBytes, totalBytes);

  if (firstChunk.status === 200 || !totalBytes) {
    return new Blob(parts, contentType ? { type: contentType } : undefined);
  }

  while (loadedBytes < totalBytes) {
    const start = loadedBytes;
    const end = Math.min(start + CHUNK_SIZE, totalBytes) - 1;
    const chunk = await fetchChunkWithRetry(url, bearerToken, start, end);
    const chunkBlob = await chunk.blob();
    parts.push(chunkBlob);
    loadedBytes += chunkBlob.size;
    onProgress(loadedBytes, totalBytes);
  }

  return new Blob(parts, contentType ? { type: contentType } : undefined);
}

export async function openAuthenticatedFacsimilePDF(
  facsimileId: string,
  bearerToken: string,
  pageNumber?: number,
  downloadName?: string,
): Promise<void> {
  const pdfWindow = window.open("", "_blank");
  if (!pdfWindow) {
    throw new Error("The browser blocked the PDF window.");
  }
  pdfWindow.opener = null;
  writeWindowMessage(pdfWindow, "Opening scan", "Loading scan...", {
    loading: true,
  });

  try {
    const pdfBlob = await downloadPDFWithRetry(
      facsimileId,
      bearerToken,
      (loadedBytes, totalBytes) => {
        const percent = totalBytes
          ? Math.min(100, Math.round((loadedBytes / totalBytes) * 100))
          : null;
        writeWindowMessage(
          pdfWindow,
          "Opening scan",
          percent === null ? "Loading scan..." : `Loading scan... ${percent}%`,
          { loading: true },
        );
      },
    );
    const pdfFile = new File([pdfBlob], downloadName || `${facsimileId}.pdf`, {
      type: pdfBlob.type || "application/pdf",
    });
    const pdfURL = URL.createObjectURL(pdfFile);
    const pageFragment = pageNumber === undefined ? "" : `#page=${pageNumber}`;
    pdfWindow.location.replace(`${pdfURL}${pageFragment}`);
    window.setTimeout(() => URL.revokeObjectURL(pdfURL), 60_000);
  } catch (error) {
    writeWindowMessage(
      pdfWindow,
      "Scan unavailable",
      error instanceof Error ? error.message : "Failed to open the scan.",
    );
    throw error;
  }
}
