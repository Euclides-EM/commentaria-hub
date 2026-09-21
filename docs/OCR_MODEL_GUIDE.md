# OCR model quick guide for early-modern print

This guide lists **Kraken models only**. The order shown is the recommended trial order.

| Language / era                 | Try in this order                                                                                                                                                                                                                                |
|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Greek, c. 1600–1700            | [Savile 1612](https://github.com/pharos-alexandria/ocr-greek_cursive) → [CLLG](https://zenodo.org/records/21295925) → [Ajax/PoGreTra](https://github.com/AjaxMultiCommentary/OCR-kraken-models)                                                  |
| French, c. 1500–1700           | [`Gallicorpora+_best.mlmodel`](https://zenodo.org/records/7410529) → [CATMuS-Print Large](https://zenodo.org/records/10592716) → [McCATMuS](https://zenodo.org/records/13788177)                                                                 |
| English, c. 1500–1700          | [CATMuS-Print Large](https://zenodo.org/records/10592716) → [McCATMuS](https://zenodo.org/records/13788177) → [`en_best`](https://zenodo.org/records/2577813)                                                                                    |
| Italian, c. 1550–1650          | [CATMuS-Print Large](https://zenodo.org/records/10592716) → [McCATMuS](https://zenodo.org/records/13788177)                                                                                                                                      |
| Spanish, c. 1600–1700          | [CATMuS-Print Large](https://zenodo.org/records/10592716) → [McCATMuS](https://zenodo.org/records/13788177)                                                                                                                                      |
| Dutch, c. 1600–1700            | [CATMuS-Print Large](https://zenodo.org/records/10592716) → [German Print](https://zenodo.org/records/10519596) → [McCATMuS](https://zenodo.org/records/13788177)                                                                                |
| German, c. 1500–1700 (Antiqua) | [German Print](https://zenodo.org/records/10519596) → [CATMuS-Print Large](https://zenodo.org/records/10592716) → [McCATMuS](https://zenodo.org/records/13788177)                                                                                |
| German, c. 1500–1700 (Fraktur) | [German Print](https://zenodo.org/records/10519596) → [CATMuS-Print Large](https://zenodo.org/records/10592716) → [Swedish Fraktur](https://huggingface.co/MagnusKolsjo/svensk-fraktur-kraken) → [McCATMuS](https://zenodo.org/records/13788177) |
| Latin, especially 16th c.      | [`gallicorpora_ajuste`](https://zenodo.org/records/19218113) → [CATMuS-Print Large](https://zenodo.org/records/10592716)                                                                                                                         |

## Quick rules

- CATMuS is the default Latin-script baseline and preserves `ſ`, historical `u/v` and `i/j`, and abbreviations.
- `en_best` is only a control for diplomatic OCR: its alphabet lacks long `ſ`.
