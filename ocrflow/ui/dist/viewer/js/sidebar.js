const KEY = "indexCollapsed";

function setIndexCollapsed(collapsed) {
    document.documentElement.classList.toggle("index-collapsed", collapsed);
    try { localStorage.setItem(KEY, collapsed ? "1" : "0"); } catch {}
}

function isIndexCollapsed() {
    try { return localStorage.getItem(KEY) === "1"; } catch { return false; }
}

export function wireSidebarToggle(toggleBtn) {
    setIndexCollapsed(isIndexCollapsed());

    if (!toggleBtn) return;

    toggleBtn.addEventListener("click", () => {
        const collapsed = document.documentElement.classList.contains("index-collapsed");
        setIndexCollapsed(!collapsed);
    });
}
