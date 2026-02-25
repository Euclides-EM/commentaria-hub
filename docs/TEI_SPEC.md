# TEI Export Pipeline Design

## OCR → NER → Entity Profiles → TEI Encoding

## Overview

This document discusses the conversion between OCRed texts into structured TEI XML enriched with:

* Named entity recognition results
* Entity metadata and profiles
* Relations between entities
* Parallel translations aligned line-by-line
* Optional page images and layout geometry

The pipeline supports two main source types:

### ALTO OCR XML

Contains:

* Layout geometry (blocks, lines, bounding boxes)
* OCR text per line

Used when precise positional data is available.

### Plain text transcription (+ translations)

Used when:

* OCR layout is unavailable
* Transcription already normalized
* Translation files exist line-aligned with transcription

Both flows produce a consistent TEI output model.

## Core Data Model

### Unified entity input

All entity information is expressed through one struct:

```go
type EntityItem struct {
    Ref string

    // Inline mention (optional)
    PageID  string
    BlockID string
    LineID  string
    Start   int
    End     int
    Element string
    Ana     string

    // Profiles / relations
    Type               string
    Value              string
    ObjectRef          string
    Cert               float64
    EvidenceMentionIDs []string
}
```

This merges:

* entity mentions
* entity metadata
* entity relations

Benefits:

* One ingestion pipeline
* Easy deduplication
* Flexible enrichment
* Async entity processing possible

---

## TEI Structure Used

### High-level TEI document

```xml
<TEI xmlns="http://www.tei-c.org/ns/1.0">
  <teiHeader>
    <fileDesc>…</fileDesc>
    <profileDesc>…</profileDesc>
    <standOff>…</standOff>
  </teiHeader>

  <facsimile>…</facsimile>

  <text>
    <body>
      <pb/>
      <div type="transcription">…</div>
      <div type="translation">…</div>
    </body>
  </text>
</TEI>
```

## TEI Header Encoding

### Entity profiles

Profiles are stored generically:

```xml
<profileDesc>
  <textClass>
    <keywords scheme="entity-profiles">
      <term type="latinized_name" corresp="#ent_ibn_rushd">
        Averroes
      </term>
    </keywords>
  </textClass>
</profileDesc>
```

Properties:

* Generic key/value storage
* Entity referenced via `@corresp`
* No strict ontology required

This avoids premature schema locking.

### Relations between entities

Stored in `<standOff>`:

```xml
<standOff>
  <listRelation>
    <relation
      xml:id="fact_1"
      name="educatedAt"
      active="#ent_person"
      passive="#ent_oxford"
      source="#m_12"
      cert="0.91"/>
  </listRelation>
</standOff>
```

Fallback for unknown relation types:

```xml
<interpGrp type="fact_fallback">
  <interp xml:id="fact_x"
          type="custom_relation"
          corresp="#ent_a">
    #ent_b
  </interp>
</interpGrp>
```

## Text Body Encoding

### Transcription layer

Each page:

```xml
<div type="transcription" n="page1">
  <ab xml:id="b1_page1">
    <seg xml:id="ln_page1_0001">
      <persName xml:id="m_1"
                ref="#ent_ibn_rushd"
                ana="#feat_person">
        Ibn Rushd
      </persName>
      was a philosopher.
    </seg>
  </ab>
</div>
```

Key points:

* Each line gets a stable `xml:id`
* Entities embedded inline
* Mention IDs allow relation linking

### 5.2 Translation alignment

Translations remain parallel:

```xml
<div type="translation" xml:lang="en" n="page1">
  <ab xml:id="tr_en_page1">
    <seg corresp="#ln_page1_0001">
      Ibn Rushd was a philosopher.
    </seg>
  </ab>
</div>
```

Benefits:

* Explicit alignment
* Multiple languages supported
* No duplication of entity markup

## 6. Facsimile Layer

Used when ALTO data exists:

```xml
<facsimile>
  <surface xml:id="page_1" facs="page1.png">
    <zone xml:id="block_1" type="block" …/>
    <zone xml:id="line_1" type="line" …/>
  </surface>
</facsimile>
```

Purpose:

* Preserve layout geometry
* Enable diagram extraction
* Support future visualization

## 7. Builder Architecture

Two primary builders:

### ALTO Builder

```go
BuildTEIFromALTO(alto, entities, imageURL)
```

Uses:

* ALTO block and line IDs
* Bounding boxes
* OCR geometry

Recommended when OCR layout exists.

### Lines Builder

```go
BuildTEIFromLines(linesInput, entities, imageURLs)
```

Used when:

* Only transcription available
* Translations line-aligned
* No geometry needed

Important convention:

```
BlockID = "b1"
LineID  = l0001, l0002…
```

## Mention ID Strategy

Mention IDs are assigned in document order:

```
m_1, m_2, m_3…
```

Mention IDs are emitted as:

```xml
<persName xml:id="m_12">
```

---

## Deduplication Strategy

### Profiles

Deduplicated by:

```
(entity ref, type, value)
```

### Relations

Sorted and optionally deduplicated by:

```
(subject, object, feature)
```

Ensures stable output and avoids duplicates.

---

## 10. Offset Handling

Offsets must match the exact reconstructed line text.

For ALTO:

```go
strings.Join(strings, " ")
```

NER must compute offsets against this same normalization.

UTF-8 safety:

* Byte offsets used
* Adjusted to valid rune boundaries

## Service Layer Flow

Main entry point:

```go
func GetTEI(datasetID, annotationID, pageKey, features)
```

Steps:

1. Load annotation
2. Try ALTO page retrieval
3. If ALTO exists:

    * build TEI from ALTO
4. Otherwise:

    * load TXT transcription
    * load translations
    * resolve image URL
    * build TEI from lines
5. Convert to XML

## 12. Feature Mapping to TEI Elements

Example mapping:

```go
func (f *Feature) TEIElement() string {
    switch strings.ToLower(f.Name) {
    case "person":
        return "persName"
    case "org":
        return "orgName"
    case "place":
        return "placeName"
    default:
        return "name"
    }
}
```

Separates:

* semantic detection
* TEI encoding choice

## Known Edge Cases

### BlockID mismatches

Must default missing BlockID to `"b1"` in lines builder.

### Offset drift

Occurs if OCR normalization changes.

### Overlapping entities

Currently filtered by simple non-nesting rule.

## Recommended Future Enhancements

### Authority lists

Optional:

```
<listPerson>
<listPlace>
```

for persistent entity registries.

### RDF export

Possible from standOff relations.

### Diagram annotation

Using facsimile zones.

### Confidence propagation

Integrate `cert` deeper into entity markup.

## Example Minimal Final TEI

```xml
<TEI xmlns="http://www.tei-c.org/ns/1.0">
  <teiHeader>
    <profileDesc>
      <textClass>
        <keywords scheme="entity-profiles">
          <term type="latinized_name" corresp="#ent_ibn_rushd">
            Averroes
          </term>
        </keywords>
      </textClass>
    </profileDesc>
  </teiHeader>

  <text>
    <body>
      <div type="transcription">
        <ab>
          <seg xml:id="ln_page1_0001">
            <persName xml:id="m_1" ref="#ent_ibn_rushd">
              Ibn Rushd
            </persName>
            was a philosopher.
          </seg>
        </ab>
      </div>
    </body>
  </text>
</TEI>
```
