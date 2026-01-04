import { escapeHtml, parseXml } from "./utils.js";

function isElement(node, localName) {
    return node && node.nodeType === Node.ELEMENT_NODE && node.localName === localName;
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

        html += toReadingHtml(child, opts);
    }

    return html;
}

export function renderTeiText({ teiInput, out, teiStatus }) {
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

        const opts = { showPB: true };
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
