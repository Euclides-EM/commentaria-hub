# alto_to_html.py
# usage: python alto_to_html.py alto.xml page.jpg out.html
import sys, base64, xml.etree.ElementTree as ET
from pathlib import Path

alto_path, img_path, out_html = sys.argv[1], sys.argv[2], sys.argv[3]

# base64 embed image for a single-file viewer
b64 = base64.b64encode(Path(img_path).read_bytes()).decode("ascii")

root = ET.parse(alto_path).getroot()
ns = {"a": root.tag.split('}')[0].strip('{')} if '}' in root.tag else {}

page = root.find(".//a:Layout/a:Page", ns) or root.find(".//Layout/Page", ns)
W_alto = int(page.get("WIDTH"))
H_alto = int(page.get("HEIGHT"))

words = []
for s in root.findall(".//a:String", ns) + root.findall(".//String", ns):
    a = s.attrib
    x = float(a["HPOS"])
    y = float(a["VPOS"])
    w = float(a["WIDTH"])
    h = float(a["HEIGHT"])
    content = a.get("CONTENT", "")
    words.append((x, y, w, h, content))

html = f"""<!doctype html>
<meta charset="utf-8">
<title>ALTO overlay</title>
<style>
  .page {{
    position: relative;
    width: 100%;
    max-width: 95vw;
    margin: 1rem auto;
    background: #eee;
    outline: 1px solid #ccc;
  }}
  .canvas {{
    display: block;
    width: 100%;
    height: auto;
  }}
  .word {{
    position: absolute;
    border: 1px solid rgba(0,0,0,.2);
    white-space: nowrap;
    font: 12px/1.1 system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif;
    transform-origin: 0 0;
    background: rgba(255,255,255,.4);
  }}
</style>
<div class="page" style="aspect-ratio: {W_alto}/{H_alto}">
  <img class="canvas" src="data:image/jpeg;base64,{b64}">
  <!-- Words will be sized via percentages of ALTO page size -->
  {"".join(
    f'<span class="word" style="left:{x/W_alto*100:.4f}%;top:{y/H_alto*100:.4f}%'
    f';width:{w/W_alto*100:.4f}%;height:{h/H_alto*100:.4f}%">{content}</span>'
    for x,y,w,h,content in words
  )}
</div>
"""
Path(out_html).write_text(html, encoding="utf-8")
print(f"wrote {out_html}")

