# render_alto_overlay.py
# usage: python render_alto_overlay.py alto.xml page.jpg out.png
import sys
import xml.etree.ElementTree as ET
from PIL import Image, ImageDraw

def parse_float(a, k, default=0.0):
    v = a.get(k)
    return float(v) if v is not None else default

alto_path, img_path, out_path = sys.argv[1], sys.argv[2], sys.argv[3]

# load image
img = Image.open(img_path).convert("RGB")
W_img, H_img = img.size

# parse ALTO
tree = ET.parse(alto_path)
root = tree.getroot()

# ALTO namespaces vary. Try common default
ns = {"a": root.tag.split('}')[0].strip('{')} if '}' in root.tag else {}

# get page size in ALTO
page = root.find(".//a:Layout/a:Page", ns) or root.find(".//Layout/Page", ns)
W_alto = int(page.get("WIDTH")) if page is not None and page.get("WIDTH") else W_img
H_alto = int(page.get("HEIGHT")) if page is not None and page.get("HEIGHT") else H_img

sx = W_img / float(W_alto) if W_alto else 1.0
sy = H_img / float(H_alto) if H_alto else 1.0

draw = ImageDraw.Draw(img)

# helpers to draw rectangles
def rect(a, w="WIDTH", h="HEIGHT", x="HPOS", y="VPOS"):
    x0 = parse_float(a, x) * sx
    y0 = parse_float(a, y) * sy
    x1 = x0 + parse_float(a, w) * sx
    y1 = y0 + parse_float(a, h) * sy
    return [x0, y0, x1, y1]

# draw TextBlock, TextLine, String with different widths
for blk in root.findall(".//a:TextBlock", ns) + root.findall(".//TextBlock", ns):
    draw.rectangle(rect(blk.attrib), outline=(255, 0, 0), width=2)

for line in root.findall(".//a:TextLine", ns) + root.findall(".//TextLine", ns):
    draw.rectangle(rect(line.attrib), outline=(255, 165, 0), width=1)

for word in root.findall(".//a:String", ns) + root.findall(".//String", ns):
    draw.rectangle(rect(word.attrib), outline=(0, 255, 0), width=1)

# optionally draw spaces too
for sp in root.findall(".//a:SP", ns) + root.findall(".//SP", ns):
    draw.rectangle(rect(sp.attrib), outline=(0, 0, 255), width=1)

img.save(out_path)
print(f"saved overlay → {out_path}")

