#!/usr/bin/env python3
"""Rebuild the curated title-page report outputs and remove uncited diagnostics."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPTS = ROOT / "scripts" / "report"
REPORT = ROOT / "results" / "report"

PIPELINE = [
    "build_core_report_analysis.py",
    "build_bridge_case_table.py",
    "build_commentary_split.py",
    "build_figures_diagrams_deep_dive.py",
    "build_format_material_analysis.py",
    "build_proposition_demonstration_analysis.py",
    "build_author_editor_portfolios.py",
    "build_controlled_portfolio_close_reading.py",
    "build_print_geography.py",
    "analyze_jesuit_non_elements.py",
]


def cited_artifacts() -> set[str]:
    cited: set[str] = set()
    for name in ["report.md", "appendices.md"]:
        text = (REPORT / name).read_text(encoding="utf-8")
        cited.update(re.findall(r"(?:tables|figures)/[^)`\" ]+\.(?:csv|png)", text))
    return cited


def main() -> None:
    for script in PIPELINE:
        print(f"Running {script}")
        subprocess.run([sys.executable, str(SCRIPTS / script)], check=True)

    cited = cited_artifacts()
    for folder in [REPORT / "tables", REPORT / "figures"]:
        for path in folder.iterdir():
            relative = f"{folder.name}/{path.name}"
            if path.is_file() and relative not in cited:
                path.unlink()

    for path in REPORT.glob("REPORT_*_OUTPUTS.md"):
        path.unlink()

    missing = sorted(item for item in cited if not (REPORT / item).exists())
    if missing:
        raise SystemExit("Missing cited outputs after rebuild:\n- " + "\n- ".join(missing))
    print(f"Rebuilt and verified {len(cited)} cited report artifacts.")


if __name__ == "__main__":
    main()
