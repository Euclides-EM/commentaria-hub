#!/usr/bin/env python3
"""Build a single printable report packet from the main draft and appendices."""

from __future__ import annotations

from pathlib import Path

import markdown


REPORT = Path(__file__).resolve().parents[2] / "results" / "report"
DRAFT = REPORT / "report.md"
APPENDICES = REPORT / "appendices.md"
OUT_MD = REPORT / "print_packet.md"
OUT_HTML = REPORT / "print_packet.html"


CSS = """
body {
  font-family: Georgia, "Times New Roman", serif;
  color: #111;
  line-height: 1.48;
  max-width: 980px;
  margin: 2.2rem auto;
  padding: 0 2rem 4rem;
}
h1, h2, h3 {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  line-height: 1.2;
}
h1 {
  font-size: 2rem;
  margin-bottom: 1.4rem;
}
h2 {
  border-top: 1px solid #ddd;
  padding-top: 1.1rem;
  margin-top: 2.4rem;
}
h3 {
  margin-top: 1.8rem;
}
table {
  border-collapse: collapse;
  width: 100%;
  margin: 1rem 0 1.4rem;
  font-size: 0.88rem;
}
th, td {
  border: 1px solid #ccc;
  padding: 0.35rem 0.45rem;
  vertical-align: top;
}
th {
  background: #f2f2f2;
}
img {
  display: block;
  max-width: 100%;
  margin: 1rem auto 1.6rem;
  page-break-inside: avoid;
}
code {
  font-family: "SFMono-Regular", Consolas, monospace;
  font-size: 0.9em;
}
blockquote {
  border-left: 3px solid #bbb;
  margin-left: 0;
  padding-left: 1rem;
  color: #333;
}
@media print {
  body {
    max-width: none;
    margin: 0;
    padding: 0;
    font-size: 10.5pt;
  }
  h1, h2 {
    page-break-after: avoid;
  }
  table, img {
    page-break-inside: avoid;
  }
}
"""


def main() -> None:
    draft = DRAFT.read_text()
    appendices = APPENDICES.read_text()
    combined = draft.rstrip() + "\n\n---\n\n" + appendices.lstrip()
    OUT_MD.write_text(combined)

    body = markdown.markdown(
        combined,
        extensions=["tables", "fenced_code", "toc"],
        output_format="html5",
    )
    html = f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>The Place of Euclid's Elements in Early Modern Mathematical Print</title>
<style>{CSS}</style>
</head>
<body>
{body}
</body>
</html>
"""
    OUT_HTML.write_text(html)


if __name__ == "__main__":
    main()
