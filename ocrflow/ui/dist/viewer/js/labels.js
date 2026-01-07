export function datasetLabel(ds) {
    const name = (ds?.name || "").trim();
    const id = (ds?.id || "").trim();
    const edition = (ds?.edition_id || "").trim();
    const flags = [];

    if (ds?.deskewed === true) flags.push("deskewed");
    if (ds?.deskewed === false) flags.push("no deskew");

    const tail = [
        edition ? edition : null,
        flags.length ? flags.join(" · ") : null,
        id ? `(${ id })` : null,
    ]
        .filter(Boolean)
        .join(" ");

    return [ name || id || "Dataset", tail ].filter(Boolean).join(" ");
}

export function annotationLabel(a) {
    const name = (a?.name || "").trim();
    const id = (a?.id || "").trim();

    const seg = a?.segmented === true ? "segmented" : (a?.segmented === false ? "not segmented" : "");
    const pages = (a?.pages || "").trim();
    const desc = (a?.description || "").trim();

    const tail = [
        seg || null,
        pages ? `p. ${ pages }` : null,
        id ? `(${ id })` : null,
    ]
        .filter(Boolean)
        .join(" · ");

    // Keep option text reasonably short
    const head = name || id || "Annotation";
    const shortDesc = desc ? desc.slice(0, 80) + (desc.length > 80 ? "…" : "") : "";

    return [ head, tail ].filter(Boolean).join(" ") + (shortDesc ? ` — ${ shortDesc }` : "");
}