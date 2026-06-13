# Phase 1B Subject Families And Calibration

This memo reduces the 20-label classification into visible subject relationships, while flagging calibration issues before interpretation.

## First Calibration Note

The classifier prompt intentionally says not to classify a work as Practical Geometry or Theoretical Mathematics merely because it is a standard Euclidean/Elements text. Therefore, some Euclid title pages have no positive subject classification. This is not automatically an error; it means Euclid/Elements reception must be tracked through TPS features such as `elements_designation`, `references_to_euclid`, and `base_content`, not only through subject labels.

- Representative works with no primary subject: 160
- Title-page editions with no primary subject but geometry/Euclid/Elements-like TPS fields: 100
- Title-page editions whose representative is all-unrelated but has geometry/Euclid/Elements-like TPS fields: 77

These are saved as:

- `no_primary_but_geometry_euclid_like_title_pages.csv`
- `all_unrelated_but_geometry_euclid_like_title_pages.csv`

## Subject Families

| family | primary | secondary | unknown | unrelated |
| --- | --- | --- | --- | --- |
| Arithmetic/Commerce | 144 | 29 | 9 | 661 |
| Geometry/Theory | 424 | 66 | 77 | 276 |
| Visual/Spatial Arts | 110 | 36 | 17 | 680 |
| Instruments/Measurement | 104 | 46 | 40 | 653 |
| Cosmos/Earth | 80 | 30 | 17 | 716 |
| Applied Mechanics/Military | 57 | 29 | 9 | 748 |
| Music | 18 | 6 | 12 | 807 |

## Most Common Primary Subject Pairings

| subject_a | subject_b | count |
| --- | --- | --- |
| Instrument Use | Practical Geometry | 36 |
| Practical Geometry | Surveying | 35 |
| Practical Geometry | Theoretical Mathematics | 27 |
| Arithmetic | Commercial Mathematics | 27 |
| Arithmetic | Theoretical Mathematics | 25 |
| Arithmetic | Practical Geometry | 24 |
| Theoretical Mathematics | Trigonometry | 18 |
| Practical Geometry | Trigonometry | 17 |
| Arithmetic | Trigonometry | 16 |
| Astronomy | Geography | 14 |
| Perspective | Practical Geometry | 13 |
| Astronomy | Theoretical Mathematics | 12 |
| Instrument Construction | Instrument Use | 12 |
| Arithmetic | Astronomy | 12 |
| Construction | Practical Geometry | 11 |
| Astronomy | Instrument Use | 11 |
| Construction | Military Engineering | 11 |
| Military Engineering | Practical Geometry | 11 |
| Instrument Use | Surveying | 10 |
| Arithmetic | Military Engineering | 9 |
| Architecture | Practical Geometry | 9 |
| Arithmetic | Music Theory | 9 |
| Instrument Construction | Practical Geometry | 9 |
| Architecture | Military Engineering | 8 |
| Astronomy | Practical Geometry | 8 |

## Most Common Primary Or Secondary Subject Pairings

| subject_a | subject_b | count |
| --- | --- | --- |
| Instrument Use | Practical Geometry | 72 |
| Practical Geometry | Surveying | 52 |
| Arithmetic | Practical Geometry | 48 |
| Arithmetic | Theoretical Mathematics | 48 |
| Practical Geometry | Theoretical Mathematics | 46 |
| Arithmetic | Commercial Mathematics | 37 |
| Architecture | Practical Geometry | 32 |
| Astronomy | Theoretical Mathematics | 31 |
| Construction | Practical Geometry | 28 |
| Perspective | Practical Geometry | 28 |
| Arithmetic | Trigonometry | 28 |
| Theoretical Mathematics | Trigonometry | 27 |
| Arithmetic | Astronomy | 26 |
| Practical Geometry | Trigonometry | 24 |
| Military Engineering | Practical Geometry | 23 |
| Astronomy | Instrument Use | 22 |
| Construction | Military Engineering | 22 |
| Architecture | Perspective | 21 |
| Instrument Use | Surveying | 21 |
| Astronomy | Geography | 21 |
| Astronomy | Practical Geometry | 20 |
| Instrument Construction | Instrument Use | 19 |
| Architecture | Military Engineering | 18 |
| Arithmetic | Instrument Use | 17 |
| Arithmetic | Surveying | 17 |

## Preliminary Reading

1. The largest named family is Geometry/Theory, but this does not exhaust Euclid: standard Euclidean editions may sit outside positive subject labels.
2. Practical Geometry is a bridge subject. It co-occurs with perspective, instruments, surveying, architecture, military engineering, arithmetic, and theoretical mathematics.
3. The corpus is not one hierarchy from pure to applied mathematics. It looks more like overlapping work-zones: elementary/theoretical, practical/measuring, visual/spatial arts, instruments, commerce, and cosmographic/navigation material.
4. For the conference argument, subject classification should be used together with title-page feature grammar. Subject labels tell us broad topical zones; TPS features tell us how books position themselves.
