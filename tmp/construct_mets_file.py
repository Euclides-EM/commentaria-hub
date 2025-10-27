#!/usr/bin/env python3
import argparse
import re
from pathlib import Path
import xml.etree.ElementTree as ET

NS_METS = "http://www.loc.gov/METS/"
NS_XLINK = "http://www.w3.org/1999/xlink"

ET.register_namespace("", NS_METS)          # default xmlns
ET.register_namespace("xlink", NS_XLINK)    # xmlns:xlink

def natural_key(s: str):
    return [int(t) if t.isdigit() else t.lower() for t in re.split(r"(\d+)", s)]

def build_mets(file_paths):
    mets = ET.Element(f"{{{NS_METS}}}mets")

    fileSec = ET.SubElement(mets, f"{{{NS_METS}}}fileSec")
    fileGrp = ET.SubElement(fileSec, f"{{{NS_METS}}}fileGrp", {"USE": "export"})

    # map index → id → filename
    id_by_index = []
    for idx, p in enumerate(file_paths, start=1):
        file_el = ET.SubElement(fileGrp, f"{{{NS_METS}}}file", {"ID": f"export{idx}"})
        flocat = ET.SubElement(
            file_el, f"{{{NS_METS}}}FLocat", {f"{{{NS_XLINK}}}href": p.name}
        )
        id_by_index.append((idx, p.name))

    structMap = ET.SubElement(mets, f"{{{NS_METS}}}structMap", {"TYPE": "physical"})
    div_doc = ET.SubElement(structMap, f"{{{NS_METS}}}div", {"TYPE": "document"})

    for idx, _name in id_by_index:
        div_page = ET.SubElement(div_doc, f"{{{NS_METS}}}div", {"TYPE": "page"})
        ET.SubElement(div_page, f"{{{NS_METS}}}fptr", {"FILEID": f"export{idx}"})

    return ET.ElementTree(mets)

def main():
    ap = argparse.ArgumentParser(description="Build METS.xml for a folder of XML pages.")
    ap.add_argument("folder", type=Path, help="Folder containing XML files")
    ap.add_argument("-o", "--output", type=Path, default=None, help="Output path for METS.xml (default: <folder>/METS.xml)")
    ap.add_argument("--pattern", default="*.xml", help='Glob for page files (default: "*.xml")')
    args = ap.parse_args()

    folder = args.folder
    out_path = args.output or (folder / "METS.xml")

    files = [p for p in folder.glob(args.pattern) if p.name != out_path.name]
    if not files:
        raise SystemExit("No XML files found")

    files = sorted(files, key=lambda p: natural_key(p.name))

    tree = build_mets(files)

    # Pretty print (Python 3.9+)
    try:
        ET.indent(tree, space="  ")
    except AttributeError:
        pass

    tree.write(out_path, encoding="utf-8", xml_declaration=False)
    print(f"Wrote {out_path.resolve()} with {len(files)} files.")

if __name__ == "__main__":
    main()
