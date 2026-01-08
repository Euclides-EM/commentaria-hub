import { escapeHtml, normalizeBaseUrl } from "./utils.js";
import { annotationLabel, datasetLabel } from "./labels.js";

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

export async function fetchDatasets(baseUrlEl) {
    const b = normalizeBaseUrl(baseUrlEl.value);
    const url = `${b}/datasets`;

    const res = await fetch(url, { headers: { Accept: "application/json" } });
    if (!res.ok) throw new Error(`Datasets HTTP ${res.status}`);

    const data = await res.json();
    if (!Array.isArray(data)) throw new Error("Datasets response is not an array");

    return data;
}

export async function populateDatasetSelect(els, opts = {}) {
    const { preferredId = "" } = opts;

    if (!els.datasetId) return;

    // Remember current selection if any
    const keepId = (preferredId || els.datasetId.value || "").trim();

    els.datasetId.innerHTML = `<option value="">Loading…</option>`;
    els.datasetId.disabled = true;

    try {
        const datasets = await fetchDatasets(els.baseUrl);

        // Sort for stable UX
        datasets.sort((a, b) => {
            const ae = (a?.edition_id || "").localeCompare(b?.edition_id || "");
            if (ae !== 0) return ae;
            return (a?.name || a?.id || "").localeCompare(b?.name || b?.id || "");
        });

        const optionsHtml = datasets
            .map((ds) => {
                const id = String(ds?.id || "").trim();
                const label = datasetLabel(ds);
                return `<option value="${escapeHtml(id)}">${escapeHtml(label)}</option>`;
            })
            .join("");

        els.datasetId.innerHTML = optionsHtml || `<option value="">No datasets found</option>`;
        els.datasetId.disabled = false;

        // Restore selection if possible, else pick first dataset option
        const options = [...els.datasetId.options];
        const keepOpt = keepId ? options.find((o) => o.value === keepId) : null;
        const firstDataset = options.find((o) => o.value) || null;

        if (keepOpt) {
            els.datasetId.value = keepOpt.value;
        } else if (firstDataset) {
            els.datasetId.value = firstDataset.value;
        } else {
            els.datasetId.value = "";
        }
    } catch (e) {
        els.datasetId.innerHTML = `<option value="">Error loading datasets</option>`;
        els.datasetId.disabled = false;

        if (els.reqSummary) {
            els.reqSummary.textContent = `Failed to load datasets: ${String(e?.message || e)}`;
        }
    }
}

export async function loadFromApi(ctx) {
    const { els, renderTeiText } = ctx;
    const { teiUrl, imgUrl } = buildUrls(els);

    if (els.reqSummary) els.reqSummary.textContent = `GET ${imgUrl} and ${teiUrl}`;
    if (els.imgStatus) els.imgStatus.textContent = "Loading…";
    if (els.teiStatus) els.teiStatus.textContent = "Loading…";

    const imgWithBust = `${imgUrl}${imgUrl.includes("?") ? "&" : "?"}t=${Date.now()}`;
    els.pageImg.onload = () => (els.imgStatus.textContent = "Loaded");
    els.pageImg.onerror = () => (els.imgStatus.textContent = "Error loading image");
    els.pageImg.src = imgWithBust;

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

// simple in-memory cache keyed by dataset id
const annotationsCache = new Map(); // datasetId -> annotations[]

export function getCachedAnnotations(datasetId) {
    const key = String(datasetId || "").trim();
    return annotationsCache.get(key) || [];
}

export async function fetchAnnotations(baseUrlEl, datasetId) {
    const b = normalizeBaseUrl(baseUrlEl.value);
    const d = encodeURIComponent(String(datasetId || "").trim());
    const url = `${b}/datasets/${d}/annotations`;

    const res = await fetch(url, { headers: { Accept: "application/json" } });
    if (!res.ok) throw new Error(`Annotations HTTP ${res.status}`);

    const data = await res.json();
    if (data === null) return [];
    if (!Array.isArray(data)) throw new Error("Annotations response is not an array");

    // store in cache
    annotationsCache.set(String(datasetId || "").trim(), data);

    return data;
}

export async function populateAnnotationSelect(els, opts = {}) {
    const { preferredId = "" } = opts;
    if (!els.annotationId) return;

    const datasetId = (els.datasetId?.value || "").trim();
    const keepId = (preferredId || els.annotationId.value || "").trim();

    if (!datasetId) {
        els.annotationId.innerHTML = `<option value="">Select dataset…</option>`;
        els.annotationId.disabled = true;
        return;
    }

    els.annotationId.innerHTML = `<option value="">Loading…</option>`;
    els.annotationId.disabled = true;

    try {
        const annotations = await fetchAnnotations(els.baseUrl, datasetId);
        annotations.sort((a, b) => (a?.name || a?.id || "").localeCompare(b?.name || b?.id || ""));

        const only = !!els.onlyTranscribed && els.onlyTranscribed.checked;
        const placeholderHtml = `<option value="" disabled selected>Select annotation…</option>`;

        const optionsHtml = annotations
            .map((a) => {
                const id = String(a?.id || "").trim();
                const label = annotationLabel(a);
                const disabled = only && a?.ocred === false;

                return `<option value="${escapeHtml(id)}" ${disabled ? "disabled" : ""}>${escapeHtml(
                    label
                )}</option>`;
            })
            .join("");

        els.annotationId.innerHTML =
            placeholderHtml + (optionsHtml || `<option value="" disabled>No annotations found</option>`);
        els.annotationId.disabled = false;

        const options = [...els.annotationId.options];
        const keepOpt = keepId ? options.find((o) => o.value === keepId && !o.disabled) : null;
        const firstEnabled = options.find((o) => o.value !== "" && !o.disabled) || null;

        if (keepOpt) {
            els.annotationId.value = keepOpt.value;
        } else if (firstEnabled) {
            els.annotationId.value = firstEnabled.value;
        } else {
            els.annotationId.value = "";
        }
    } catch (e) {
        els.annotationId.innerHTML = `<option value="">Error loading annotations</option>`;
        els.annotationId.disabled = false;

        if (els.reqSummary) {
            els.reqSummary.textContent = `Failed to load annotations: ${String(e?.message || e)}`;
        }
    }
}
