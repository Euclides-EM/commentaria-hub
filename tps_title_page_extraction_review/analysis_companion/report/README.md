# Report Workspace

Polished report files, printable packet, report-specific scripts, generated tables, and generated figures.

## Read / Print

- `REPORT_PRINT_PACKET.html`: best file for browser reading and printing.
- `REPORT_PRINT_PACKET.md`: combined Markdown packet.

## Edit

- `REPORT_DRAFT.md`: main report body.
- `REPORT_APPENDICES.md`: technical appendices.

After editing the draft or appendices, rebuild the packet:

```bash
python3 tps_title_page_extraction_review/analysis_companion/report/scripts/build_print_packet.py
```

## Supporting Material

| Path | Purpose |
|---|---|
| `figures/` | PNG figures referenced by the report. |
| `tables/` | Report-ready generated CSV tables. |
| `scripts/` | Scripts that generate report tables, figures, and print packet. |
| `supporting_notes/` | Report-construction notes, output summaries, and shortlists used while drafting. |
