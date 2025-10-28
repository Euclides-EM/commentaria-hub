# Test Corpus
I'm starting with four Latin editions from the 16th century. 
I downloaded their scans from Google Books.

| Year | Editor/Printer   | Link                                                                                               | PDF                                                      | Notes    |
|------|------------------|----------------------------------------------------------------------------------------------------|----------------------------------------------------------|----------|
| 1557 | Peletier         | [Link](https://www.google.com/books/edition/Iacobi_Peletarii_Cenomani_In_Euclidis_El/bsz37Ilvq3QC) | [Peletier_1557.pdf](pdfs/full/Peletier_1557.pdf)         |          |
| 1566 | de Foix-Candalle | [Link](https://www.google.com/books/edition/Euclidis_Elementa_geometrica_libris_15_a/9nWsxqL3RZgC) | [de_Foix_1566_vol1.pdf](pdfs/full/de_Foix_1566_vol1.pdf) | Volume 1 |
| 1572 | Commandino       | [Link](https://www.google.com/books/edition/Euclidis_elementorum_libri_XV/jEoRPuxe_pcC)            | [Commandino_1572.pdf](pdfs/full/Commandino_1572.pdf)     |          |
| 1574 | Clavius          | [Link](https://www.google.com/books/edition/Euclidis_Elementorum_libri_XV/9E2xEit4CKkC)            | [Clavius_1574.pdf](pdfs/full/Clavius_1574.pdf)           | Volume 1 |

For each edition, I identified manually where is the second book of the Elements, to make the OCR and transcription easier. 

| Year | Editor/Printer   | PDF | Second Book Pages |
|------|------------------|-----|-------------------|
| 1557 | Peletier         |     | 71-82             |
| 1566 | de Foix-Candalle |     | 54-62             |
| 1572 | Commandino       |     | 85-100            |
| 1574 | Clavius          |     | 238-275           |

Then, to process it, I run the following scripts to create the JPGs of the relevant pages (second book only):
```bash
python scripts/pdf_to_jpg.py datasets/test/pdfs/full/Peletier_1557.pdf datasets/test/jpgs/book2/Peletier_1557 71 82
python scripts/pdf_to_jpg.py datasets/test/pdfs/full/de_Foix_1566_vol1.pdf datasets/test/jpgs/book2/de_Foix_1566 54 62
python scripts/pdf_to_jpg.py datasets/test/pdfs/full/Commandino_1572.pdf datasets/test/jpgs/book2/Commandino_1572 85 100
python scripts/pdf_to_jpg.py datasets/test/pdfs/full/Clavius_1574.pdf datasets/test/jpgs/book2/Clavius_1574 238 275
```

# Test Dataset

```bash
yaltai kraken --device cpu -I "/Users/mia/dev/personal/elements-dh/datasets/test/jpgs/book2/Clavius_1574/*.jpg" --alto --suffix ".xml" segment --yolo /Users/mia/.asdf/runs/detect/train3/weights/last.pt ocr -m /Users/mia/dev/personal/elements-dh/tmp/Gallicorpor.mlmodel
```