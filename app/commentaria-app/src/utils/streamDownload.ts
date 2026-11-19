const escapeHTML = (value: string) =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')

const writeWindowMessage = (
  targetWindow: Window,
  title: string,
  body: string,
  options?: { loading?: boolean },
) => {
  targetWindow.document.title = title
  targetWindow.document.body.innerHTML = `
    <main style="min-height: 100vh; display: grid; place-items: center; margin: 0; font-family: sans-serif; background: #f8fafc; color: #0f172a;">
      <section style="display: flex; flex-direction: column; align-items: center; gap: 1rem; padding: 2rem; text-align: center;">
        ${
          options?.loading
            ? '<div style="width: 2.5rem; height: 2.5rem; border: 3px solid #cbd5e1; border-top-color: #0f172a; border-radius: 9999px; animation: download-spin 0.8s linear infinite;" aria-hidden="true"></div>'
            : ''
        }
        <p style="margin: 0; line-height: 1.5;">${body}</p>
      </section>
    </main>
    <style>
      @keyframes download-spin {
        to { transform: rotate(360deg); }
      }
      body { margin: 0; }
    </style>
  `
}

const formatBytes = (bytes: number) => {
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(0)} KB`
  }
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

export async function streamDownloadInNewTab(
  url: string,
  downloadName: string,
  bearerToken?: string | null,
): Promise<void> {
  const downloadWindow = window.open('', '_blank')
  if (!downloadWindow) {
    throw new Error('The browser blocked the download window.')
  }
  downloadWindow.opener = null
  const title = `Downloading ${downloadName}`
  const safeName = escapeHTML(downloadName)
  writeWindowMessage(downloadWindow, title, `Preparing ${safeName}...`, {
    loading: true,
  })

  try {
    const response = await fetch(url, {
      headers: bearerToken ? { Authorization: `Bearer ${bearerToken}` } : {},
    })
    if (!response.ok) {
      const details = (await response.text()).trim()
      throw new Error(
        `Download failed (${response.status})${details ? `: ${details}` : ''}`,
      )
    }
    const totalBytes = Number(response.headers.get('Content-Length')) || null
    const contentType = response.headers.get('Content-Type') || undefined
    const parts: Uint8Array[] = []
    let loadedBytes = 0
    let lastProgressAt = 0
    if (response.body) {
      const reader = response.body.getReader()
      for (;;) {
        const { done, value } = await reader.read()
        if (done) {
          break
        }
        parts.push(value)
        loadedBytes += value.byteLength
        if (Date.now() - lastProgressAt < 250) {
          continue
        }
        lastProgressAt = Date.now()
        writeWindowMessage(
          downloadWindow,
          title,
          totalBytes
            ? `Downloading ${safeName}... ${Math.min(100, Math.round((loadedBytes / totalBytes) * 100))}%`
            : `Downloading ${safeName}... ${formatBytes(loadedBytes)}`,
          { loading: true },
        )
      }
    } else {
      parts.push(new Uint8Array(await response.arrayBuffer()))
    }

    const blob = new Blob(
      parts as BlobPart[],
      contentType ? { type: contentType } : undefined,
    )
    const blobURL = URL.createObjectURL(blob)
    writeWindowMessage(
      downloadWindow,
      `Downloaded ${downloadName}`,
      `Downloaded ${safeName} (${formatBytes(blob.size)}). <a id="download-link" href="${blobURL}" download="${safeName}">Save again</a>`,
    )
    downloadWindow.document.getElementById('download-link')?.click()
    window.setTimeout(() => URL.revokeObjectURL(blobURL), 10 * 60_000)
  } catch (error) {
    writeWindowMessage(
      downloadWindow,
      'Download failed',
      escapeHTML(
        error instanceof Error ? error.message : 'Failed to download the file.',
      ),
    )
    throw error
  }
}
