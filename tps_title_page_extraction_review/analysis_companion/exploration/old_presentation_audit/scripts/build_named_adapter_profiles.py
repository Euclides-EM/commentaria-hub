#!/usr/bin/env python3
"""Enriched same-author/editor circulation profiles inside the Elements corpus."""

from __future__ import annotations

from pathlib import Path
import itertools
import os
import re

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs" / "local_circulation_network"
FIG = OUT / "named_adapter_figures"
FIG.mkdir(parents=True, exist_ok=True)


def clean(value: object) -> str:
    if pd.isna(value):
        return ""
    return str(value).strip()


def split_multi(value: object) -> list[str]:
    text = clean(value)
    if not text:
        return []
    return [p.strip().upper() for p in re.split(r"[,;/|]", text) if p.strip()]


def join_sorted(values: list[str]) -> str:
    return " | ".join(sorted({v for v in values if v}))


def format_label(value: object) -> str:
    text = clean(value)
    if not text or text.lower() == "nan":
        return "unknown"
    try:
        numeric = float(text)
        if numeric.is_integer():
            return str(int(numeric))
    except ValueError:
        pass
    return text


def canonical_author(value: object) -> str:
    text = clean(value)
    text = text.replace(" (?)", "")
    if "Claude-François Milliet Dechales" in text:
        return "Claude-François Milliet Dechales"
    if "Conrad Dasypodius" in text:
        return "Conrad Dasypodius"
    return text


def mode(series: pd.Series) -> str:
    values = [clean(v) for v in series if clean(v)]
    if not values:
        return ""
    counts = pd.Series(values).value_counts()
    return str(counts.index[0])


def language_mode(series: pd.Series) -> str:
    values = list(itertools.chain.from_iterable(split_multi(v) for v in series))
    if not values:
        return ""
    counts = pd.Series(values).value_counts()
    return str(counts.index[0])


def write_profile_tables(df: pd.DataFrame, edges: pd.DataFrame, translations: pd.DataFrame) -> pd.DataFrame:
    df = df[df["author_canonical"].ne("")].copy()
    rows = []
    for author, sub in df.groupby("author_canonical"):
        years = sub["year_num"].dropna()
        langs = list(itertools.chain.from_iterable(split_multi(v) for v in sub["language"]))
        formats = [format_label(v) for v in sub["format"]]
        author_edges = edges[
            edges["from_author_canonical"].eq(author) | edges["to_author_canonical"].eq(author)
        ].copy()
        author_trans = translations[translations["author_canonical"].eq(author)].copy()
        rows.append(
            {
                "author_canonical": author,
                "edition_count": len(sub),
                "earliest_year": int(years.min()) if len(years) else "",
                "latest_year": int(years.max()) if len(years) else "",
                "year_span": int(years.max() - years.min()) if len(years) else "",
                "city_count": sub["city_clean"].dropna().nunique(),
                "cities": join_sorted([clean(v) for v in sub["city_clean"]]),
                "language_count": len(set(langs)),
                "languages": join_sorted(langs),
                "format_count": len(set(formats)),
                "formats": join_sorted(formats),
                "primary_city": mode(sub["city_clean"]),
                "primary_language": language_mode(sub["language"]),
                "primary_format": mode(sub["format"].map(format_label)),
                "known_format_count": int(sub["format"].notna().sum()),
                "reprint_edge_count": len(author_edges),
                "same_city_reprint_edges": int(author_edges["same_city"].sum()) if len(author_edges) else 0,
                "cross_city_reprint_edges": int((~author_edges["same_city"]).sum()) if len(author_edges) else 0,
                "same_language_reprint_edges": int(author_edges["same_language"].sum()) if len(author_edges) else 0,
                "cross_language_reprint_edges": int((~author_edges["same_language"]).sum()) if len(author_edges) else 0,
                "translation_case_count": author_trans["key"].nunique() if len(author_trans) else 0,
                "items": " | ".join(sub.sort_values(["year_num", "key"])["key"].astype(str)),
            }
        )
    profiles = pd.DataFrame(rows).sort_values(
        ["edition_count", "city_count", "language_count", "year_span"], ascending=False
    )
    profiles["is_recurring_3plus"] = profiles["edition_count"].ge(3)
    profiles.to_csv(OUT / "to_1705_named_adapter_profiles_all.csv", index=False)
    profiles.to_csv(OUT / "to_1705_named_adapter_profiles_enriched.csv", index=False)
    profiles[profiles["is_recurring_3plus"]].to_csv(OUT / "to_1705_named_adapter_profiles_recurring_3plus.csv", index=False)

    long = []
    for _, row in df.iterrows():
        for lang in split_multi(row["language"]) or ["UNKNOWN"]:
            long.append(
                {
                    "author_canonical": row["author_canonical"],
                    "key": row["key"],
                    "year": row["year"],
                    "year_num": row["year_num"],
                    "period_25": row["period_25"],
                    "city": row["city_clean"],
                    "language": lang,
                    "format": format_label(row["format"]),
                    "short_title": row["short_title"],
                }
            )
    pd.DataFrame(long).to_csv(OUT / "to_1705_named_adapter_item_long.csv", index=False)
    return profiles


def enrich_edges(edges: pd.DataFrame, df: pd.DataFrame) -> pd.DataFrame:
    mini = df[["key", "author_canonical"]].copy()
    out = edges.merge(mini, left_on="from_key", right_on="key", how="left").rename(
        columns={"author_canonical": "from_author_canonical"}
    )
    out = out.drop(columns=["key"])
    out = out.merge(mini, left_on="to_key", right_on="key", how="left").rename(
        columns={"author_canonical": "to_author_canonical"}
    )
    out = out.drop(columns=["key"])
    out.to_csv(OUT / "to_1705_reprint_cluster_edges_with_authors.csv", index=False)
    return out


