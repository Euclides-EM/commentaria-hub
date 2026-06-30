#!/usr/bin/env python3
"""Experiment 4: local circulation network for metadata-defined Elements editions.

The corpus is complete for metadata-defined Elements editions currently recorded
in metadata, including reprints. It is not a complete bibliography of all works
by each author/editor.
"""

from __future__ import annotations

from pathlib import Path
import itertools
import math
import os
import re

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
META = ROOT.parents[1] / "ocrflow" / "store" / "items_metadata"
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs" / "local_circulation_network"
OUT.mkdir(parents=True, exist_ok=True)


def first_year(value: object) -> float:
    text = str(value)
    for token in text.replace("–", "/").replace("-", "/").split("/"):
        token = token.strip()
        if token.isdigit() and len(token) == 4:
            return float(token)
    return float("nan")


def clean_text(value: object) -> str:
    if pd.isna(value):
        return ""
    return str(value).strip()


def clean_language(value: object) -> str:
    text = clean_text(value)
    if not text:
        return ""
    parts = [p.strip().upper() for p in re.split(r"[,;/|]", text) if p.strip()]
    return " | ".join(parts)


def split_multi(value: object) -> list[str]:
    text = clean_text(value)
    if not text:
        return []
    return [p.strip().upper() for p in re.split(r"[,;/|]", text) if p.strip()]


def join_unique(values: pd.Series) -> str:
    items = sorted({clean_text(v) for v in values if clean_text(v)})
    return " | ".join(items)


def join_unique_lang(values: pd.Series) -> str:
    langs: set[str] = set()
    for value in values:
        langs.update(split_multi(value))
    return " | ".join(sorted(langs))


def scoped(df: pd.DataFrame, scope: str) -> pd.DataFrame:
    if scope == "to_1705":
        return df[df["year_num"].notna() & df["year_num"].le(1705)].copy()
    return df.copy()


def write_counts(df: pd.DataFrame, scope_name: str) -> None:
    city_counts = (
        df.groupby("city_clean", dropna=False)
        .agg(
            count=("key", "count"),
            earliest_year=("year_num", "min"),
            latest_year=("year_num", "max"),
            languages=("language_clean", join_unique_lang),
            authors_editors=("author_or_editor", join_unique),
        )
        .reset_index()
        .rename(columns={"city_clean": "city"})
        .sort_values(["count", "city"], ascending=[False, True])
    )
    city_counts.to_csv(OUT / f"{scope_name}_city_counts.csv", index=False)

    lang_rows = []
    for _, row in df.iterrows():
        langs = split_multi(row["language"])
        if not langs:
            langs = ["missing/unknown"]
        for lang in langs:
            lang_rows.append(
                {
                    "key": row["key"],
                    "language": lang,
                    "year_num": row["year_num"],
                    "city": row["city_clean"],
                    "author_or_editor": row["author_or_editor"],
                }
            )
    lang_df = pd.DataFrame(lang_rows)
    language_counts = (
        lang_df.groupby("language")
        .agg(
            count=("key", "nunique"),
            earliest_year=("year_num", "min"),
            latest_year=("year_num", "max"),
            cities=("city", join_unique),
            authors_editors=("author_or_editor", join_unique),
        )
        .reset_index()
        .sort_values(["count", "language"], ascending=[False, True])
    )
    language_counts.to_csv(OUT / f"{scope_name}_language_counts.csv", index=False)

    city_lang = (
        lang_df.groupby(["city", "language"])
        .agg(count=("key", "nunique"), earliest_year=("year_num", "min"), latest_year=("year_num", "max"))
        .reset_index()
        .sort_values(["count", "city", "language"], ascending=[False, True, True])
    )
    city_lang.to_csv(OUT / f"{scope_name}_city_language_counts.csv", index=False)

    period_city = (
        df.groupby(["period_25", "city_clean"])
        .size()
        .reset_index(name="count")
        .rename(columns={"city_clean": "city"})
        .sort_values(["period_25", "count"], ascending=[True, False])
    )
    period_city.to_csv(OUT / f"{scope_name}_city_counts_by_25_year_period.csv", index=False)


