# Title-Page Base Keys

These sorted, one-key-per-line files freeze the title-page populations used by the current report:

- `title_page_keys.txt`: 909 title pages, including reprints;
- `title_page_representative_keys.txt`: 843 reprint-deduplicated representatives;
- `title_page_elements_representative_keys.txt`: 286 metadata-defined Elements representatives;
- `title_page_elements_print_geography_keys.txt`: 320 Elements editions used for print-geography denominators.

New live metadata keys do not enter the current study automatically. Revise these files only when deliberately creating a new corpus revision and regenerating affected results. `metadata_checksums.sha256` makes edits to existing metadata files visible.

Check all locally owned key sets with:

```sh
python3 research_analysis/common/scripts/check_base_keys.py
```
