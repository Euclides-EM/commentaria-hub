import { writeWindowMessage } from "./facsimilePdf.ts";

const escapeHTML = (value: string) =>
  value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");

const formatBytes = (bytes: number) => {
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(0)} KB`;
  }
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
};

export async function streamDownloadInNewTab(
  url: string,
  downloadName: string,
  bearerToken?: string | null,
): Promise<void> {
  const downloadWindow = window.open("", "_blank");
  if (!downloadWindow) {
    throw new Error("The browser blocked the download window.");
  }
  downloadWindow.opener = null;
  const title = `Downloading ${downloadName}`;
  const safeName = escapeHTML(downloadName);
  writeWindowMessage(downloadWindow, title, `Preparing ${safeName}...`, {
    loading: true,
  });

  try {
    const response = await fetch(url, {
      headers: bearerToken ? { Authorization: `Bearer ${bearerToken}` } : {},
    });
    if (!response.ok) {
      const details = (await response.text()).trim();
      throw new Error(
        `Download failed (${response.status})${details ? `: ${details}` : ""}`,
      );
    }
    const totalBytes = Number(response.headers.get("Content-Length")) || null;
    const contentType = response.headers.get("Content-Type") || undefined;
    const parts: Uint8Array[] = [];
    let loadedBytes = 0;
    let lastProgressAt = 0;
    if (response.body) {
      const reader = response.body.getReader();
      for (;;) {
        const { done, value } = await reader.read();
        if (done) {
          break;
        }
        parts.push(value);
        loadedBytes += value.byteLength;
        if (Date.now() - lastProgressAt < 250) {
          continue;
        }
        lastProgressAt = Date.now();
        writeWindowMessage(
          downloadWindow,
          title,
          totalBytes
            ? `Downloading ${safeName}... ${Math.min(100, Math.round((loadedBytes / totalBytes) * 100))}%`
            : `Downloading ${safeName}... ${formatBytes(loadedBytes)}`,
          { loading: true },
        );
      }
    } else {
      parts.push(new Uint8Array(await response.arrayBuffer()));
    }

    const blob = new Blob(
      parts as BlobPart[],
      contentType ? { type: contentType } : undefined,
    );
    const blobURL = URL.createObjectURL(blob);
    writeWindowMessage(
      downloadWindow,
      `Downloaded ${downloadName}`,
      `Downloaded ${safeName} (${formatBytes(blob.size)}). <a id="download-link" href="${blobURL}" download="${safeName}">Save again</a>`,
    );
    downloadWindow.document.getElementById("download-link")?.click();
    window.setTimeout(() => URL.revokeObjectURL(blobURL), 10 * 60_000);
  } catch (error) {
    writeWindowMessage(
      downloadWindow,
      "Download failed",
      escapeHTML(
        error instanceof Error ? error.message : "Failed to download the file.",
      ),
    );
    throw error;
  }
}
