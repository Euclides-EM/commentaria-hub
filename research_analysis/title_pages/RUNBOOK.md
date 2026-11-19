# Title-Page Analysis Runbook

## Normal Return Path

1. Read `README.md`.
2. Read `results/report/report.md` for the argument.
3. Use `DATA_GUIDE.md` to choose evidence at the correct row grain.
4. Read `results/report/appendices.md` before modifying classifications or denominators.
5. Run `python3 research_analysis/common/scripts/check_base_keys.py` to see whether live metadata has changed around the frozen research population.

## Rebuild The Retained Results

From the repository root:

```sh
python3 -m venv /tmp/elements-analysis-env
/tmp/elements-analysis-env/bin/pip install -r research_analysis/title_pages/scripts/requirements.txt
/tmp/elements-analysis-env/bin/python research_analysis/title_pages/scripts/rebuild_report.py
```

`rebuild_report.py` runs the report analyses in dependency order and removes diagnostic outputs that are not cited by the curated report. It does not rebuild the five analysis-ready matrices from raw extraction rows; see `DATA_GUIDE.md` for that explicit provenance boundary.

## Before Changing A Conclusion

- Record the denominator and corpus scope.
- Link the conclusion to a retained table or add a clearly named final table.
- Keep representative-edition and title-page levels distinct.
- Close-read cited cases against title-page transcription/image evidence.
- Re-run the report-link validator in `rebuild_report.py`.
- Do not retain scratch tables, failed variants, review queues, or chronological phase notes.
