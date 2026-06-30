#!/usr/bin/env python3
"""Build a compact corpus histogram for presentation use."""

from __future__ import annotations

import os
from pathlib import Path

os.environ.setdefault("MPLCONFIGDIR", "/tmp/elements_dh_matplotlib")

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "derived_data"
REPORT = ROOT / "report"
FIGURES = REPORT / "figures"
TABLES = REPORT / "tables"
FIGURES.mkdir(parents=True, exist_ok=True)
TABLES.mkdir(parents=True, exist_ok=True)


def first_year(value: object) -> float:
    text = str(value)
    for token in text.replace("–", "/").replace("-", "/").split("/"):
        token = token.strip()
        if token.isdigit() and len(token) == 4:
            return float(token)
    return np.nan


def count_clean(series: pd.Series, normalize_upper: bool = False) -> int:
    cleaned = series.dropna().astype(str).str.strip()
    cleaned = cleaned[cleaned.ne("")]
    if normalize_upper:
        cleaned = cleaned.str.upper()
    return cleaned.nunique()


def main() -> None:
    df = pd.read_csv(DERIVED / "metadata_elements_corpus_ecology_matrix.csv")
    df["year_num"] = df["year"].map(first_year)
    df = df[df["year_num"].notna()].copy()
    df = df[df["year_num"].le(1705)].copy()
    df["is_elements"] = df["is_metadata_elements_representative"].astype(str).str.lower().eq("true")

    start = int(df["year_num"].min())
    end = int(df["year_num"].max())
    bin_start = (start // 25) * 25
    bin_end = ((end // 25) + 1) * 25
    bins = np.arange(bin_start, bin_end + 25, 25)
    labels = [f"{int(a)}-{int(min(b - 1, end))}" for a, b in zip(bins[:-1], bins[1:])]
    df["bin"] = pd.cut(df["year_num"], bins=bins, right=False, labels=labels)

    counts = (
        df.groupby(["bin", "is_elements"], observed=False)
        .size()
        .unstack(fill_value=0)
        .rename(columns={False: "Other mathematical books", True: "Editions of the Elements"})
    )
    for col in ["Other mathematical books", "Editions of the Elements"]:
        if col not in counts:
            counts[col] = 0
    counts = counts[["Other mathematical books", "Editions of the Elements"]]
    counts.to_csv(TABLES / "report_corpus_histogram_25_year_bins.csv")

    fig, ax = plt.subplots(figsize=(10.8, 4.8))
    x = np.arange(len(counts))
    ecology = counts["Other mathematical books"].to_numpy()
    elements = counts["Editions of the Elements"].to_numpy()

    other_bars = ax.bar(x, ecology, color="#b9ad98", width=0.74, label="Other mathematical books")
    elements_bars = ax.bar(x, elements, bottom=ecology, color="#f26400", width=0.74, label="Editions of the Elements")

    ax.set_title("Corpus Distribution by Quarter-Century", fontsize=17, weight="bold", pad=14, color="#1f1f1f")
    ax.set_ylabel("")
    ax.set_xlabel("")
    ax.set_xticks(x)
    ax.set_xticklabels(counts.index, rotation=45, ha="right", fontsize=8.5)
    ax.tick_params(axis="y", labelsize=9)
    ax.grid(axis="y", color="#e7e2d8", linewidth=0.8)
    ax.set_axisbelow(True)
    ax.spines["top"].set_visible(False)
    ax.spines["right"].set_visible(False)
    ax.spines["left"].set_color("#c9c0b2")
    ax.spines["bottom"].set_color("#c9c0b2")
    ax.legend(
        [elements_bars, other_bars],
        ["Editions of the Elements", "Other mathematical books"],
        frameon=False,
        loc="upper left",
        fontsize=10,
    )
    fig.patch.set_facecolor("#ffffff")
    ax.set_facecolor("#ffffff")
    plt.tight_layout()
    plt.savefig(FIGURES / "corpus_histogram_25_year_bins.png", dpi=240)
    plt.close()

    summary = pd.DataFrame(
        [
            {
                "corpus": "all representative title-page records",
                "count": len(df),
                "earliest_year": int(df["year_num"].min()),
                "latest_year": int(df["year_num"].max()),
                "city_count": count_clean(df["city"]),
                "language_count": count_clean(df["language_first"], normalize_upper=True),
            },
            {
                "corpus": "metadata-defined Elements",
                "count": int(df["is_elements"].sum()),
                "earliest_year": int(df.loc[df["is_elements"], "year_num"].min()),
                "latest_year": int(df.loc[df["is_elements"], "year_num"].max()),
                "city_count": count_clean(df.loc[df["is_elements"], "city"]),
                "language_count": count_clean(df.loc[df["is_elements"], "language_first"], normalize_upper=True),
            },
            {
                "corpus": "surrounding mathematical ecology",
                "count": int((~df["is_elements"]).sum()),
                "earliest_year": int(df.loc[~df["is_elements"], "year_num"].min()),
                "latest_year": int(df.loc[~df["is_elements"], "year_num"].max()),
                "city_count": count_clean(df.loc[~df["is_elements"], "city"]),
                "language_count": count_clean(df.loc[~df["is_elements"], "language_first"], normalize_upper=True),
            },
        ]
    )
    summary.to_csv(TABLES / "report_corpus_scope_summary.csv", index=False)


if __name__ == "__main__":
    main()
