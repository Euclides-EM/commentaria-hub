const baseUrl = document.getElementById("baseUrl");
const datasetId = document.getElementById("datasetId");
const annotationId = document.getElementById("annotationId");
const pageNum = document.getElementById("pageNum");

const pageInput = document.getElementById("pageNum");
const loadBtn   = document.getElementById("loadBtn");
const prevBtn   = document.getElementById("prevPage");
const nextBtn   = document.getElementById("nextPage");

const pageImg = document.getElementById("pageImg");
const teiInput = document.getElementById("teiInput");
const out = document.getElementById("out");

const imgStatus = document.getElementById("imgStatus");
const teiStatus = document.getElementById("teiStatus");
const reqSummary = document.getElementById("reqSummary");

const showPB = document.getElementById("showPB");
const keepEmptyP = document.getElementById("keepEmptyP");

function escapeHtml(s) {
    return s
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;");
}

function isElement(node, localName) {
    return node && node.nodeType === Node.ELEMENT_NODE && node.localName === localName;
}

function parseXml(xmlString) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(xmlString, "application/xml");
    const pe = doc.getElementsByTagName("parsererror")[0];
    if (pe) throw new Error("XML parse error (often unescaped & or mismatched tags).");
    return doc;
}

// Convert a TEI node subtree into HTML text, respecting <lb/> and <pb/>
function toReadingHtml(node, opts) {
    let html = "";

    for (const child of node.childNodes) {
        if (child.nodeType === Node.TEXT_NODE) {
            html += escapeHtml(child.nodeValue);
            continue;
        }

        if (child.nodeType !== Node.ELEMENT_NODE) continue;

        if (isElement(child, "lb")) {
            html += "<br>";
            continue;
        }

        if (isElement(child, "pb")) {
            if (opts.showPB) {
                const facs = child.getAttribute("facs") || "";
                const n = child.getAttribute("n") || "";
                const label = (n || facs) ? `Page break ${ escapeHtml(n || facs) }` : "Page break";
                html += `<div class="pb">${ label }</div>`;
            }
            continue;
        }

        html += toReadingHtml(child, opts);
    }

    return html;
}

function renderTeiText() {
    const xml = teiInput.value.trim();
    if (!xml) {
        out.innerHTML = '<div class="empty">No TEI loaded.</div>';
        teiStatus.textContent = "";
        return;
    }

    try {
        const doc = parseXml(xml);

        const body = doc.getElementsByTagNameNS("*", "body")[0] || doc.getElementsByTagNameNS("*", "text")[0] || doc.documentElement;

        const opts = { showPB: showPB.checked };

        const ps = Array.from(body.getElementsByTagNameNS("*", "p"));

        let parts = [];

        if (ps.length) {
            for (const p of ps) {
                const inner = toReadingHtml(p, opts).trim();
                if (!inner && !keepEmptyP.checked) continue;
                parts.push(`<p>${ inner || "&nbsp;" }</p>`);
            }
        } else {
            const inner = toReadingHtml(body, opts).trim();
            parts.push(`<p>${ inner || "&nbsp;" }</p>`);
        }

        out.innerHTML = parts.join("");
        teiStatus.textContent = `Rendered ${ parts.length } paragraph(s).`;
    } catch (e) {
        out.innerHTML = `<div class="empty"><span class="mono">${ escapeHtml(String(e.message || e)) }</span></div>`;
        teiStatus.textContent = "Error";
    }
}

function normalizeBaseUrl(u) {
    return u.replace(/\/+$/, "");
}

function buildUrls() {
    const b = normalizeBaseUrl(baseUrl.value.trim() || "");
    const d = encodeURIComponent(datasetId.value.trim());
    const a = encodeURIComponent(annotationId.value.trim());
    const p = String(pageNum.value).trim();

    const teiUrl = `${ b }/datasets/${ d }/annotations/${ a }/tei/${ p }`;
    const imgUrl = `${ b }/datasets/${ d }/images/${ p }`;

    return { teiUrl, imgUrl };
}

async function loadFromApi() {
    const { teiUrl, imgUrl } = buildUrls();
    reqSummary.textContent = `GET ${ imgUrl } and ${ teiUrl }`;

    imgStatus.textContent = "Loading…";
    teiStatus.textContent = "Loading…";

    // Load image: simplest is setting src, but add cache-busting to ensure page changes reload
    const imgWithBust = `${ imgUrl }${ imgUrl.includes("?") ? "&" : "?" }t=${ Date.now() }`;
    pageImg.src = imgWithBust;

    // Wait for image load/error to update status
    pageImg.onload = () => {
        imgStatus.textContent = "Loaded";
    };
    pageImg.onerror = () => {
        imgStatus.textContent = "Error loading image";
    };

    // Load TEI XML via fetch
    try {
        const res = await fetch(teiUrl, { headers: { "Accept": "application/xml,text/xml,*/*" } });
        if (!res.ok) throw new Error(`TEI HTTP ${ res.status }`);
        const xml = await res.text();
        teiInput.value = xml;
        teiStatus.textContent = "Loaded";
        renderTeiText();
    } catch (e) {
        teiInput.value = "";
        out.innerHTML = `<div class="empty"><span class="mono">${ escapeHtml(String(e.message || e)) }</span></div>`;
        teiStatus.textContent = "Error";
    }
}

loadBtn.addEventListener("click", loadFromApi);

prevBtn.addEventListener("click", () => changePage(-1));
nextBtn.addEventListener("click", () => changePage(1));
function changePage(delta) {
    const current = parseInt(pageInput.value, 10) || 0;
    pageInput.value = Math.max(0, current + delta);
    loadBtn.click();
}

// optional: allow Enter in page field
pageInput.addEventListener("keydown", e => {
    if (e.key === "Enter") loadBtn.click();
});

document.getElementById("renderBtn").addEventListener("click", renderTeiText);
document.getElementById("clearBtn").addEventListener("click", () => {
    teiInput.value = "";
    teiStatus.textContent = "";
    out.innerHTML = '<div class="empty">Click Load.</div>';
});

// re-render when toggles change (if something is already rendered)
showPB.addEventListener("change", () => teiInput.value.trim() && renderTeiText());
keepEmptyP.addEventListener("change", () => teiInput.value.trim() && renderTeiText());

// Convenience: press Enter in any input to load
for (const el of [ baseUrl, datasetId, annotationId, pageNum ]) {
    el.addEventListener("keydown", (e) => {
        if (e.key === "Enter") loadFromApi();
    });
}