def build_reprint_network(df: pd.DataFrame, cluster_items: pd.DataFrame, clusters: pd.DataFrame, scope_name: str) -> None:
    reprint_clusters = clusters[clusters["type"].eq("reprint")][["key", "type"]].rename(columns={"key": "cluster_key"})
    ci = cluster_items.merge(reprint_clusters, on="cluster_key", how="inner")
    ci = ci[ci["item_key"].isin(df["key"])]
    cluster_df = ci.merge(df, left_on="item_key", right_on="key", how="left", suffixes=("", "_item"))

    summary_rows = []
    edge_rows = []
    complete_cluster_items = []
    for cluster_key, sub in cluster_df.groupby("cluster_key"):
        # Avoid repeated cluster-item rows in metadata.
        sub = sub.drop_duplicates("item_key").copy()
        if len(sub) < 2:
            continue
        sub = sub.sort_values(["year_num", "city_clean", "item_key"], na_position="last")
        years = sub["year_num"].dropna()
        cities = [clean_text(x) for x in sub["city_clean"] if clean_text(x)]
        languages = sorted(set(itertools.chain.from_iterable(split_multi(x) for x in sub["language"])))
        authors = join_unique(sub["author_or_editor"])
        summary_rows.append(
            {
                "cluster_key": cluster_key,
                "item_count": len(sub),
                "earliest_year": int(years.min()) if len(years) else "",
                "latest_year": int(years.max()) if len(years) else "",
                "year_span": int(years.max() - years.min()) if len(years) else "",
                "city_count": len(set(cities)),
                "cities": " | ".join(sorted(set(cities))),
                "language_count": len(languages),
                "languages": " | ".join(languages),
                "authors_editors": authors,
                "items": " | ".join(sub["item_key"].astype(str)),
            }
        )
        for _, row in sub.iterrows():
            complete_cluster_items.append(
                {
                    "cluster_key": cluster_key,
                    "item_key": row["item_key"],
                    "year": row["year"],
                    "year_num": row["year_num"],
                    "city": row["city_clean"],
                    "language": row["language_clean"],
                    "author_or_editor": row["author_or_editor"],
                    "short_title": row["short_title"],
                }
            )
        ordered = sub[sub["year_num"].notna()].copy()
        for idx in range(len(ordered) - 1):
            a = ordered.iloc[idx]
            b = ordered.iloc[idx + 1]
            edge_rows.append(
                {
                    "cluster_key": cluster_key,
                    "from_key": a["item_key"],
                    "to_key": b["item_key"],
                    "from_year": int(a["year_num"]),
                    "to_year": int(b["year_num"]),
                    "year_delta": int(b["year_num"] - a["year_num"]),
                    "from_city": a["city_clean"],
                    "to_city": b["city_clean"],
                    "from_language": a["language_clean"],
                    "to_language": b["language_clean"],
                    "same_city": a["city_clean"] == b["city_clean"],
                    "same_language": a["language_clean"] == b["language_clean"],
                }
            )

    summary = pd.DataFrame(summary_rows).sort_values(["city_count", "item_count", "year_span"], ascending=False)
    edges = pd.DataFrame(edge_rows)
    items = pd.DataFrame(complete_cluster_items)
    summary.to_csv(OUT / f"{scope_name}_reprint_cluster_summary.csv", index=False)
    edges.to_csv(OUT / f"{scope_name}_reprint_cluster_edges.csv", index=False)
    items.to_csv(OUT / f"{scope_name}_reprint_cluster_items.csv", index=False)

    if len(edges):
        city_edges = (
            edges.groupby(["from_city", "to_city"])
            .agg(
                edge_count=("cluster_key", "count"),
                cluster_count=("cluster_key", "nunique"),
                earliest_from_year=("from_year", "min"),
                latest_to_year=("to_year", "max"),
            )
            .reset_index()
            .sort_values(["edge_count", "cluster_count"], ascending=False)
        )
        city_edges.to_csv(OUT / f"{scope_name}_reprint_city_edges.csv", index=False)
        lang_edges = (
            edges.groupby(["from_language", "to_language"])
            .agg(
                edge_count=("cluster_key", "count"),
                cluster_count=("cluster_key", "nunique"),
                earliest_from_year=("from_year", "min"),
                latest_to_year=("to_year", "max"),
            )
            .reset_index()
            .sort_values(["edge_count", "cluster_count"], ascending=False)
        )
        lang_edges.to_csv(OUT / f"{scope_name}_reprint_language_edges.csv", index=False)


