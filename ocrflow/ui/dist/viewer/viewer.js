import { dom } from "./js/dom.js";
import { onEnter, escapeHtml } from "./js/utils.js";
import { renderTeiText } from "./js/tei.js";
import {
    loadFromApi,
    populateDatasetSelect,
    populateAnnotationSelect,
    getCachedAnnotations,
} from "./js/api.js";
import { wirePaging } from "./js/paging.js";
import { createIndexController } from "./js/index.js";
import { wireSidebarToggle } from "./js/sidebar.js";

document.addEventListener("DOMContentLoaded", async () => {
    const { els } = dom();

    // sidebar collapse
    wireSidebarToggle(els.toggleIndexBtn);

    // TEI rendering helpers
    const doRender = () =>
        renderTeiText({
            teiInput: els.teiInput,
            out: els.out,
            teiStatus: els.teiStatus,
        });

    // Load API wrapper
    const doLoad = () => loadFromApi({ els, renderTeiText: doRender });

    // Wire load + paging
    els.loadBtn?.addEventListener("click", doLoad);
    const { jumpToPage } = wirePaging({ els, loadFromApi: doLoad });

    // Index controller
    const getIndexUrl = () => {
        const b = (els.baseUrl?.value || "").trim().replace(/\/+$/, "");
        const d = (els.datasetId?.value || "").trim();
        const a = (els.annotationId?.value || "").trim();
        if (!b || !d || !a) return "";
        return `${b}/datasets/${encodeURIComponent(d)}/annotations/${encodeURIComponent(a)}/index`;
    };

    const index = createIndexController({ els, getIndexUrl, jumpToPage });
    index.wire();

    // --- Annotation details rendering ---
    const renderAnnotationDetails = () => {
        if (!els.annotationDetails) return;

        const datasetId = (els.datasetId?.value || "").trim();
        const annotationId = (els.annotationId?.value || "").trim();

        if (!datasetId || !annotationId) {
            els.annotationDetails.innerHTML = `<div class="empty">Select a dataset and an annotation.</div>`;
            return;
        }

        const anns = getCachedAnnotations(datasetId) || [];
        const ann = anns.find((x) => String(x?.id || "").trim() === annotationId);

        if (!ann) {
            els.annotationDetails.innerHTML = `<div class="empty">Annotation details not available.</div>`;
            return;
        }

        const appliedRules = ann.applied_rules
            ? JSON.stringify(ann.applied_rules, null, 2)
            : "null";

        els.annotationDetails.innerHTML = `
            <div class="kv">
                <div class="k">Name</div><div class="v">${escapeHtml(String(ann.name || ""))}</div>
                <div class="k">ID</div><div class="v mono">${escapeHtml(String(ann.id || ""))}</div>
                <div class="k">Dataset</div><div class="v mono">${escapeHtml(String(ann.dataset_id || ""))}</div>
                <div class="k">Pages</div><div class="v">${escapeHtml(String(ann.pages || ""))}</div>
                <div class="k">Segmented</div><div class="v">${escapeHtml(String(!!ann.segmented))}</div>
                <div class="k">Ground truth</div><div class="v">${escapeHtml(String(!!ann.ground_truth))}</div>
                <div class="k">OCRed</div><div class="v">${escapeHtml(String(!!ann.ocred))}</div>
                <div class="k">Created</div><div class="v mono">${escapeHtml(String(ann.created_at || ""))}</div>
                <div class="k">Updated</div><div class="v mono">${escapeHtml(String(ann.updated_at || ""))}</div>
                <div class="k">Description</div><div class="v">${escapeHtml(String(ann.description || ""))}</div>
                <div class="k">Applied rules</div><div class="v"><div class="pre">${escapeHtml(appliedRules)}</div></div>
            </div>
        `;
    };

    // --- View state (no cycles) ---
    let showSource = false;
    let showAnnotation = false;

    const applyView = () => {
        // Source controls
        els.teiInput?.classList.toggle("hidden", !showSource);
        els.teiSourceTools?.classList.toggle("hidden", !showSource);

        // Annotation replaces rendered TEI
        els.annotationDetails?.classList.toggle("hidden", !showAnnotation);
        els.out?.classList.toggle("hidden", showAnnotation);

        // Button visual states
        els.toggleTeiSourceBtn?.classList.toggle("active", showSource);
        els.toggleAnnotationBtn?.classList.toggle("active", showAnnotation);

        if (showAnnotation) renderAnnotationDetails();
    };

    // Default: rendered TEI
    applyView();

    // TEI source toggle (like before)
    els.toggleTeiSourceBtn?.addEventListener("click", () => {
        showSource = !showSource;
        applyView();
    });

    // Annotation toggle (explicit, no cycle)
    els.toggleAnnotationBtn?.addEventListener("click", () => {
        showAnnotation = !showAnnotation;
        showSource = false; // make it predictable
        applyView();
    });

    // Populate datasets then annotations
    await populateDatasetSelect(els);
    await populateAnnotationSelect(els);

    // Base URL change
    els.baseUrl?.addEventListener("change", async () => {
        await populateDatasetSelect(els);
        await populateAnnotationSelect(els);
        await index.reload();

        if (showAnnotation) renderAnnotationDetails();
    });

    // Dataset change
    els.datasetId?.addEventListener("change", async () => {
        await populateAnnotationSelect(els, { preferredId: "" });
        await index.reload();

        if (showAnnotation) renderAnnotationDetails();
    });

    // Annotation change
    els.annotationId?.addEventListener("change", async () => {
        await index.reload();
        if (showAnnotation) renderAnnotationDetails();
        // doLoad(); // optional
    });

    // TEI toolbar
    els.renderBtn?.addEventListener("click", doRender);

    // Enter to load
    [els.baseUrl, els.pageNum].forEach((el) => onEnter(el, doLoad));

    // Initial index load
    await index.reload();
});