def write_heatmaps(profiles: pd.DataFrame, long: pd.DataFrame) -> None:
    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    top_authors = profiles.head(18)["author_canonical"].tolist()
    plot_long = long[long["author_canonical"].isin(top_authors)].copy()

    def heatmap(pivot: pd.DataFrame, title: str, filename: str, figsize=(10, 7), cmap="YlOrBr") -> None:
        if pivot.empty:
            return
        fig, ax = plt.subplots(figsize=figsize)
        im = ax.imshow(pivot.values, aspect="auto", cmap=cmap)
        ax.set_yticks(range(len(pivot.index)), labels=pivot.index)
        ax.set_xticks(range(len(pivot.columns)), labels=pivot.columns, rotation=45, ha="right")
        ax.set_title(title, loc="left", fontsize=13, weight="bold", pad=10)
        for i in range(len(pivot.index)):
            for j in range(len(pivot.columns)):
                value = int(pivot.iat[i, j])
                if value:
                    ax.text(j, i, str(value), ha="center", va="center", fontsize=8, color="#111111")
        fig.colorbar(im, ax=ax, fraction=0.025, pad=0.02, label="Editions")
        fig.tight_layout()
        fig.savefig(FIG / filename, dpi=220)
        plt.close(fig)

    period = plot_long.pivot_table(
        index="author_canonical", columns="period_25", values="key", aggfunc=pd.Series.nunique, fill_value=0
    )
    ordered_period_cols = sorted(period.columns, key=lambda x: int(str(x).split("-")[0]))
    period = period.reindex(index=top_authors, columns=ordered_period_cols).fillna(0).astype(int)
    heatmap(period, "Named Adapter/Editor Editions By Quarter-Century (to 1705)", "author_period_heatmap.png", (12, 7))

    city_counts = plot_long.groupby("city")["key"].nunique().sort_values(ascending=False)
    top_cities = city_counts.head(14).index.tolist()
    city = plot_long.pivot_table(
        index="author_canonical", columns="city", values="key", aggfunc=pd.Series.nunique, fill_value=0
    )
    city = city.reindex(index=top_authors, columns=top_cities).fillna(0).astype(int)
    heatmap(city, "Named Adapter/Editor Editions By City (to 1705)", "author_city_heatmap.png", (11, 7), "YlGnBu")

    fmt_order = ["2", "4", "8", "12", "16", "24", "32", "unknown"]
    fmt = plot_long.pivot_table(
        index="author_canonical", columns="format", values="key", aggfunc=pd.Series.nunique, fill_value=0
    )
    fmt_cols = [c for c in fmt_order if c in fmt.columns] + [c for c in fmt.columns if c not in fmt_order]
    fmt = fmt.reindex(index=top_authors, columns=fmt_cols).fillna(0).astype(int)
    heatmap(fmt, "Named Adapter/Editor Editions By Format (to 1705)", "author_format_heatmap.png", (8, 7), "PuBuGn")

    signal_cols = [
        "edition_count",
        "year_span",
        "city_count",
        "language_count",
        "format_count",
        "reprint_edge_count",
        "cross_city_reprint_edges",
        "translation_case_count",
    ]
    signals = profiles.set_index("author_canonical").reindex(top_authors)[signal_cols].fillna(0).astype(int)
    fig, ax = plt.subplots(figsize=(10, 7))
    im = ax.imshow(signals.values, aspect="auto", cmap="Greens")
    labels = [
        "editions",
        "year span",
        "cities",
        "languages",
        "formats",
        "reprint links",
        "cross-city links",
        "translation cases",
    ]
    ax.set_yticks(range(len(signals.index)), labels=signals.index)
    ax.set_xticks(range(len(signal_cols)), labels=labels, rotation=45, ha="right")
    ax.set_title("Named Adapter/Editor Circulation Signals (to 1705)", loc="left", fontsize=13, weight="bold", pad=10)
    for i in range(len(signals.index)):
        for j in range(len(signal_cols)):
            value = int(signals.iat[i, j])
            if value:
                ax.text(j, i, str(value), ha="center", va="center", fontsize=8, color="#111111")
    fig.colorbar(im, ax=ax, fraction=0.025, pad=0.02, label="Raw count")
    fig.tight_layout()
    fig.savefig(FIG / "author_circulation_signals_heatmap.png", dpi=220)
    plt.close(fig)


def main() -> None:
    df = pd.read_csv(OUT / "all_metadata_elements_items_enriched.csv")
    df = df[df["year_num"].notna() & df["year_num"].le(1705)].copy()
    df["author_canonical"] = df["author_or_editor"].map(canonical_author)

    edges = pd.read_csv(OUT / "to_1705_reprint_cluster_edges.csv")
    edges = enrich_edges(edges, df)

    translations = pd.read_csv(OUT / "to_1705_translation_origin_destination_cases.csv")
    translations = translations.merge(df[["key", "author_canonical"]], on="key", how="left")
    translations.to_csv(OUT / "to_1705_translation_origin_destination_cases_with_authors.csv", index=False)

    profiles = write_profile_tables(df, edges, translations)
    long = pd.read_csv(OUT / "to_1705_named_adapter_item_long.csv")
    write_heatmaps(profiles, long)


if __name__ == "__main__":
    main()
