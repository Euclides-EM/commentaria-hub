import { normalizeBaseUrl, escapeHtml } from "./utils.js";

export function buildUrls({ baseUrl, datasetId, annotationId, pageNum }) {
    const b = normalizeBaseUrl(baseUrl.value);
    const d = encodeURIComponent(datasetId.value.trim());
    const a = encodeURIComponent(annotationId.value.trim());
    const p = String(pageNum.value).trim();

    return {
        teiUrl: `${b}/datasets/${d}/annotations/${a}/tei/${p}`,
        imgUrl: `${b}/datasets/${d}/images/${p}`,
    };
}

export async function loadFromApi(ctx) {
    const { els, renderTeiText } = ctx;
    const { teiUrl, imgUrl } = buildUrls(els);

    if (els.reqSummary) els.reqSummary.textContent = `GET ${imgUrl} and ${teiUrl}`;
    if (els.imgStatus) els.imgStatus.textContent = "Loading…";
    if (els.teiStatus) els.teiStatus.textContent = "Loading…";

    // image
    const imgWithBust = `${imgUrl}${imgUrl.includes("?") ? "&" : "?"}t=${Date.now()}`;
    els.pageImg.onload = () => (els.imgStatus.textContent = "Loaded");
    els.pageImg.onerror = () => (els.imgStatus.textContent = "Error loading image");
    els.pageImg.src = imgWithBust;

    // TEI
    try {
        const res = await fetch(teiUrl, { headers: { Accept: "application/xml,text/xml,*/*" } });
        if (!res.ok) throw new Error(`TEI HTTP ${res.status}`);
        const xml = await res.text();
        els.teiInput.value = xml;
        els.teiStatus.textContent = "Loaded";
        renderTeiText();
    } catch (e) {
        els.teiInput.value = "";
        els.out.innerHTML = `<div class="empty"><span class="mono">${escapeHtml(String(e.message || e))}</span></div>`;
        els.teiStatus.textContent = "Error";
    }
}
