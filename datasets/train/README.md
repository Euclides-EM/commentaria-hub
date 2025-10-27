# Ground Truth

I used the following [catalog](https://htr-united.github.io/catalog.html) and filtered by:
Language: Latin
Script: All scripts
Script type: Typed only
Dates / Not before: -250
Dates / Not after: 1800

In total, there are six relevant repos:
* https://github.com/OCR-D/gt_structure_text
* https://github.com/HTR-United/cremma-16-17-print
* https://github.com/PSL-Chartes-HTR-Students/HN2021-Boccace
* https://github.com/DesenrollandoElCordel/FoNDUE-Spanish-chapbooks-Dataset
* https://zenodo.org/records/11046062
* https://github.com/ksefil/NuBIS-OCR

To this I'm adding the data from [Gallicorpora](https://gallicorpora.github.io/data/htr/)

# Download Datasets

```bash
python scripts/download_gh_folder.py "https://github.com/OCR-D/gt_structure_text/tree/main/data/alberti_pictura_1540" -o datasets/train/alberti_pictura_1540
```

# Training Dataset

```bash
yaltai convert alto-to-yolo /Users/mia/dev/personal/elements-dh/datasets/train/alberti_pictura_1540/GT-PAGE/*.xml my-dataset --shuffle .1 --segmonto region
yolo task=detect mode=train model=yolov8n.pt data=my-dataset/config.yml epochs=100 plots=True batch=8 imgsz=960
```