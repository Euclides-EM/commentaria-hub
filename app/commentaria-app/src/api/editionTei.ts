import { OpenAPI } from '@hub-api'

/**
 * Fetches TEI XML for a single page from the edition endpoint.
 * GET /editions/{editionId}/tei/{pageNum}
 */
export async function fetchEditionTei(
  editionId: string,
  pageNum: string | number,
): Promise<string> {
  const base = OpenAPI.BASE.replace(/\/$/, '')
  const url = `${base}/editions/${encodeURIComponent(editionId)}/tei/${encodeURIComponent(String(pageNum))}`
  const token = await (typeof OpenAPI.TOKEN === 'function'
    ? OpenAPI.TOKEN({} as any)
    : Promise.resolve(OpenAPI.TOKEN))
  const headers: Record<string, string> = {
    Accept: 'application/xml, text/xml, text/plain, */*',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  // Use 'omit' to avoid CORS: server may send Access-Control-Allow-Origin: *
  // which is not allowed when credentials are 'include'. Auth is via Authorization header.
  const res = await fetch(url, { headers, credentials: 'omit' })
  if (!res.ok) {
    throw new Error(`Edition TEI failed: ${res.status} ${res.statusText}`)
  }
  return res.text()
}
