import { dom } from "./js/dom.js";
import { onEnter } from "./js/utils.js";
import { renderTeiText } from "./js/tei.js";
import { loadFromApi } from "./js/api.js";
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

    // TEI toolbar
    els.renderBtn?.addEventListener("click", doRender);

    // Enter to load in key fields
    [els.baseUrl, els.datasetId, els.annotationId, els.pageNum].forEach((el) => onEnter(el, doLoad));

    // Index controller
    const getIndexUrl = () => {
        const b = (els.baseUrl.value || "").trim().replace(/\/+$/, "");
        const d = (els.datasetId.value || "").trim();
        const a = (els.annotationId.value || "").trim();
        if (!b || !d || !a) return "";
        return `${b}/datasets/${encodeURIComponent(d)}/annotations/${encodeURIComponent(a)}/index`;
    };

    const index = createIndexController({ els, getIndexUrl, jumpToPage });
    index.wire();
    await index.reload();
});
