export function wirePaging({ els, loadFromApi }) {
    const changePage = (delta) => {
        const current = parseInt(els.pageNum.value, 10) || 0;
        els.pageNum.value = String(Math.max(0, current + delta));
        loadFromApi();
    };

    els.prevBtn?.addEventListener("click", () => changePage(-1));
    els.nextBtn?.addEventListener("click", () => changePage(1));

    els.pageNum?.addEventListener("keydown", (e) => {
        if (e.key === "Enter") loadFromApi();
    });

    // used by index clicks
    const jumpToPage = (page) => {
        els.pageNum.value = String(page);
        loadFromApi();
    };

    return { jumpToPage };
}
