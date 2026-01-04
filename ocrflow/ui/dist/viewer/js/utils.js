export function escapeHtml(str) {
    return (str || "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

export function parseXml(xmlString) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(xmlString, "application/xml");
    const pe = doc.getElementsByTagName("parsererror")[0];
    if (pe) throw new Error("XML parse error (often unescaped & or mismatched tags).");
    return doc;
}

export function normalizeBaseUrl(u) {
    return (u || "").toString().trim().replace(/\/+$/, "");
}

export function normalizeText(s) {
    return (s || "").toString().trim();
}

export function onEnter(el, fn) {
    if (!el) return;
    el.addEventListener("keydown", (e) => {
        if (e.key === "Enter") fn();
    });
}
