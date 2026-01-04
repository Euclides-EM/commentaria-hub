import { normalizeText, escapeHtml } from "./utils.js";

function nodeLabel(node) {
    const cat = normalizeText(node.category);
    const txt = normalizeText(node.content);
    const page = node?.location?.page;
    const pageStr = Number.isFinite(page) ? `p.${page}` : "";
    return { txt, cat, pageStr, page };
}

function markActiveIndexNode(el) {
    document.querySelectorAll(".indexActive").forEach((x) => x.classList.remove("indexActive"));
    el.classList.add("indexActive");
    el.scrollIntoView({ block: "nearest" });
}

function buildTreeNode(node, filterLower, jumpToPage) {
    const { txt, cat, pageStr, page } = nodeLabel(node);
    const children = Array.isArray(node.children) ? node.children : null;

    const selfText = `${txt} ${cat} ${pageStr}`.toLowerCase();
    const selfMatches = !filterLower || selfText.includes(filterLower);

    const childEls = [];
    let anyChildMatches = false;

    if (children?.length) {
        for (const ch of children) {
            const built = buildTreeNode(ch, filterLower, jumpToPage);
            if (built) {
                anyChildMatches = true;
                childEls.push(built);
            }
        }
    }

    if (!selfMatches && !anyChildMatches) return null;

    const wrapper = document.createElement("div");
    wrapper.className = "indexNode";

    if (children?.length) {
        const details = document.createElement("details");
        const summary = document.createElement("summary");

        summary.innerHTML = `
      <span>${escapeHtml(txt || "(untitled)")}</span>
      ${pageStr ? `<span class="meta">${escapeHtml(pageStr)}</span>` : ""}
      ${cat ? `<span class="meta">${escapeHtml(cat)}</span>` : ""}
    `;

        summary.addEventListener("click", () => {
            if (Number.isFinite(page)) jumpToPage(page);
            markActiveIndexNode(summary);
        });

        details.appendChild(summary);

        const kids = document.createElement("div");
        kids.style.marginLeft = "14px";
        for (const el of childEls) kids.appendChild(el);
        details.appendChild(kids);

        if (filterLower) details.open = true;

        wrapper.appendChild(details);
    } else {
        const leaf = document.createElement("div");
        leaf.className = "leaf";
        leaf.innerHTML = `
      <span>${escapeHtml(txt || "(untitled)")}</span>
      ${pageStr ? `<span class="meta">${escapeHtml(pageStr)}</span>` : ""}
      ${cat ? `<span class="meta">${escapeHtml(cat)}</span>` : ""}
    `;
        leaf.addEventListener("click", () => {
            if (Number.isFinite(page)) jumpToPage(page);
            markActiveIndexNode(leaf);
        });
        wrapper.appendChild(leaf);
    }

    return wrapper;
}

function renderIndexTree(rootEl, payload, filterText, jumpToPage) {
    rootEl.innerHTML = "";
    if (!payload || !Array.isArray(payload.nodes)) return;

    const filterLower = normalizeText(filterText).toLowerCase();

    const frag = document.createDocumentFragment();
    for (const n of payload.nodes) {
        const el = buildTreeNode(n, filterLower, jumpToPage);
        if (el) frag.appendChild(el);
    }
    rootEl.appendChild(frag);
}

export function createIndexController({ els, getIndexUrl, jumpToPage }) {
    let currentPayload = null;

    const setIndexStatus = (msg) => {
        if (els.indexStatus) els.indexStatus.textContent = msg || "";
    };

    async function fetchIndex() {
        const url = getIndexUrl();
        if (!url) {
            setIndexStatus("Missing baseUrl, datasetId, or annotationId");
            return null;
        }

        setIndexStatus("Loading…");
        try {
            const res = await fetch(url, { headers: { Accept: "application/json" } });
            if (!res.ok) throw new Error(`Index fetch failed: ${res.status} ${res.statusText}`);
            const payload = await res.json();
            setIndexStatus(`Loaded (${(payload.nodes || []).length} root nodes)`);
            return payload;
        } catch (e) {
            console.error(e);
            setIndexStatus("Failed to load");
            return null;
        }
    }

    async function reload() {
        currentPayload = await fetchIndex();
        const filterText = els.indexFilter?.value || "";
        renderIndexTree(els.indexTree, currentPayload, filterText, jumpToPage);
    }

    function wire() {
        els.reloadIndexBtn?.addEventListener("click", reload);

        els.indexFilter?.addEventListener("input", () => {
            renderIndexTree(els.indexTree, currentPayload, els.indexFilter.value, jumpToPage);
        });

        // reload when dataset or annotation changes
        els.datasetId?.addEventListener("change", reload);
        els.annotationId?.addEventListener("change", reload);
    }

    return { wire, reload };
}
