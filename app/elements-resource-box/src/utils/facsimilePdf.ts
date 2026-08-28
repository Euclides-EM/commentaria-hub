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
    const response = await fetch(facsimilePDFURL(facsimileId), {
      headers: {
        Accept: "application/pdf",
        Authorization: `Bearer ${bearerToken}`,
      },
    });
    if (!response.ok) {
      writeWindowMessage(
        pdfWindow,
        "Scan unavailable",
        `Opening the scan failed (${response.status}).`,
      );
      throw new Error(`Opening the scan failed (${response.status}).`);
    }

    const pdfBlob = await response.blob();
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
