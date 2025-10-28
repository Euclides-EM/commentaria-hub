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
* https://github.com/PSL-Chartes-HTR-Students/HN2021-Boccace -- 2 texts, do not look like my kind of texts
* https://github.com/DesenrollandoElCordel/FoNDUE-Spanish-chapbooks-Dataset -- it's spanish, do not look like my kind of texts
* https://zenodo.org/records/11046062
* https://github.com/ksefil/NuBIS-OCR

To this I'm adding the data from [Gallicorpora](https://gallicorpora.github.io/data/htr/)

# Download Datasets

```bash
python scripts/download_gh_folder.py "https://github.com/OCR-D/gt_structure_text/tree/main/data/alberti_pictura_1540" -o datasets/train/alberti_pictura_1540
```

# Training Dataset

```bash
yaltai convert alto-to-yolo /Users/mia/dev/personal/elements-dh/datasets/train/alberti_pictura_1540/GT-PAGE/*.xml /Users/mia/dev/personal/elements-dh/datasets/train/yolo-ground-truth --shuffle .1 --segmonto region
yolo detect train data=/Users/mia/dev/personal/elements-dh/datasets/train/yolo-ground-truth/config.yml model=/Users/mia/dev/personal/elements-dh/models/downloaded/nano-yolo-ladas.pt  epochs=100 imgsz=640

# Start training from a pretrained *.pt model
yolo detect train data=coco8.yaml model=yolo11n.pt epochs=100 imgsz=640

# Build a new model from YAML, transfer pretrained weights to it and start training
yolo detect train data=coco8.yaml model=yolo11n.yaml pretrained=yolo11n.pt epochs=100 imgsz=640
```

# OCR models

| Model Name          | Link                                |
|---------------------|-------------------------------------|
| Gallicorpor.mlmodel |                                     |
| CATMuS-Print Large  | https://zenodo.org/records/10592716 |

# Segmonto Classes
Note: based on https://universe.roboflow.com/colaftextes/segmonto

AdvertisementZone
DigitizationArtefactZone
DropCapitalZone
FigureZone
FigureZone-FigDesc
FigureZone-Head
FormZone
GraphicZone
GraphicZone-Decoration
GraphicZone-FigDesc
GraphicZone-Head
GraphicZone-Maths
GraphicZone-Part
GraphicZone-TextualContent
MainZone-Continued
MainZone-Date
MainZone-Entry
MainZone-Entry-Continued
MainZone-Head
MainZone-Lg
MainZone-Lg-Continued
MainZone-List-Continued
MainZone-ListItem
MainZone-Maths
MainZone-Other
MainZone-P
MainZone-P-Continued
MainZone-Signature
MainZone-Sp
MainZone-Sp-Continued
MarginTextZone-ContinuedNotes
MarginTextZone-ManuscriptAddendum
MarginTextZone-Notes
MarginTextZone-Notes-Continued
MusicZone
NumberingZone
PageTitleZone
PageTitleZone-Index
QuireMarksZone
RunningTitleZone
StampZone
StampZone-Sticker
TableZone
TableZone-Continued
TableZone-Head
TitlePageZone
TitlePageZone-Index