def build_translation_edges(df: pd.DataFrame, title_page: pd.DataFrame, scope_name: str) -> None:
    tp = title_page[["key", "origin_language", "destination_language"]].copy()
    merged = df.merge(tp, on="key", how="left")
    rows = []
    for _, row in merged.iterrows():
        origins = split_multi(row.get("origin_language"))
        dests = split_multi(row.get("destination_language"))
        if not origins or not dests:
            continue
        for origin in origins:
            for dest in dests:
                rows.append(
                    {
                        "key": row["key"],
                        "year": row["year"],
                        "year_num": row["year_num"],
                        "city": row["city_clean"],
                        "language": row["language_clean"],
                        "origin_language": origin,
                        "destination_language": dest,
                        "author_or_editor": row["author_or_editor"],
                        "short_title": row["short_title"],
                    }
                )
    out = pd.DataFrame(rows)
    out.to_csv(OUT / f"{scope_name}_translation_origin_destination_cases.csv", index=False)
    if len(out):
        edges = (
            out.groupby(["origin_language", "destination_language"])
            .agg(
                count=("key", "nunique"),
                earliest_year=("year_num", "min"),
                latest_year=("year_num", "max"),
                cities=("city", join_unique),
                examples=("key", lambda s: " | ".join(s.astype(str).head(12))),
            )
            .reset_index()
            .sort_values(["count", "origin_language", "destination_language"], ascending=[False, True, True])
        )
        edges.to_csv(OUT / f"{scope_name}_translation_origin_destination_edges.csv", index=False)


def build_author_within_elements(df: pd.DataFrame, scope_name: str) -> None:
    rows = []
    for author, sub in df[df["author_or_editor"].notna()].groupby("author_or_editor"):
        author = clean_text(author)
        if not author:
            continue
        years = sub["year_num"].dropna()
        rows.append(
            {
                "author_or_editor": author,
                "elements_item_count": len(sub),
                "earliest_year": int(years.min()) if len(years) else "",
                "latest_year": int(years.max()) if len(years) else "",
                "city_count": sub["city_clean"].dropna().nunique(),
                "cities": join_unique(sub["city_clean"]),
                "language_count": len(set(itertools.chain.from_iterable(split_multi(x) for x in sub["language"]))),
                "languages": join_unique_lang(sub["language"]),
                "items": " | ".join(sub["key"].astype(str)),
            }
        )
    pd.DataFrame(rows).sort_values(["elements_item_count", "city_count"], ascending=False).to_csv(
        OUT / f"{scope_name}_author_editor_within_elements_summary.csv", index=False
    )


def build_scope_summary(df: pd.DataFrame, scope_name: str) -> dict[str, object]:
    years = df["year_num"].dropna()
    langs = sorted(set(itertools.chain.from_iterable(split_multi(x) for x in df["language"])))
    return {
        "scope": scope_name,
        "item_count": len(df),
        "known_year_count": int(df["year_num"].notna().sum()),
        "earliest_year": int(years.min()) if len(years) else "",
        "latest_year": int(years.max()) if len(years) else "",
        "city_count": df["city_clean"].dropna().nunique(),
        "language_count": len(langs),
    }


