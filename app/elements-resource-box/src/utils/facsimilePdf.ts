import { OpenAPI } from "@hub-api";

const facsimilePDFURL = (editionKey: string) =>
  `${OpenAPI.BASE.replace(/\/$/, "")}/editions/${encodeURIComponent(editionKey)}/facsimile.pdf`;

export async function openAuthenticatedFacsimilePDF(
  editionKey: string,
  bearerToken: string,
  pageNumber?: number,
): Promise<void> {
  const pdfWindow = window.open("", "_blank");
  if (!pdfWindow) {
    throw new Error("The browser blocked the PDF window.");
  }
  pdfWindow.opener = null;

  try {
    const response = await fetch(facsimilePDFURL(editionKey), {
      headers: {
        Accept: "application/pdf",
        Authorization: `Bearer ${bearerToken}`,
      },
    });
    if (!response.ok) {
      throw new Error(`Opening the main scan failed (${response.status}).`);
    }

    const pdfURL = URL.createObjectURL(await response.blob());
    const pageFragment = pageNumber === undefined ? "" : `#page=${pageNumber}`;
    pdfWindow.location.replace(`${pdfURL}${pageFragment}`);

    window.setTimeout(() => URL.revokeObjectURL(pdfURL), 60_000);
  } catch (error) {
    pdfWindow.close();
    throw error;
  }
}
