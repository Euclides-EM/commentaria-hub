# Research Analysis

This is the canonical home for the project's three active analysis areas. Each area starts with its own `README.md` and uses the same compact vocabulary:

- `data/`: retained evidence or analysis-ready source tables;
- `scripts/`: code that regenerates retained results;
- `results/`: conclusions, final tables, and final figures actually worth reading.

## Areas

- [`edition_classification/`](edition_classification/README.md): final reviewed subject classifications.
- [`title_pages/`](title_pages/README.md): title-page ecology report and supporting analyses.
- [`dotted_lines/`](dotted_lines/README.md): dotted-line conclusions and evidence tables.

Intermediate runs, review queues, abandoned questions, unused figures, superseded versions, logs, and presentation/document copies are intentionally excluded.

Frozen study memberships live with the data they govern: `edition_classification/data/`, `dotted_lines/data/`, and `title_pages/data/base_keys/`. Run `python3 research_analysis/common/scripts/check_base_keys.py` after adding or modifying metadata.