def maybe_write_figures(scope_name: str) -> None:
    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    city_counts_path = OUT / f"{scope_name}_city_counts.csv"
    city_edges_path = OUT / f"{scope_name}_reprint_city_edges.csv"
    if not city_counts_path.exists():
        return

    city_counts = pd.read_csv(city_counts_path).head(15)
    fig, ax = plt.subplots(figsize=(8.8, 4.8))
    ax.barh(city_counts["city"][::-1], city_counts["count"][::-1], color="#d95f02")
    ax.set_title(f"Top Elements Cities ({scope_name.replace('_', ' ')})", loc="left", fontsize=15, weight="bold")
    ax.set_xlabel("Editions")
    ax.spines[["top", "right", "left"]].set_visible(False)
    ax.grid(axis="x", alpha=0.2)
    fig.tight_layout()
    fig.savefig(OUT / f"{scope_name}_top_city_counts.png", dpi=220)
    plt.close(fig)

    if not city_edges_path.exists():
        return
    edges = pd.read_csv(city_edges_path)
    cross = edges[edges["from_city"].ne(edges["to_city"])].copy()
    if len(cross) == 0:
        return
    cities = sorted(set(cross["from_city"]) | set(cross["to_city"]))
    matrix = pd.DataFrame(0, index=cities, columns=cities)
    for _, row in cross.iterrows():
        matrix.loc[row["from_city"], row["to_city"]] += int(row["edge_count"])

    fig, ax = plt.subplots(figsize=(7.2, 5.8))
    im = ax.imshow(matrix.values, cmap="YlGnBu")
    ax.set_xticks(range(len(cities)), labels=cities, rotation=45, ha="right")
    ax.set_yticks(range(len(cities)), labels=cities)
    ax.set_title(f"Cross-City Reprint Links ({scope_name.replace('_', ' ')})", loc="left", fontsize=15, weight="bold")
    for i, from_city in enumerate(cities):
        for j, to_city in enumerate(cities):
            value = matrix.loc[from_city, to_city]
            if value:
                ax.text(j, i, str(value), ha="center", va="center", color="#111111", fontsize=9, weight="bold")
    fig.colorbar(im, ax=ax, fraction=0.046, pad=0.04, label="Links")
    fig.tight_layout()
    fig.savefig(OUT / f"{scope_name}_cross_city_reprint_heatmap.png", dpi=220)
    plt.close(fig)


def main() -> None:
    items = pd.read_csv(META / "items_print.csv")
    elements = pd.read_csv(META / "metadata_elements_print.csv")
    cluster_items = pd.read_csv(META / "cluster_items.csv")
    clusters = pd.read_csv(META / "clusters.csv")
    title_page = pd.read_csv(META / "title_page.csv")

    df = elements.rename(columns={"key": "element_key"}).merge(
        items, left_on="element_key", right_on="key", how="left", indicator=True
    )
    qa = df[df["_merge"].ne("both")][["element_key", "_merge"]]
    qa.to_csv(OUT / "metadata_elements_missing_from_items_print.csv", index=False)
    df = df[df["_merge"].eq("both")].drop(columns=["_merge", "element_key"])
    df["year_num"] = df["year"].map(first_year)
    df["city_clean"] = df["city"].map(clean_text)
    df["language_clean"] = df["language"].map(clean_language)
    df["period_25"] = (df["year_num"] // 25 * 25).map(lambda x: f"{int(x)}-{int(x + 24)}" if not math.isnan(x) else "missing")
    df.to_csv(OUT / "all_metadata_elements_items_enriched.csv", index=False)

    summaries = []
    for scope_name in ["all_metadata_elements", "to_1705"]:
        sub = scoped(df, scope_name)
        summaries.append(build_scope_summary(sub, scope_name))
        write_counts(sub, scope_name)
        build_reprint_network(sub, cluster_items, clusters, scope_name)
        build_translation_edges(sub, title_page, scope_name)
        build_author_within_elements(sub, scope_name)
        maybe_write_figures(scope_name)
    pd.DataFrame(summaries).to_csv(OUT / "scope_summary.csv", index=False)


if __name__ == "__main__":
    main()
