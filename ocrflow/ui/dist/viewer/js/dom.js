export function dom() {
    const $ = (id) => document.getElementById(id);

    const els = {
        // inputs
        baseUrl: $("baseUrl"),
        datasetId: $("datasetId"),
        annotationId: $("annotationId"),
        pageNum: $("pageNum"),

        // slider
        minCert: $("minCert"),
        minCertVal: $("minCertVal"),

        // buttons
        loadBtn: $("loadBtn"),
        prevBtn: $("prevPage"),
        nextBtn: $("nextPage"),
        renderBtn: $("renderBtn"),
        toggleTeiSourceBtn: $("toggleTeiSourceBtn"),
        toggleAnnotationBtn: $("toggleAnnotationBtn"),

        // TEI source container (new)
        teiSourceTools: $("teiSourceTools"),

        // Annotation info
        annotationDetails: $("annotationDetails"),

        // image + tei
        pageImg: $("pageImg"),
        teiInput: $("teiInput"),
        out: $("out"),

        // status
        imgStatus: $("imgStatus"),
        teiStatus: $("teiStatus"),
        reqSummary: $("reqSummary"),

        // index UI
        indexStatus: $("indexStatus"),
        indexTree: $("indexTree"),
        indexFilter: $("indexFilter"),
        reloadIndexBtn: $("reloadIndexBtn"),

        // sidebar toggle
        toggleIndexBtn: $("toggleIndexBtn"),
    };

    return { $, els };
}
