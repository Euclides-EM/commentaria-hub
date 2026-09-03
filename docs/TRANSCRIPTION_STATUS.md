| Edition               | Status             |
|-----------------------|--------------------|
| Paris_1667            | Annotation ALTO    |
| Nuremberg_1821        | Annotation ALTO    |
| Basel_1562            | Annotation ALTO    |
| Paris_1598a           | Annotation ALTO    |
| Paris_1615            | Annotation ALTO    |
| Basel_1537            | Raw Markdown       |
| Kiel_and_Leipzig_1669 | Corrected Markdown |
| Paris_1536            | Raw Markdown       |
| Paris_1634            | Corrected Markdown |
| Paris_1639            | Corrected Markdown |
| Rome_1574             | Edition ALTO       |
| Venice_1482           | Edition ALTO       |



## Venice_1482

Two facsimiles are available:

* **`Venice_1482_transkribus`**: The Transkribus facsimile is openly licensed, but the image quality is low. I scraped the transcription from Transkribus and stored locally, but its quality is poor. On the server, I reached the **LineDetect** stage for this facsimile.
* **`Venice_1482`**: The copyright status of this facsimile is currently unknown. I have already OCRed it.

**Next steps:** Wait for Liri to update the copyright status of the facsimiles, then decide which version to continue working with.

---

## Venice_1505 vol1 denoise deskew

Copyright status is currently unknown. At the moment, there is only a **Base annotation**.

**Next steps:** Wait for Liri to update the copyright status of the facsimile. If it is public domain, continue the pipeline. Otherwise, delete it.

---

## Paris_1536 and Basel_1537

Copyright status is currently unknown. I have a full Fable transcription, but it still needs to be manually curated.

**Next steps:** Wait for Liri to update the copyright status of the facsimile. If it can be used, manually curate the existing Fable transcription. Otherwise, try to fit the Fable Markdown files to the new public-domain facsimile and only if it works, manually curate the transcription.

---

## Lyon_1557

The facsimile is open/public domain. The OCR step is currently running.

**Next steps:** Once OCR is complete, run the LLM corrector.

---

## Basel_1562

Copyright status is currently unknown. Kraken OCR is already complete.

**Next steps:** Wait for Liri to update the copyright status of the facsimile. If it can be used, proceed to the LLM corrector. Otherwise, probably restart the pipeline from scratch with a suitable facsimile.

---

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. Transkribus transcriptions are scraped and commited and I also OCRed the facsimile.
* **`Rome_1574`**: Copyright status is currently unknown. OCR is also complete.

**Next steps:** Wait for Liri to finish checking the copyright status, then decide which version to send to the LLM corrector.

---

## Paris_1598a, Paris_1615, Paris_1667 and Nuremberg_1821

OCR is complete, but the copyright status is currently unknown.

**Next steps:** Wait for Liri to update the copyright status, then decide whether to continue with the current facsimile.

---

## Paris_1634, Paris_1639 and Kiel_and_Leipzig_1699

A full Fable transcription exists, manual curation preformed, but the copyright status of the facsimile is currently unknown.

**Next steps:** Wait for Liri to update the copyright status before proceeding.

---
