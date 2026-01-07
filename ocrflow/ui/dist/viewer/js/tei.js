import { escapeHtml, parseXml } from "./utils.js";

function isElement(node, localName) {
    return node && node.nodeType === Node.ELEMENT_NODE && node.localName === localName;
}

function findFirstCertaintyDegree(segEl) {
    // Look for a descendant <certainty degree="..."> inside the seg
    const certs = segEl.getElementsByTagNameNS("*", "certainty");
    if (!certs || !certs.length) return null;

    const degreeStr = certs[0].getAttribute("degree");
    const degree = degreeStr == null ? NaN : parseFloat(degreeStr);
    return Number.isFinite(degree) ? degree : null;
}

function textContentExcludingCertainty(node) {
    let out = "";
    for (const child of node.childNodes) {
        if (child.nodeType === Node.TEXT_NODE) {
            out += child.nodeValue || "";
            continue;
        }
        if (child.nodeType !== Node.ELEMENT_NODE) continue;

        // Skip certainty markup from the visible text
        if (isElement(child, "certainty")) continue;

        // Recurse
        out += textContentExcludingCertainty(child);
    }
    return out;
}

function maskText(s, maskChar) {
    const m = maskChar && String(maskChar).length ? String(maskChar)[0] : "@";
    let out = "";
    for (const ch of s) {
        out += /\s/.test(ch) ? ch : m;
    }
    return out;
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
                const label = (n || facs) ? `Page break ${escapeHtml(n || facs)}` : "Page break";
                html += `<div class="pb">${label}</div>`;
            }
            continue;
        }

        // Special handling: <seg> ... <certainty degree="..."/> ... </seg>
        if (isElement(child, "seg")) {
            const degree = findFirstCertaintyDegree(child);
            const rawText = textContentExcludingCertainty(child);

            const minCert = Number.isFinite(opts.minCert) ? opts.minCert : 0;
            const masked =
                degree != null && degree < minCert
                    ? maskText(rawText, opts.maskChar)
                    : rawText;

            html += escapeHtml(masked);
            continue;
        }

        html += toReadingHtml(child, opts);
    }

    return html;
}

export function renderTeiText({ teiInput, out, teiStatus, minCert = 0, maskChar = "@" }) {
    const xml = teiInput.value.trim();
    if (!xml) {
        out.innerHTML = '<div class="empty">No TEI loaded.</div>';
        teiStatus.textContent = "";
        return;
    }

    try {
        const doc = parseXml(xml);

        const body =
            doc.getElementsByTagNameNS("*", "body")[0] ||
            doc.getElementsByTagNameNS("*", "text")[0] ||
            doc.documentElement;

        const opts = { showPB: true, minCert, maskChar };
        const ps = Array.from(body.getElementsByTagNameNS("*", "p"));

        const parts = [];

        if (ps.length) {
            for (const p of ps) {
                const inner = toReadingHtml(p, opts).trim();
                if (!inner) continue;
                parts.push(`<p>${inner || "&nbsp;"}</p>`);
            }
        } else {
            const inner = toReadingHtml(body, opts).trim();
            parts.push(`<p>${inner || "&nbsp;"}</p>`);
        }

        out.innerHTML = parts.join("");
        teiStatus.textContent = `Rendered ${parts.length} paragraph(s).`;
    } catch (e) {
        out.innerHTML = `<div class="empty"><span class="mono">${escapeHtml(String(e.message || e))}</span></div>`;
        teiStatus.textContent = "Error";
    }
}
