import json
import re
import sys
from pathlib import Path

ROMAN = {'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100}
PROP_RE = re.compile(r'PRO(?:P|PO|POS)?\.*[,.:]*\s*([IVXLCHivxlcE]+)\b')
NUM_RE = re.compile(r'(?:THEOREME|THEOR|THEO|IHE0|PRROB|PROBL|PROB)[\s.,:]*([0-9ISOlD]+)(?![A-Za-z])')
CHAP_RE = re.compile(r'CHAP\.?\s*([IVXLCH]+)\b|\b([IVXLCH]+)\.\s*CHAP\b')
CORRECTION_RE = re.compile(r'\S+?\{printer-error-correction:([^}]*)\}')
CURATED_RE = re.compile(r'^\[Curated heading (?=[^\]]*level=(\d+))[^\]:]*:\s*(.*)\]\s*$')
PLAIN_RE = re.compile(r'^\s*(?:THEO|THEOR|PROB|PROBL)\b.*\bPRO')


def roman(s):
    s = s.upper().replace('H', 'I')
    if not s or any(c not in ROMAN for c in s):
        return None
    total = 0
    for i, c in enumerate(s):
        nxt = ROMAN[s[i + 1]] if i + 1 < len(s) else 0
        total += -ROMAN[c] if ROMAN[c] < nxt else ROMAN[c]
    return total


def arabic(s):
    s = s.rstrip('.').replace('I', '1').replace('l', '1').replace('S', '8').replace('O', '0')
    return int(s) if s.isdigit() else None


def from_index(path):
    data = json.loads(Path(path).read_text(encoding='utf-8'))
    rows = []

    def walk(nodes):
        for n in nodes or []:
            level = int(n['category'].replace('header', '') or 0)
            rows.append((n['location']['page'], level, n['content'], True))
            walk(n['children'])

    walk(data['nodes'])
    return rows


def from_dir(path):
    rows = []
    for f in sorted(Path(path).glob('page-*/original.md')):
        page = str(int(f.parent.name.split('-')[1]))
        for line in f.read_text(encoding='utf-8').splitlines():
            line = CORRECTION_RE.sub(r'\1', line)
            m = re.match(r'^(#+)\s+(.*)', line)
            c = CURATED_RE.match(line)
            if m:
                rows.append((page, len(m.group(1)), m.group(2).strip(), True))
            elif c:
                rows.append((page, int(c.group(1)), c.group(2).strip(), True))
            elif PLAIN_RE.match(line) or re.match(r'^\s*(?:THEO\w*|PROB\w*)[\s.]*[0-9ISOl]*\.?\s*$', line):
                rows.append((page, 2, line.strip(), False))
    return rows


def sections(rows):
    current, items = None, []
    for page, level, text, is_header in rows:
        if level == 1 and is_header:
            if text != current:
                if items:
                    yield current, items
                current, items = text, []
            continue
        items.append((page, text, is_header))
    if items:
        yield current, items


def check(name, items):
    out = []
    expected_prop, counters, expected_chap = 0, {'T': 0, 'P': 0}, 0
    for page, text, is_header in items:
        if not is_header:
            out.append(f'  p{page}: NOT A HEADER: {text[:70]!r}')
        c = CHAP_RE.search(text)
        if c and not re.search(r'PROP', text):
            n = roman(c.group(1) or c.group(2))
            if n != expected_chap + 1:
                out.append(f'  p{page}: {text[:60]!r} -> CHAP expected {expected_chap + 1} got {n}')
            expected_chap = n if n and abs(n - expected_chap - 1) <= 2 else expected_chap + 1
            continue
        if not re.search(r'PRO|THE|IHE', text) or re.search(r'Fin d', text):
            continue
        stripped = re.sub(r'P+R+OB\w*', '', text)
        pm = PROP_RE.search(stripped)
        nm = NUM_RE.search(text)
        kind = 'T' if re.search(r'TH|IHE', text) else 'P'
        prop = roman(pm.group(1)) if pm else None
        num = arabic(nm.group(1)) if nm else None
        notes = []
        if prop and prop <= 2 and expected_prop > 5:
            out.append(f'  p{page}: {text[:60]!r} -> numbering restarts: missing section header (level 1) before this?')
            expected_prop, counters = prop - 1, {'T': 0, 'P': 0}
            counters[kind] = (num or 1) - 1
        if prop is None:
            notes.append(f'PROP unreadable (expected {expected_prop + 1})')
        elif prop != expected_prop + 1:
            notes.append(f'PROP expected {expected_prop + 1} got {prop}')
        if num is None:
            notes.append(f'{kind} number unreadable (expected {counters[kind] + 1})')
        elif num != counters[kind] + 1:
            notes.append(f'{kind} expected {counters[kind] + 1} got {num}')
        if notes:
            out.append(f'  p{page}: {text[:60]!r} -> ' + '; '.join(notes))
        if prop and abs(prop - expected_prop - 1) <= 2:
            expected_prop = prop
        else:
            expected_prop += 1
        if num and abs(num - counters[kind] - 1) <= 2:
            counters[kind] = num
        else:
            counters[kind] += 1
    if out:
        print(f'===== {name}')
        print('\n'.join(out))


def main():
    src = Path(sys.argv[1])
    rows = from_index(src) if src.suffix == '.json' else from_dir(src)
    for name, items in sections(rows):
        check(name, items)


if __name__ == '__main__':
    main()
