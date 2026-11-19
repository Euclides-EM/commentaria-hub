import re
import sys
from collections import Counter, defaultdict
from difflib import SequenceMatcher
from pathlib import Path

from find_gaps import CORRECTION_RE, CURATED_RE, from_index

CHECKS = [
    (re.compile(r'\S {2,}\S'), 'repeated spaces'),
    (re.compile(r'\s[.,:;]'), 'space before punctuation'),
    (re.compile(r'([.,:;])\s*\1|[.,][,:;]|[,:;]\.'), 'repeated or mixed punctuation'),
    (re.compile(r'[A-Za-z]\.?(?=\d)'), 'no space between label and number'),
    (re.compile(r'\d[IlSOD]|[IlSOD]\d|\d[a-z]'), 'letter mixed into number'),
    (re.compile(r'[a-zà-ÿ][A-Z]'), 'case change inside word'),
    (re.compile(r"[^\w\s.,:;'’()&\-\[\]̀-ͯ]"), 'unexpected character'),
]
NUMERAL_RE = re.compile(r'^(?:[IVXLCDMH]+|\d+)\.?$')


def from_dir(path):
    rows = []
    for f in sorted(Path(path).glob('page-*/original.md')):
        page = str(int(f.parent.name.split('-')[1]))
        for line in f.read_text(encoding='utf-8').splitlines():
            line = CORRECTION_RE.sub(r'\1', line)
            m = re.match(r'^(#+)(\s*)(.*?)(\s*)$', line)
            c = CURATED_RE.match(line)
            if m and m.group(3):
                rows.append((page, len(m.group(1)), m.group(3), m.group(2) != ' ' or bool(m.group(4))))
            elif c:
                rows.append((page, int(c.group(1)), c.group(2).strip(), False))
    return rows


def label_head(text):
    return re.split(r'[.,:;\d]', text, maxsplit=1)[0].strip()


def label_key(text):
    return re.sub(r'[^A-Za-zÀ-ÿſ]', '', label_head(text)).upper()


def main():
    args = sys.argv[1:]
    outline = '--outline' in args
    src = Path([a for a in args if a != '--outline'][0])
    rows = [(p, l, t, False) for p, l, t, _ in from_index(src)] if src.suffix == '.json' else from_dir(src)
    keys = [label_key(t) for _, _, t, _ in rows]
    heads = [label_head(t) for _, _, t, _ in rows]
    key_counts = Counter(keys)
    head_forms = defaultdict(Counter)
    prefix_levels = defaultdict(Counter)
    for (_, level, _, _), key, head in zip(rows, keys, heads):
        head_forms[key][head] += 1
        prefix_levels[key[:4]][level] += 1
    prev_level = 0
    for (page, level, text, bad_spacing), key, head in zip(rows, keys, heads):
        notes = [msg for rx, msg in CHECKS if rx.search(text)]
        if bad_spacing:
            notes.append('spacing around header text')
        letters = [c for c in text if c.isalpha() and c != 'ſ']
        upper = sum(c.isupper() for c in letters)
        if letters and upper != len(letters) and upper >= 0.7 * len(letters):
            notes.append('lowercase letters in uppercase header')
        if NUMERAL_RE.match(text):
            notes.append('header is only a numeral')
        if level > prev_level + 1:
            notes.append(f'level skips from {prev_level}')
        prev_level = level
        if key:
            levels = prefix_levels[key[:4]]
            total = sum(levels.values())
            usual, n = levels.most_common(1)[0]
            if total >= 3 and usual != level and levels[level] * 4 <= total:
                notes.append(f'{key[:4]}* headers are usually level {usual} ({n}/{total})')
        if key_counts[key] >= 3:
            form = head_forms[key].most_common(1)[0][0]
            if head != form:
                notes.append(f'label differs from usual {form!r}')
        elif len(key) >= 5:
            similar = [k for k, c in key_counts.items()
                       if k != key and len(k) >= 5 and c >= 3 and SequenceMatcher(None, key, k).ratio() >= 0.8]
            if similar:
                notes.append('possible misspelling of ' + ', '.join(repr(head_forms[k].most_common(1)[0][0]) for k in similar))
        if notes or outline:
            print(f'p{page} L{level} {text!r}' + (' -> ' + '; '.join(notes) if notes else ''))


if __name__ == '__main__':
    main()
